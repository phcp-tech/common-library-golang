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

import "time"

// Default HTTP client values; used when the corresponding Config field is zero.
const (
	defaultTimeout          = 10 * time.Second // default HTTP client request timeout
	defaultRetryMax         = 3                // default maximum number of retry attempts
	defaultRetryWaitTime    = 1 * time.Second  // default minimum wait time between retries
	defaultRetryMaxWaitTime = 30 * time.Second // default maximum wait time between retries
)

// Config holds all configuration for HttpClient (resty-based).
// Zero-value duration and count fields fall back to the package defaults above.
// The caller is responsible for reading values from env (or any other source)
// at the composition root so this package has no dependency on env or log.
type Config struct {
	Timeout            time.Duration // default: 10s
	RetryMax           int           // default: 3
	RetryWaitTime      time.Duration // default: 1s
	RetryMaxWaitTime   time.Duration // default: 30s
	InsecureSkipVerify bool          // skip TLS certificate verification; use only for internal services with self-signed certs

	// RetryOnServerErrors also retries on HTTP 429 and 5xx responses, not just
	// transport-level errors (connection refused, timeout, DNS failure, etc).
	// resty's own default retry condition only looks at the error returned by
	// the underlying http.Client.Do — a well-formed 500/503/429 response is
	// not a Go error, so without this flag RetryMax never fires for those.
	// Defaults to false: existing callers that only expect transport-level
	// retries keep that behavior; turning this on can noticeably lengthen a
	// call against a genuinely-down downstream (RetryMax attempts, each
	// waiting up to RetryMaxWaitTime) instead of failing fast.
	RetryOnServerErrors bool
}

// resolve returns a copy of cfg with zero-value fields replaced by defaults.
func (c Config) resolve() Config {
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}
	if c.RetryMax == 0 {
		c.RetryMax = defaultRetryMax
	}
	if c.RetryWaitTime == 0 {
		c.RetryWaitTime = defaultRetryWaitTime
	}
	if c.RetryMaxWaitTime == 0 {
		c.RetryMaxWaitTime = defaultRetryMaxWaitTime
	}
	return c
}
