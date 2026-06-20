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

// Package componentwithlambda provides HTTP server lifecycle integration for
// bootstrap, with AWS Lambda support.
//
// Use this package instead of httpserver/component when the service may be
// deployed on AWS Lambda. It reads app.runmode from env during Init() and
// selects the appropriate runner:
//   - "aws_lambda" → httpserver/lambda.NewHttpServer()
//   - anything else → httpserver.NewHttpServer() (plain HTTP/HTTPS)
//
// Importing this package pulls in the AWS Lambda SDK. Services that never run
// on Lambda should use httpserver/component to keep their binary free of
// unnecessary dependencies.
package componentwithlambda

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	httpComp "github.com/phcp-tech/common-library-golang/httpserver/component"
	"github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// loadFromEnv selects the appropriate IRunner based on app.runmode.
func loadFromEnv() httpserver.IRunner {
	if strings.EqualFold(env.Env().String("app.runmode"), "aws_lambda") {
		slog.Info("Http server is running under AWS-LAMBDA")
		return lambda.NewHttpServer()
	}

	port := env.Env().String("http.server.port")
	slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s", port))
	return httpserver.NewHttpServer(httpserver.Config{Port: port})
}

// Component wraps the HTTP server as a bootstrap.IComponent with Lambda support.
//
// handler is a lazy provider called during Init() to obtain the http.Handler
// (typically *gin.Engine). The runner is selected at Init() time based on the
// app.runmode env key.
//
// Typical usage with a bridge variable in main:
//
//	var router *gin.Engine
//	Add(ginComp.Component(func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(httpComp.Component(func() http.Handler { return router }))
func Component(handler func() http.Handler) bootstrap.IComponent {
	return httpComp.ComponentWithRunner(handler, loadFromEnv)
}
