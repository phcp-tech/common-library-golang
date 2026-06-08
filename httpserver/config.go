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

import "time"

// Default HTTP server timeout values.
// NOTE: WriteTimeout must be 0 (no limit) to support large file downloads via
// /admin/files/download. A non-zero WriteTimeout applies to the entire response
// write duration and will kill long-running file transfers server-side regardless
// of the client's own timeout settings.
const (
	defaultReadTimeout       = 30 * time.Second  // maximum duration for reading the full HTTP request
	defaultWriteTimeout      = 60 * time.Second  // maximum duration for writing the full HTTP response; 0 means no limit (required for file downloads)
	defaultIdleTimeout       = 120 * time.Second // maximum duration to keep an idle keep-alive connection
	defaultReadHeaderTimeout = 10 * time.Second  // maximum duration for reading the HTTP request headers
	defaultShutdownTimeout   = 5 * time.Second   // graceful shutdown timeout
)

// Config holds all configuration for the HTTP server runner.
// Zero values for duration fields fall back to the package defaults above.
// CrtFile and KeyFile must both be non-empty to enable TLS.
type Config struct {
	Port    string // TCP port to listen on, e.g. "8080"
	CrtFile string // TLS certificate file path; empty disables TLS
	KeyFile string // TLS private key file path; empty disables TLS

	ReadTimeout       time.Duration // default: 30s
	WriteTimeout      time.Duration // default: 60s; set to 0 for unlimited (file downloads)
	IdleTimeout       time.Duration // default: 120s
	ReadHeaderTimeout time.Duration // default: 10s
	ShutdownTimeout   time.Duration // default: 5s
}

// resolve returns a copy of cfg with any zero-value duration fields replaced
// by their package defaults.
func (cfg Config) resolve() Config {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = defaultIdleTimeout
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}
	return cfg
}
