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

// Package retryable wraps hashicorp/go-retryablehttp and exposes it as a
// standard *http.Client with configurable retry and timeout.
// Import this sub-package only when a standard *http.Client with retry is
// needed; for structured service-to-service calls with JWT and JSON use the
// parent httpclient package (resty-based) instead.
package retryable

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// HttpClient wraps a retryablehttp.Client with pre-configured retry and timeout.
type HttpClient struct {
	client *retryablehttp.Client
}

// NewHttpClient creates and returns a new HttpClient.
// Pass an optional Config to override defaults; omit for all-default settings.
// Log output is routed through slog.Default(), which is the project logger
// after log.InitLog() has been called.
func NewHttpClient(cfg ...Config) *HttpClient {
	c := Config{}
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c = c.resolve()

	rc := retryablehttp.NewClient()
	rc.RetryMax = c.RetryMax
	rc.HTTPClient = &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify}, //nolint:gosec
		},
	}
	// slogWrapper bridges retryablehttp.LeveledLogger (key-value pairs) to slog —
	// no fmt.Sprintf needed because both interfaces use the same key-value convention.
	rc.Logger = &slogWrapper{logger: slog.Default()}

	return &HttpClient{client: rc}
}

// Client returns the underlying retryablehttp.Client for advanced configuration
// or for obtaining a standard *http.Client via StandardClient().
func (c *HttpClient) Client() *retryablehttp.Client {
	return c.client
}

// slogWrapper bridges retryablehttp.LeveledLogger to slog.Logger.
// retryablehttp passes key-value pairs (e.g. Error("msg", "key", val)),
// which is directly compatible with slog — no format conversion needed.
type slogWrapper struct {
	logger *slog.Logger
}

// Error implements retryablehttp.LeveledLogger.
func (w *slogWrapper) Error(msg string, args ...interface{}) { w.logger.Error(msg, args...) }

// Warn implements retryablehttp.LeveledLogger.
func (w *slogWrapper) Warn(msg string, args ...interface{}) { w.logger.Warn(msg, args...) }

// Info implements retryablehttp.LeveledLogger.
func (w *slogWrapper) Info(msg string, args ...interface{}) { w.logger.Info(msg, args...) }

// Debug implements retryablehttp.LeveledLogger.
func (w *slogWrapper) Debug(msg string, args ...interface{}) { w.logger.Debug(msg, args...) }
