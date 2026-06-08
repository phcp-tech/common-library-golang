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
	"crypto/tls"
	"net/http"

	"github.com/pkg/errors"
)

// Compile-time check that httpRunner implements httpserver.Runner.
var _ Runner = (*httpRunner)(nil)

// httpRunner implements Runner for plain HTTP and HTTPS servers.
type httpRunner struct {
	cfg    Config
	server *http.Server
}

// NewHttpServer returns a plain HTTP/HTTPS Runner configured by cfg.
// Zero-value duration fields in cfg fall back to the package defaults.
// For AWS Lambda mode import and use the httpserver/lambda subpackage instead.
func NewHttpServer(cfg Config) Runner {
	return &httpRunner{cfg: cfg.resolve()}
}

func (r *httpRunner) Start(handler http.Handler) error {
	r.server = &http.Server{
		Addr:    ":" + r.cfg.Port,
		Handler: handler,
		// Add timeouts to prevent goroutine leaks
		ReadTimeout:       r.cfg.ReadTimeout,
		WriteTimeout:      r.cfg.WriteTimeout,
		IdleTimeout:       r.cfg.IdleTimeout,
		ReadHeaderTimeout: r.cfg.ReadHeaderTimeout,
	}

	// non-TLS service connections. http.ErrServerClosed means server is closed by Shutdown
	if r.cfg.CrtFile == "" || r.cfg.KeyFile == "" {
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return errors.Wrap(err, "listen error")
		}
		return nil
	}

	// TLS service connections — configure secure cipher suites and minimum TLS 1.2
	r.server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	if err := r.server.ListenAndServeTLS(r.cfg.CrtFile, r.cfg.KeyFile); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "listen TLS error")
	}
	return nil
}

func (r *httpRunner) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Shutdown(ctx)
}
