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

package loader_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/redis"
	"github.com/phcp-tech/common-library-golang/redis/loader"
)

// ExampleLoadFromEnv shows the typical call-site pattern for LoadFromEnv.
// It reads Redis connection parameters from the koanf env singleton
// (keys: redis.clusters, redis.database, redis.password) and initialises
// the package-level default client via redis.InitDefault.
//
// go-redis creates the client without establishing a connection (lazy dial),
// so LoadFromEnv always returns nil — connectivity errors are discovered
// on the first real command instead.
//
// redis.InitDefault uses sync.Once: only the first call takes effect.
// After LoadFromEnv returns, redis.Default() is non-nil regardless of
// whether the Redis server is reachable.
func ExampleLoadFromEnv() {
	err := loader.LoadFromEnv()
	fmt.Println(err)                    // <nil> — go-redis is lazy, no immediate connect
	fmt.Println(redis.Default() != nil) // true  — singleton registered
	// Output:
	// <nil>
	// true
}
