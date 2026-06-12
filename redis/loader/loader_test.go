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
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/redis"
	"github.com/phcp-tech/common-library-golang/redis/loader"
)

// TestMain initialises the env singleton before all tests.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("redis/loader tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestLoadFromEnv_ReturnsNil verifies that LoadFromEnv always returns nil.
//
// redis.InitDefault calls NewRedisClient which creates a go-redis client but
// does NOT ping the server — the connection is lazy. As a result InitDefault
// always returns nil regardless of whether the Redis address is reachable.
//
// redis.InitDefault uses sync.Once — this is the sole LoadFromEnv call in
// this test binary.
func TestLoadFromEnv_ReturnsNil(t *testing.T) {
	err := loader.LoadFromEnv()
	if err != nil {
		t.Errorf("LoadFromEnv() returned unexpected error: %v", err)
	}
}

// TestLoadFromEnv_DefaultClientIsNonNil verifies that redis.Default() is
// non-nil after LoadFromEnv succeeds, confirming the singleton was set.
func TestLoadFromEnv_DefaultClientIsNonNil(t *testing.T) {
	loader.LoadFromEnv() //nolint:errcheck — always nil, see TestLoadFromEnv_ReturnsNil
	if redis.Default() == nil {
		t.Error("redis.Default() should be non-nil after LoadFromEnv")
	}
}
