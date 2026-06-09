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

package httpserver

import (
	"context"
	"net/http"
)

// Runner abstracts a server that can be started and gracefully stopped.
// The two concrete implementations are:
//   - httpRunner (this package) — plain HTTP/HTTPS server
//   - httpserver/lambda.NewHttpServer() — AWS Lambda adapter (unexported lambdaRunner)
//
// Import httpserver/lambda only when Lambda support is required; it carries
// the AWS SDK dependencies and does not affect callers that only need HTTP.
type Runner interface {
	// Start starts the server and blocks until it is stopped.
	// Returns nil after a graceful shutdown; returns a non-nil error on failure.
	Start(handler http.Handler) error

	// Shutdown gracefully stops the server, waiting for in-flight requests to complete until ctx is cancelled or times out.
	Shutdown(ctx context.Context) error
}
