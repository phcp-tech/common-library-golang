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

package redis

import "time"

const (
	// defaultPoolSize is the default number of socket connections per node.
	defaultPoolSize = 100

	// defaultScanCount is the default page size used by SCAN in GetKeysCount.
	defaultScanCount = 1000
	// defaultGetKeysTimeout is the default overall timeout for GetKeysCount.
	defaultGetKeysTimeout = 10 * time.Second
)

// Config holds connection and pool settings for a Redis client.
// The caller is responsible for reading values from env (or any other source)
// at the composition root so this package has no dependency on env.
type Config struct {
	Clusters     []string // Redis node addresses; more than one address enables cluster mode
	DB           int      // database index (ignored in cluster mode)
	Password     string   // authentication password
	PoolSize     int      // max socket connections per node; zero falls back to default: 100
	MinIdleConns int      // minimum idle connections to maintain; zero is used as-is (go-redis's own native "no forced minimum" behavior) — there is no way to request the old default via this field
}

// resolve returns a copy of cfg with a zero PoolSize replaced by the package
// default. MinIdleConns is passed through unchanged: zero is a legitimate,
// meaningful value (no eagerly pre-warmed idle connections — see
// NewRedisClient's ConnPool, which dials MinIdleConns connections up front),
// not a "use the default" sentinel, so callers that want 0 idle connections
// (e.g. one goroutine-per-invocation environments like AWS Lambda, where an
// eagerly-warmed pool is pure per-container connection overhead) can actually
// get it.
func (c Config) resolve() Config {
	if c.PoolSize == 0 {
		c.PoolSize = defaultPoolSize
	}
	return c
}
