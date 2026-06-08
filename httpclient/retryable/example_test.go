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

// package retryable_test demonstrates the public API from a caller's perspective.
//
// # When to use httpclient vs httpclient/retryable
//
// Use the parent [httpclient] package when you need a rich, service-to-service
// HTTP client with built-in JWT authentication, automatic JSON handling, and
// convenience methods (Get/Post/Put/Delete).
//
// Use this [retryable] sub-package when you need a standard *http.Client with
// retry capability to drop into existing code that already uses net/http directly.
package retryable_test

import (
	"fmt"
	"time"

	"github.com/phcp-tech/common-library-golang/httpclient/retryable"
)

// ExampleNewHttpClient shows how to create a retryable HttpClient with default settings.
// Log output is routed through slog.Default() — call log.InitLog() first
// to integrate with the project's structured logger.
func ExampleNewHttpClient() {
	cli := retryable.NewHttpClient()
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleNewHttpClient_customConfig shows how to override default timeout and retry settings.
// Zero-value fields fall back to package defaults (Timeout=10s, RetryMax=3).
func ExampleNewHttpClient_customConfig() {
	cli := retryable.NewHttpClient(retryable.Config{
		Timeout:  15 * time.Second,
		RetryMax: 5,
	})
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleNewHttpClient_insecureSkipVerify shows how to disable TLS certificate verification
// for internal services that use self-signed certificates.
// Do NOT use this in production for external services.
func ExampleNewHttpClient_insecureSkipVerify() {
	cli := retryable.NewHttpClient(retryable.Config{
		InsecureSkipVerify: true, // only for internal services with self-signed certs
	})
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleHttpClient_Client shows how to obtain the underlying retryablehttp.Client
// and its standard *http.Client for use with existing net/http code.
func ExampleHttpClient_Client() {
	cli := retryable.NewHttpClient()
	underlying := cli.Client()               // *retryablehttp.Client
	stdClient := underlying.StandardClient() // *http.Client with retry
	fmt.Println(stdClient != nil)
	// Output:
	// true
}
