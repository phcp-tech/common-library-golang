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

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/shutdown"
)

// Compile-time check: httpComponent implements bootstrap.IComponent.
var _ bootstrap.IComponent = (*httpComponent)(nil)

type httpComponent struct {
	handler func() http.Handler
	factory func() httpserver.IRunner
	runner  httpserver.IRunner
}

// loadFromEnv creates an HTTP server IRunner from the koanf env singleton.
// base supplies every Config field except Port, which always comes from the
// http.server.port env key regardless of what base.Port is set to - matching
// every other consumer's convention of configuring the listen port via env,
// not code.
func loadFromEnv(base httpserver.Config) func() httpserver.IRunner {
	return func() httpserver.IRunner {
		base.Port = env.Env().String("http.server.port")
		slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s", base.Port))
		return httpserver.NewHttpServer(base)
	}
}

func (h *httpComponent) Name() string { return "httpserver" }

func (h *httpComponent) Init() error {
	h.runner = h.factory()
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

// ComponentWithRunner wraps the HTTP server as a bootstrap.IComponent using a
// custom runner factory. Use this when the caller needs to control how the
// IRunner is created — for example, to inject a Lambda runner.
//
// Prefer [Component] for plain HTTP/HTTPS deployments.
func ComponentWithRunner(handler func() http.Handler, factory func() httpserver.IRunner) bootstrap.IComponent {
	return &httpComponent{handler: handler, factory: factory}
}

// Component wraps the HTTP server as a bootstrap.IComponent.
//
// handler is a lazy provider called during Init() to obtain the http.Handler
// (typically *gin.Engine). Passing a function rather than the handler directly
// defers evaluation until Init() time, at which point the Gin component has
// already been initialised.
//
// cfg is optional; pass at most one - extra values beyond the first are
// ignored. It supplies every httpserver.Config field except Port, which is
// always overwritten from the http.server.port env key regardless of what
// cfg.Port is set to. Omitting cfg reproduces this function's previous
// behavior (a Config with only Port set).
//
// Typical usage with a bridge variable in main:
//
//	var router *gin.Engine
//	Add(ginComp.Component(func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(httpComp.Component(func() http.Handler { return router }))
//
// A service with a long-lived SSE/streaming endpoint can raise (or disable)
// the write timeout without giving up bootstrap's lifecycle management (see
// [ComponentWithRunner] for the previous, more verbose way to do this):
//
//	Add(httpComp.Component(func() http.Handler { return router },
//	    httpserver.Config{WriteTimeout: httpserver.NoWriteTimeout}))
func Component(handler func() http.Handler, cfg ...httpserver.Config) bootstrap.IComponent {
	var base httpserver.Config
	if len(cfg) > 0 {
		base = cfg[0]
	}
	return ComponentWithRunner(handler, loadFromEnv(base))
}
