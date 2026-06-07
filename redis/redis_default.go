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

import (
	"sync"
)

var (
	// Redis client instance and its once for initialization
	instance *RedisClient
	once     sync.Once
)

// InitDefault initialises the package-level default Redis client singleton using the
// provided configuration. It is safe to call multiple times; only the first call takes effect.
// InitDefault initialises the package-level default Redis client singleton using the
// provided configuration. It is safe to call multiple times; only the first call takes effect.
func InitDefault(cfg *Config) error {
	once.Do(func() {
		instance = NewRedisClient(cfg)
	})
	return nil
}

// Default returns a default singleton instance of the Redis client. If you want to use any other instance, please call NewRedisClient.
func Default() *RedisClient {
	return instance
}
