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

// package httpserver_test demonstrates the public API from a caller's perspective.
//
// # Choosing between httpserver and httpserver/lambda
//
// Use this [httpserver] package when the service runs on a VM, container,
// or Kubernetes pod — anywhere a long-lived TCP listener is appropriate.
//
// Use the [httpserver/lambda] sub-package when the service is deployed as an
// AWS Lambda function behind API Gateway. The key differences are:
//
//   - AWS Lambda receives HTTP-like requests as APIGatewayProxyRequest events,
//     not as real TCP connections. There is no port to bind to.
//   - The Lambda runner's Start calls lambda.Start (from aws-lambda-go) which
//     blocks until the Lambda runtime terminates the process; os.Exit is called
//     on runtime shutdown, so Shutdown is always a no-op.
//   - Importing the lambda sub-package pulls in the AWS Lambda SDK
//     (aws-lambda-go and aws-lambda-go-api-proxy). Services that never run on
//     Lambda should NOT import it to keep their binary free of unnecessary
//     dependencies.
//
// # Composition-root pattern
//
// The [Runner] interface is identical for both modes, so the composition root
// (application.go) is the only place that needs to know which runner to create.
// All other code just calls runner.Start / runner.Shutdown:
//
//	runner := httpserver.NewHttpServer(httpserver.Config{
//	    Port: env.Env().String("server.port"),
//	})
//	go func() {
//	    if err := runner.Start(ginRouter); err != nil {
//	        slog.Error("Http server stopped with error", "error", err)
//	    }
//	}()
//	// … wait for OS signal …
//	_ = runner.Shutdown(ctx)
package httpserver_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/phcp-tech/common-library-golang/httpserver"
)

// ExampleNewHttpServer shows how to create a plain HTTP runner and start the server.
// The caller is responsible for reading the port (and optional TLS paths) from
// configuration (e.g. env.Env()) at the composition root.
func ExampleNewHttpServer() {
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port: "8080",
	})
	fmt.Println(runner != nil)
	// Output:
	// true
}

// ExampleNewHttpServer_tls shows how to enable HTTPS with custom TLS certificate files.
// CrtFile and KeyFile must both be non-empty to activate TLS.
func ExampleNewHttpServer_tls() {
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port:    "8443",
		CrtFile: "/etc/ssl/server.crt",
		KeyFile: "/etc/ssl/server.key",
	})
	fmt.Println(runner != nil)
	// Output:
	// true
}

// ExampleNewHttpServer_customTimeouts shows how to override the default timeout values.
// Zero-value duration fields fall back to the package defaults:
// ReadTimeout=30s, WriteTimeout=60s, IdleTimeout=120s, ReadHeaderTimeout=10s.
// Set WriteTimeout to 0 to allow unlimited response time (e.g. file downloads).
func ExampleNewHttpServer_customTimeouts() {
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port:              "8080",
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      0, // unlimited — required for file downloads
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	})
	fmt.Println(runner != nil)
	// Output:
	// true
}

// ExampleRunner_Start shows how Start is used in the standard server loop.
// Start blocks until the server is stopped by a Shutdown call or a fatal error.
// It must be called in a goroutine so the rest of the application can continue.
// Start returns nil after a clean shutdown triggered by Shutdown().
// Port "0" lets the OS assign a free port automatically.
func ExampleRunner_Start() {
	runner := httpserver.NewHttpServer(httpserver.Config{Port: "0"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	startErr := make(chan error, 1)
	go func() { startErr <- runner.Start(handler) }()

	// Wait briefly for the server to bind, then shut it down.
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = runner.Shutdown(ctx)

	if err := <-startErr; err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("stopped cleanly")
	// Output:
	// stopped cleanly
}

// ExampleRunner_Shutdown shows the recommended composition-root pattern:
// start the server in a goroutine and shut it down gracefully on signal.
// Shutdown waits for in-flight requests to complete until ctx is cancelled.
// Port "0" lets the OS assign a free port automatically.
func ExampleRunner_Shutdown() {
	runner := httpserver.NewHttpServer(httpserver.Config{Port: "0"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	startErr := make(chan error, 1)
	go func() { startErr <- runner.Start(handler) }()

	// Wait briefly for the server to bind, then shut it down.
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := runner.Shutdown(ctx)
	fmt.Println(err)
	<-startErr // wait for the server goroutine to finish
	// Output:
	// <nil>
}
