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
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps a go-redis universal client and provides a simplified API
// for common Redis operations, supporting both standalone and cluster modes.
type RedisClient struct {
	rds       redis.UniversalClient
	isCluster bool // true when more than one address is configured
}

// NewRedisClient creates a new RedisClient using the provided configuration.
// Zero-value PoolSize and MinIdleConns fall back to the package defaults.
func NewRedisClient(conf *Config) *RedisClient {
	c := conf.resolve()

	rds := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    c.Clusters,
		DB:       c.DB,
		Password: c.Password,
		// Pool size settings
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
		// Don't set MaxIdleConns/MaxActiveConns, default is 0, means no limit

		/*
			// Optimized configuration for high-concurrency scenarios
			// Connection pool settings to prevent "connection pool timeout"
			PoolSize:        200,              // Maximum number of socket connections per node (default: 10 * runtime.GOMAXPROCS)
			PoolTimeout:     10 * time.Second, // Pool timeout when getting connection from pool (default: ReadTimeout + 1 second)
			MinIdleConns:    5,                // Minimum number of idle connections to maintain (default: 0)
			MaxIdleConns:    10,               // Maximum number of idle connections (default: 0, means no limit)
			MaxActiveConns:  0,                // Maximum number of active connections (default: 0, means no limit)
			ConnMaxIdleTime: 5 * time.Minute,  // Close connections after remaining idle for this duration (default: 30 minutes)
			ConnMaxLifetime: 1 * time.Hour,    // Close connections after this lifetime (default: 0, means no limit)

			// Basic timeout settings
			DialTimeout:  5 * time.Second, // Timeout for establishing new connections
			ReadTimeout:  3 * time.Second, // Timeout for socket reads
			WriteTimeout: 3 * time.Second, // Timeout for socket writes

			// Retry settings
			MaxRetries:      3,                      // Maximum number of retries before giving up
			MinRetryBackoff: 8 * time.Millisecond,   // Minimum backoff between each retry
			MaxRetryBackoff: 512 * time.Millisecond, // Maximum backoff between each retry
		*/
	})
	return &RedisClient{rds: rds, isCluster: len(c.Clusters) > 1}
}

// Set returns "OK" on success, or an error if the operation fails
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (result string, err error) {
	// Rely on go-redis built-in timeouts (WriteTimeout: 3s, ReadTimeout: 3s)
	return c.rds.Set(ctx, key, value, expiration).Result()
}

// Exists returns the number of keys. If the key does not exist, count is 0 and err is nil.
func (c *RedisClient) Exists(ctx context.Context, key ...string) (count int64, err error) {
	return c.rds.Exists(ctx, key...).Result()
}

// Get returns the value for the key, or an error if the key does not exist. The error will be redis.Nil with message "redis: nil".
func (c *RedisClient) Get(ctx context.Context, key string) (value string, err error) {
	return c.rds.Get(ctx, key).Result()
}

// Del returns the number of keys deleted. If the key does not exist, count is 0 and err is nil.
func (c *RedisClient) Del(ctx context.Context, key ...string) (count int64, err error) {
	return c.rds.Del(ctx, key...).Result()
}

// Unlink returns the number of keys unlinked. If the key does not exist, count is 0 and err is nil.
// Unlike Del, Unlink performs asynchronous deletion in a separate thread, making it non-blocking for large keys.
func (c *RedisClient) Unlink(ctx context.Context, key ...string) (count int64, err error) {
	return c.rds.Unlink(ctx, key...).Result()
}

// Ping checks connectivity to the Redis server and returns "PONG" on success,
// or an error if the client is not initialised or the server is unreachable.
func (c *RedisClient) Ping(ctx context.Context) (string, error) {
	if c.rds == nil {
		return "", errors.New("redis client not initialized")
	}
	return c.rds.Ping(ctx).Result()
}

// Close closes the underlying Redis connection if it is open.
func (c *RedisClient) Close() {
	if c.rds != nil {
		c.rds.Close()
	}
}

// GetKeysCount returns the total number of keys matching the given pattern across
// all Redis nodes. It automatically handles both standalone and cluster deployments.
func (c *RedisClient) GetKeysCount(ctx context.Context, key string) (int, error) {
	// Set a context with timeout to control the overall operation duration
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultGetKeysTimeout)
	defer cancel()

	if c.isCluster {
		return getKeysCountInClusters(timeoutCtx, key, c.rds.(*redis.ClusterClient))
	}
	return getKeysCount(timeoutCtx, key, c.rds.(*redis.Client))
}

func getKeysCountInClusters(ctx context.Context, key string, rdsc *redis.ClusterClient) (int, error) {
	var keys int
	if err := rdsc.ForEachMaster(ctx, func(ctx context.Context, rds *redis.Client) error {
		if count, err := getKeysCount(ctx, key, rds); err == nil {
			keys += count
		} else {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return keys, nil
}

func getKeysCount(ctx context.Context, key string, rds *redis.Client) (int, error) {
	var cursor uint64
	var keys int
	for {
		var (
			k   []string
			err error
		)
		if k, cursor, err = rds.Scan(ctx, cursor, key, defaultScanCount).Result(); err == nil {
			keys += len(k)
			// no keys or scan done
			if len(k) == 0 || cursor == 0 {
				break
			}
		} else {
			return 0, err
		}
	}
	return keys, nil
}

// CleanCache removes all keys whose names begin with the given prefix.
// It uses SCAN to iterate without blocking and Unlink for async deletion.
// Both standalone and cluster deployments are handled automatically.
func (c *RedisClient) CleanCache(ctx context.Context, key string) error {
	pattern := key + "*"
	if c.isCluster {
		return cleanCacheInClusters(ctx, pattern, c.rds.(*redis.ClusterClient))
	}
	return cleanCache(ctx, pattern, c.rds.(*redis.Client))
}

func cleanCacheInClusters(ctx context.Context, pattern string, rdsc *redis.ClusterClient) error {
	return rdsc.ForEachMaster(ctx, func(ctx context.Context, rds *redis.Client) error {
		return cleanCache(ctx, pattern, rds)
	})
}

func cleanCache(ctx context.Context, pattern string, rds *redis.Client) error {
	var cursor uint64
	for {
		keys, nextCursor, err := rds.Scan(ctx, cursor, pattern, defaultScanCount).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := rds.Unlink(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
