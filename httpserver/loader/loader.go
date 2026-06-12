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

package loader

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// LoadFromEnv creates an HTTP Runner from the koanf env singleton.
// It reads app.runmode and http.server.port, then returns the appropriate Runner.
// The caller is responsible for starting the server (call Start in a goroutine)
// and stopping it (call Shutdown on signal).
func LoadFromEnv() httpserver.Runner {
	if strings.EqualFold(env.Env().String("app.runmode"), "aws_lambda") {
		slog.Info("Http server is running under AWS-LAMBDA")
		return lambda.NewHttpServer()
	}

	port := env.Env().String("http.server.port")
	slog.Info(fmt.Sprintf("Http server is running under Virtual Machine, listen on port %s", port))
	return httpserver.NewHttpServer(httpserver.Config{Port: port})
}
