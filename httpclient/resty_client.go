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

package httpclient

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// HttpClient wraps a resty.Client with pre-configured timeout, retry, and TLS settings.
type HttpClient struct {
	httpClient *resty.Client
}

// NewHttpClient creates and returns a new HttpClient.
// Pass an optional Config to override defaults; omit for all-default settings.
// Timeout, retry count, and wait times are configurable. TLS certificate
// verification can be disabled via InsecureSkipVerify (only for internal services
// with self-signed certificates).
// Log output is routed through slog.Default(), which is the project logger after
// log.InitLog() has been called.
func NewHttpClient(cfg ...Config) *HttpClient {
	c := Config{}
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c = c.resolve()

	client := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: c.InsecureSkipVerify}). //nolint:gosec
		SetTimeout(c.Timeout).
		// Set retry count to non zero to enable retries
		SetRetryCount(c.RetryMax).
		// WaitTime Default is 1 second.
		SetRetryWaitTime(c.RetryWaitTime).
		// MaxWaitTime Default is 30 seconds.
		SetRetryMaxWaitTime(c.RetryMaxWaitTime).
		SetLogger(&slogWrapper{Logger: slog.Default()})

	// Without this, resty's default retry condition only fires on a
	// transport-level error (err != nil) — a 429/5xx response is a
	// successful round-trip as far as resty is concerned, so RetryCount
	// alone never retries those. See the doc comment on RetryOnServerErrors.
	if c.RetryOnServerErrors {
		client.AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil ||
				r.StatusCode() == http.StatusTooManyRequests ||
				r.StatusCode() >= http.StatusInternalServerError
		})
	}

	return &HttpClient{httpClient: client}
}

// Client returns the underlying resty.Client for advanced configuration or direct use.
func (cli *HttpClient) Client() *resty.Client {
	return cli.httpClient
}

// Get sends an HTTP GET request to the given URL with the provided bearer token and request body,
// returning the response or an error.
func (cli *HttpClient) Get(url string, token string, body any) (*resty.Response, error) {
	return cli.httpClient.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetAuthToken(token).
		SetBody(body).
		Get(url)
}

// Post sends an HTTP POST request to the given URL with the provided bearer token and JSON body,
// returning the response or an error.
func (cli *HttpClient) Post(url string, token string, body any) (*resty.Response, error) {
	return cli.httpClient.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetAuthToken(token).
		SetBody(body).
		Post(url)
}

// Put sends an HTTP PUT request to the given URL with the provided bearer token and JSON body,
// returning the response or an error.
func (cli *HttpClient) Put(url string, token string, body any) (*resty.Response, error) {
	return cli.httpClient.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetAuthToken(token).
		SetBody(body).
		Put(url)
}

// Delete sends an HTTP DELETE request to the given URL with the provided bearer token and JSON body,
// returning the response or an error.
func (cli *HttpClient) Delete(url string, token string, body any) (*resty.Response, error) {
	return cli.httpClient.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetAuthToken(token).
		SetBody(body).
		Delete(url)
}

// slogWrapper adapts slog.Logger to the resty.Logger interface.
// Resty calls Errorf/Warnf/Debugf with printf-style format strings, so each
// method uses fmt.Sprintf to produce the final message before passing it to slog.
type slogWrapper struct {
	Logger *slog.Logger
}

// Errorf implements the resty.Logger interface.
func (wap *slogWrapper) Errorf(msg string, args ...interface{}) {
	wap.Logger.Error(fmt.Sprintf(msg, args...))
}

// Warnf implements the resty.Logger interface.
func (wap *slogWrapper) Warnf(msg string, args ...interface{}) {
	wap.Logger.Warn(fmt.Sprintf(msg, args...))
}

// Debugf implements the resty.Logger interface.
func (wap *slogWrapper) Debugf(msg string, args ...interface{}) {
	wap.Logger.Debug(fmt.Sprintf(msg, args...))
}
