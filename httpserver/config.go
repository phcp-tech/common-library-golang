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
// NOTE on WriteTimeout: net/http.Server treats a zero-or-negative WriteTimeout
// as "no limit" (required for large file downloads — a non-zero WriteTimeout
// applies to the entire response write duration and will kill long-running
// file transfers server-side regardless of the client's own timeout
// settings). But Config's own zero value already means "not set, apply the
// default" for every duration field via resolve() below, so a literal 0
// can't reach net/http as "no limit" — passing it just gets overwritten by
// defaultWriteTimeout. Use NoWriteTimeout instead to request no limit
// explicitly; see its doc comment.
const (
	defaultReadTimeout       = 30 * time.Second  // maximum duration for reading the full HTTP request
	defaultWriteTimeout      = 60 * time.Second  // maximum duration for writing the full HTTP response
	defaultIdleTimeout       = 120 * time.Second // maximum duration to keep an idle keep-alive connection
	defaultReadHeaderTimeout = 10 * time.Second  // maximum duration for reading the HTTP request headers

	// recommended graceful shutdown timeout for Runner.shutdown ctx; not enforced internally
	DefaultShutdownTimeout = 10 * time.Second

	// NoWriteTimeout, passed as Config.WriteTimeout, disables the write
	// timeout entirely (e.g. for a large file download endpoint). The Go
	// zero value can't be used for this — resolve() already treats a zero
	// WriteTimeout as "not set", replacing it with defaultWriteTimeout — so
	// this uses a negative duration instead. net/http.Server treats any
	// zero-or-negative WriteTimeout identically (it only ever calls
	// SetWriteDeadline when WriteTimeout > 0), so this is exactly as
	// unlimited as a literal 0 would be if net/http saw it directly.
	NoWriteTimeout = -1 * time.Nanosecond
)

// Config holds all configuration for the HTTP server runner.
// Zero values for duration fields fall back to the package defaults above,
// except WriteTimeout — see NoWriteTimeout to explicitly request no limit.
// CrtFile and KeyFile must both be non-empty to enable TLS.
type Config struct {
	Port    string // TCP port to listen on, e.g. "8080"
	CrtFile string // TLS certificate file path; empty disables TLS
	KeyFile string // TLS private key file path; empty disables TLS

	ReadTimeout       time.Duration // default: 30s
	WriteTimeout      time.Duration // default: 60s; pass NoWriteTimeout for unlimited (e.g. file downloads)
	IdleTimeout       time.Duration // default: 120s
	ReadHeaderTimeout time.Duration // default: 10s
}

// resolve returns a copy of cfg with any zero-value duration fields replaced
// by their package defaults. WriteTimeout is the one exception: a negative
// value (i.e. NoWriteTimeout) is preserved as an explicit "no limit" request
// rather than being treated as unset — see NoWriteTimeout's doc comment.
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
	return cfg
}
