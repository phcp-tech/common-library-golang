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

// package redis_test demonstrates the public API from a caller's perspective.
package redis_test

import (
	"context"
	"fmt"
	"time"

	"github.com/phcp-tech/common-library-golang/health"
	"github.com/phcp-tech/common-library-golang/redis"
)

// ExampleNewRedisClient shows how to create a standalone Redis client.
// The caller is responsible for reading configuration from env (or any other
// source) at the composition root — this package has no dependency on env.
func ExampleNewRedisClient() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters: []string{"127.0.0.1:6379"},
		DB:       0,
		Password: "",
	})
	defer cli.Close()
}

// ExampleNewRedisClient_cluster shows how to connect to a Redis cluster.
// Providing more than one address automatically enables cluster mode.
func ExampleNewRedisClient_cluster() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters: []string{
			"127.0.0.1:7000",
			"127.0.0.1:7001",
			"127.0.0.1:7002",
		},
		PoolSize:     200,
		MinIdleConns: 10,
	})
	defer cli.Close()
}

// ExampleNewRedisClient_customPool shows how to override the default connection
// pool settings. Zero-value fields fall back to package defaults
// (PoolSize=100, MinIdleConns=5).
func ExampleNewRedisClient_customPool() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters:     []string{"127.0.0.1:6379"},
		PoolSize:     50,
		MinIdleConns: 2,
	})
	defer cli.Close()
}

// ExampleInitDefault shows the singleton pattern: call InitDefault once at
// application startup. Subsequent calls are silently ignored (sync.Once).
func ExampleInitDefault() {
	if err := redis.InitDefault(&redis.Config{
		Clusters: []string{"127.0.0.1:6379"},
	}); err != nil {
		// handle error
		return
	}
	_ = redis.Default() // *RedisClient, ready to use
}

// ExampleRedisClient_Set shows how to store a key-value pair with an optional TTL.
// Pass 0 as the expiration to store the key without expiry.
func ExampleRedisClient_Set() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters: []string{"127.0.0.1:6379"},
	})
	defer cli.Close()

	ctx := context.Background()

	// Store without TTL.
	_, _ = cli.Set(ctx, "greeting", "hello", 0)

	// Store with 5-minute TTL.
	_, _ = cli.Set(ctx, "session:abc", "data", 5*time.Minute)
}

// ExampleRedisClient_CleanCache shows how to remove all keys that share a
// common prefix. Useful for invalidating a group of cached entries at once.
func ExampleRedisClient_CleanCache() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters: []string{"127.0.0.1:6379"},
	})
	defer cli.Close()

	// Remove all keys whose names start with "user:42:".
	_ = cli.CleanCache(context.Background(), "user:42:")
}

// ExampleRedisClient_GetKeysCount shows how to count all keys matching a
// pattern across all Redis nodes (standalone or cluster).
func ExampleRedisClient_GetKeysCount() {
	cli := redis.NewRedisClient(&redis.Config{
		Clusters: []string{"127.0.0.1:6379"},
	})
	defer cli.Close()

	_, _ = cli.GetKeysCount(context.Background(), "session:*")
}

// ExampleHealthChecker shows how to wire HealthChecker into a [health.Check] call
// for a /health HTTP endpoint.
// In test environments with no reachable Redis server, the Checker reports StatusUnhealthy.
func ExampleHealthChecker() {
	results := health.Check(context.Background(), redis.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — no reachable Redis
	// Output:
	// redis
	// true
}
