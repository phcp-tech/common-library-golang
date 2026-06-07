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

// package httpclient_test demonstrates the public API from a caller's perspective.
package httpclient_test

import (
	"fmt"
	"time"

	"github.com/phcp-tech/common-library-golang/httpclient"
)

// ExampleNewHttpClient shows how to create an HttpClient with default settings.
// Log output is routed through slog.Default() — call log.InitLog() first
// to integrate with the project's structured logger.
func ExampleNewHttpClient() {
	cli := httpclient.NewHttpClient()
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleNewHttpClient_customConfig shows how to override default timeout and retry settings.
// Pass a Config with only the fields you want to change; zero-value fields
// fall back to the package defaults (Timeout=10s, RetryMax=3).
func ExampleNewHttpClient_customConfig() {
	cli := httpclient.NewHttpClient(httpclient.Config{
		Timeout:          15 * time.Second,
		RetryMax:         5,
		RetryWaitTime:    2 * time.Second,
		RetryMaxWaitTime: 60 * time.Second,
	})
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleNewHttpClient_insecureSkipVerify shows how to disable TLS certificate verification
// for internal services that use self-signed certificates.
// Do NOT use this in production for external services.
func ExampleNewHttpClient_insecureSkipVerify() {
	cli := httpclient.NewHttpClient(httpclient.Config{
		InsecureSkipVerify: true, // only for internal services with self-signed certs
	})
	fmt.Println(cli != nil)
	// Output:
	// true
}

// ExampleHttpClient_Client shows how to obtain the underlying resty.Client
// for advanced configuration not exposed by Config.
func ExampleHttpClient_Client() {
	cli := httpclient.NewHttpClient()
	restyClient := cli.Client() // *resty.Client
	fmt.Println(restyClient != nil)
	// Output:
	// true
}
