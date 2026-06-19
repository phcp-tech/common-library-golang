// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package component provides HTTP server lifecycle integration for bootstrap.
package component

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/httpserver/lambda"
	"github.com/phcp-tech/common-library-golang/shutdown"
)

// Compile-time check: httpComponent implements bootstrap.IComponent.
var _ bootstrap.IComponent = (*httpComponent)(nil)

type httpComponent struct {
	handler func() http.Handler
	runner  httpserver.IRunner
}

// loadFromEnv creates an HTTP server IRunner from the koanf env singleton.
// It reads app.runmode and http.server.port, returning the appropriate IRunner.
func loadFromEnv() httpserver.IRunner {
	if strings.EqualFold(env.Env().String("app.runmode"), "aws_lambda") {
		slog.Info("Http server is running under AWS-LAMBDA")
		return lambda.NewHttpServer()
	}
	port := env.Env().String("http.server.port")
	slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s", port))
	return httpserver.NewHttpServer(httpserver.Config{Port: port})
}

func (h *httpComponent) Name() string { return "httpserver" }

func (h *httpComponent) Init() error {
	h.runner = loadFromEnv()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Panic in http server goroutine",
					"error", r,
					"stack", string(debug.Stack()))
				shutdown.Trigger()
			}
		}()
		if err := h.runner.Start(h.handler()); err != nil {
			slog.Error("Http server stopped with error", "error", err)
			shutdown.Trigger()
		}
	}()
	return nil
}

func (h *httpComponent) Close() {
	if h.runner == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpserver.DefaultShutdownTimeout)
	defer cancel()
	if err := h.runner.Shutdown(ctx); err != nil {
		slog.Error("Http server shutdown failed", "error", err)
	}
	slog.Info("Http server has been shutdown")
}

// Component wraps the HTTP server as a bootstrap.IComponent.
//
// handler is a lazy provider called during Init() to obtain the http.Handler
// (typically *gin.Engine). Passing a function rather than the handler directly
// defers evaluation until Init() time, at which point the Gin component has
// already been initialised.
//
// Typical usage with a bridge variable in main:
//
//	var router *gin.Engine
//	Add(gin.Component(origins, func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(component.Component(func() http.Handler { return router }))
func Component(handler func() http.Handler) bootstrap.IComponent {
	return &httpComponent{handler: handler}
}
