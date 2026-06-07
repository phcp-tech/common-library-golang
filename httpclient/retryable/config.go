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

package retryable

import "time"

// Default values; used when the corresponding Config field is zero.
const (
	defaultTimeout  = 10 * time.Second // default HTTP client request timeout
	defaultRetryMax = 3                // default maximum number of retry attempts
)

// Config holds configuration for the retryablehttp-based Client.
// Wait time between retries is managed by retryablehttp's built-in exponential back-off.
// The caller is responsible for reading values from env (or any other source)
// at the composition root so this package has no dependency on env or log.
type Config struct {
	Timeout            time.Duration // default: 10s
	RetryMax           int           // default: 3
	InsecureSkipVerify bool          // skip TLS certificate verification; use only for internal services with self-signed certs
}

// resolve returns a copy of cfg with zero-value fields replaced by defaults.
func (c Config) resolve() Config {
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}
	if c.RetryMax == 0 {
		c.RetryMax = defaultRetryMax
	}
	return c
}
