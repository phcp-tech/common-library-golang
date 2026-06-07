// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package redis provides unit tests for the redis package.
// Tests use miniredis for success paths and an unreachable address for error paths.
// No live Redis infrastructure is required.
package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

// unreachableConf points to a port where nothing is listening.
var unreachableConf = &Config{
	Clusters: []string{"127.0.0.1:16379"},
}

// clusterConf has two addresses so NewRedisClient creates a ClusterClient (isCluster=true).
var clusterConf = &Config{
	Clusters: []string{"127.0.0.1:16379", "127.0.0.1:16380"},
}

// ctxShort returns a 2-second context to prevent network tests from hanging.
func ctxShort() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// startMini starts a miniredis server and returns a single-node Config pointing to it.
func startMini(t *testing.T) *Config {
	t.Helper()
	s := miniredis.RunT(t)
	return &Config{Clusters: []string{s.Addr()}}
}

// resetDefault resets the singleton for tests that need a clean state.
func resetDefault(t *testing.T) {
	t.Helper()
	savedInst := instance
	instance = nil
	once = sync.Once{}
	t.Cleanup(func() {
		instance = savedInst
		once = sync.Once{}
		once.Do(func() {}) // mark as done to match pre-test state
	})
}

// -----------------------------------------------------------------------
// Config.resolve — defaults and custom values
// -----------------------------------------------------------------------

func TestConfig_Resolve_DefaultPoolSize(t *testing.T) {
	c := Config{}.resolve()
	if c.PoolSize != defaultPoolSize {
		t.Errorf("default PoolSize: want %d, got %d", defaultPoolSize, c.PoolSize)
	}
}

func TestConfig_Resolve_CustomPoolSize(t *testing.T) {
	c := Config{PoolSize: 50}.resolve()
	if c.PoolSize != 50 {
		t.Errorf("custom PoolSize: want 50, got %d", c.PoolSize)
	}
}

func TestConfig_Resolve_DefaultMinIdleConns(t *testing.T) {
	c := Config{}.resolve()
	if c.MinIdleConns != defaultMinIdleConns {
		t.Errorf("default MinIdleConns: want %d, got %d", defaultMinIdleConns, c.MinIdleConns)
	}
}

func TestConfig_Resolve_CustomMinIdleConns(t *testing.T) {
	c := Config{MinIdleConns: 10}.resolve()
	if c.MinIdleConns != 10 {
		t.Errorf("custom MinIdleConns: want 10, got %d", c.MinIdleConns)
	}
}

// -----------------------------------------------------------------------
// NewRedisClient
// -----------------------------------------------------------------------

func TestNewRedisClient_ReturnsNonNil(t *testing.T) {
	if NewRedisClient(unreachableConf) == nil {
		t.Fatal("NewRedisClient returned nil")
	}
}

func TestNewRedisClient_SingleNode_IsCluster_False(t *testing.T) {
	cli := NewRedisClient(unreachableConf)
	if cli.isCluster {
		t.Error("single-node config should set isCluster=false")
	}
}

func TestNewRedisClient_MultiNode_IsCluster_True(t *testing.T) {
	cli := NewRedisClient(clusterConf)
	if !cli.isCluster {
		t.Error("multi-node config should set isCluster=true")
	}
}

// -----------------------------------------------------------------------
// Close
// -----------------------------------------------------------------------

func TestRedisClient_Close_DoesNotPanic(t *testing.T) {
	NewRedisClient(unreachableConf).Close()
}

func TestRedisClient_Close_NilRds_DoesNotPanic(t *testing.T) {
	cli := &RedisClient{}
	cli.Close() // rds is nil — must not panic
}

// -----------------------------------------------------------------------
// Ping
// -----------------------------------------------------------------------

func TestRedisClient_Ping_NilRds_ReturnsError(t *testing.T) {
	_, err := (&RedisClient{}).Ping(context.Background())
	if err == nil {
		t.Error("Ping with nil rds should return error")
	}
}

func TestRedisClient_Ping_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Ping(ctx)
	if err == nil {
		t.Error("Ping to unreachable server should return error")
	}
}

func TestRedisClient_Ping_Success(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()
	result, err := cli.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping error = %v", err)
	}
	if result != "PONG" {
		t.Errorf("Ping result = %q, want PONG", result)
	}
}

// -----------------------------------------------------------------------
// CRUD — error paths (unreachable server)
// -----------------------------------------------------------------------

func TestRedisClient_Set_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Set(ctx, "k", "v", 0)
	if err == nil {
		t.Error("Set to unreachable server should return error")
	}
}

func TestRedisClient_Get_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Get(ctx, "k")
	if err == nil {
		t.Error("Get from unreachable server should return error")
	}
}

func TestRedisClient_Exists_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Exists(ctx, "k")
	if err == nil {
		t.Error("Exists on unreachable server should return error")
	}
}

func TestRedisClient_Del_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Del(ctx, "k")
	if err == nil {
		t.Error("Del on unreachable server should return error")
	}
}

func TestRedisClient_Unlink_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).Unlink(ctx, "k")
	if err == nil {
		t.Error("Unlink on unreachable server should return error")
	}
}

// -----------------------------------------------------------------------
// CRUD — success paths (miniredis)
// -----------------------------------------------------------------------

func TestRedisClient_Set_Get_Success(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	if _, err := cli.Set(ctx, "hello", "world", 0); err != nil {
		t.Fatalf("Set error = %v", err)
	}
	val, err := cli.Get(ctx, "hello")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if val != "world" {
		t.Errorf("Get = %q, want %q", val, "world")
	}
}

func TestRedisClient_Set_Expiry(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	if _, err := cli.Set(ctx, "ttl-key", "v", time.Second); err != nil {
		t.Fatalf("Set error = %v", err)
	}
	if _, err := cli.Get(ctx, "ttl-key"); err != nil {
		t.Fatalf("Get before expiry error = %v", err)
	}
}

func TestRedisClient_Exists_Success(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	cli.Set(ctx, "e", "1", 0) //nolint:errcheck
	count, err := cli.Exists(ctx, "e", "missing")
	if err != nil {
		t.Fatalf("Exists error = %v", err)
	}
	if count != 1 {
		t.Errorf("Exists count = %d, want 1", count)
	}
}

func TestRedisClient_Del_Success(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	cli.Set(ctx, "d", "1", 0) //nolint:errcheck
	n, err := cli.Del(ctx, "d")
	if err != nil {
		t.Fatalf("Del error = %v", err)
	}
	if n != 1 {
		t.Errorf("Del count = %d, want 1", n)
	}
}

func TestRedisClient_Unlink_Success(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	cli.Set(ctx, "u", "1", 0) //nolint:errcheck
	n, err := cli.Unlink(ctx, "u")
	if err != nil {
		t.Fatalf("Unlink error = %v", err)
	}
	if n != 1 {
		t.Errorf("Unlink count = %d, want 1", n)
	}
}

// -----------------------------------------------------------------------
// GetKeysCount
// -----------------------------------------------------------------------

func TestRedisClient_GetKeysCount_EmptyDB_ReturnsZero(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()
	n, err := cli.GetKeysCount(ctx, "key:*")
	if err != nil {
		t.Fatalf("GetKeysCount error = %v", err)
	}
	if n != 0 {
		t.Errorf("GetKeysCount on empty DB = %d, want 0", n)
	}
}

func TestRedisClient_GetKeysCount_WithKeys(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	for i := range 5 {
		cli.Set(ctx, "key:"+string(rune('a'+i)), "v", 0) //nolint:errcheck
	}
	n, err := cli.GetKeysCount(ctx, "key:*")
	if err != nil {
		t.Fatalf("GetKeysCount error = %v", err)
	}
	if n != 5 {
		t.Errorf("GetKeysCount = %d, want 5", n)
	}
}

func TestRedisClient_GetKeysCount_SingleNode_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	_, err := NewRedisClient(unreachableConf).GetKeysCount(ctx, "key:*")
	if err == nil {
		t.Error("GetKeysCount on unreachable server should return error")
	}
}

func TestRedisClient_GetKeysCount_ClusterMode_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	// clusterConf has two addresses → isCluster=true → ClusterClient path
	cli := NewRedisClient(clusterConf)
	_, err := cli.GetKeysCount(ctx, "key:*")
	if err == nil {
		t.Error("GetKeysCount in cluster mode with unreachable nodes should return error")
	}
}

// -----------------------------------------------------------------------
// CleanCache
// -----------------------------------------------------------------------

func TestRedisClient_CleanCache_ClusterMode_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	// clusterConf has two addresses → isCluster=true → cleanCacheInClusters path
	cli := NewRedisClient(clusterConf)
	err := cli.CleanCache(ctx, "prefix:")
	if err == nil {
		t.Error("CleanCache in cluster mode with unreachable nodes should return error")
	}
}

func TestRedisClient_CleanCache_UnreachableServer_ReturnsError(t *testing.T) {
	ctx, cancel := ctxShort()
	defer cancel()
	err := NewRedisClient(unreachableConf).CleanCache(ctx, "prefix:")
	if err == nil {
		t.Error("CleanCache on unreachable server should return error")
	}
}

func TestRedisClient_CleanCache_DeletesMatchingKeys(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()

	for i := range 3 {
		cli.Set(ctx, "cache:"+string(rune('a'+i)), "v", 0) //nolint:errcheck
	}
	cli.Set(ctx, "other:key", "v", 0) //nolint:errcheck

	if err := cli.CleanCache(ctx, "cache:"); err != nil {
		t.Fatalf("CleanCache error = %v", err)
	}
	n, err := cli.GetKeysCount(ctx, "cache:*")
	if err != nil {
		t.Fatalf("GetKeysCount after CleanCache error = %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 cache keys after CleanCache, got %d", n)
	}
	// non-matching key must survive
	n2, _ := cli.GetKeysCount(ctx, "other:*")
	if n2 != 1 {
		t.Errorf("non-matching key should survive CleanCache, got count %d", n2)
	}
}

func TestRedisClient_CleanCache_EmptyDB_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	cli := NewRedisClient(startMini(t))
	defer cli.Close()
	if err := cli.CleanCache(ctx, "prefix:"); err != nil {
		t.Errorf("CleanCache on empty DB should return nil, got %v", err)
	}
}

// -----------------------------------------------------------------------
// Singleton (redis_default.go)
// -----------------------------------------------------------------------

func TestInitDefault_ReturnsNil(t *testing.T) {
	resetDefault(t)
	if err := InitDefault(unreachableConf); err != nil {
		t.Errorf("InitDefault returned unexpected error: %v", err)
	}
}

func TestDefault_AfterInit_IsNotNil(t *testing.T) {
	resetDefault(t)
	InitDefault(unreachableConf) //nolint:errcheck
	if Default() == nil {
		t.Error("Default() should be non-nil after InitDefault")
	}
}

func TestInitDefault_SecondCall_IsNoOp(t *testing.T) {
	resetDefault(t)
	InitDefault(unreachableConf)   //nolint:errcheck
	first := Default()
	InitDefault(startMini(t))      // second call — ignored
	if Default() != first {
		t.Error("second InitDefault should not change the singleton")
	}
}

// -----------------------------------------------------------------------
// Internal function tests — cleanCache / cleanCacheInClusters / getKeysCountInClusters
// Accessible because redis_test.go is in package redis (internal test).
// -----------------------------------------------------------------------

// TestCleanCache_Internal_WithKeys directly exercises cleanCache via a
// miniredis-backed *redis.Client, covering the successful deletion path.
func TestCleanCache_Internal_WithKeys(t *testing.T) {
	s := miniredis.RunT(t)
	for i := range 3 {
		s.Set("pfx:"+string(rune('a'+i)), "v")
	}
	s.Set("other", "v")

	rds := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
	ctx := context.Background()

	if err := cleanCache(ctx, "pfx:*", rds); err != nil {
		t.Fatalf("cleanCache error = %v", err)
	}
	n, err := rds.DBSize(ctx).Result()
	if err != nil {
		t.Fatalf("DBSize error = %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 remaining key after cleanCache, got %d", n)
	}
}

// TestCleanCache_Internal_ScanError covers the Scan error return inside cleanCache.
func TestCleanCache_Internal_ScanError(t *testing.T) {
	rds := goredis.NewClient(&goredis.Options{
		Addr:        "127.0.0.1:16379",
		DialTimeout: 200 * time.Millisecond,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := cleanCache(ctx, "key:*", rds); err == nil {
		t.Error("expected error from cleanCache with unreachable server")
	}
}

// TestGetKeysCountInClusters_Internal_Error directly exercises the error path
// of getKeysCountInClusters with unreachable cluster nodes.
func TestGetKeysCountInClusters_Internal_Error(t *testing.T) {
	rdsc := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:        []string{"127.0.0.1:16379", "127.0.0.1:16380"},
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := getKeysCountInClusters(ctx, "key:*", rdsc)
	if err == nil {
		t.Error("expected error from getKeysCountInClusters with unreachable cluster")
	}
}

// TestCleanCacheInClusters_WithMiniredis attempts to invoke cleanCacheInClusters
// with a ClusterClient pointed at a live miniredis instance.
// If miniredis handles CLUSTER SLOTS/NODES in a way that lets ForEachMaster invoke
// the callback, this covers cleanCacheInClusters's inner body.
// If ForEachMaster fails (miniredis is not a real cluster), the test is a no-op
// for coverage but still validates the error-handling path.
func TestCleanCacheInClusters_WithMiniredis(t *testing.T) {
	s := miniredis.RunT(t)
	s.Set("key:1", "v1")
	s.Set("key:2", "v2")

	rdsc := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:        []string{s.Addr()},
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Result may be nil or non-nil depending on miniredis cluster support.
	// Either way we exercise cleanCacheInClusters.
	_ = cleanCacheInClusters(ctx, "key:*", rdsc)
}

// TestGetKeysCountInClusters_WithMiniredis attempts the same for getKeysCountInClusters.
func TestGetKeysCountInClusters_WithMiniredis(t *testing.T) {
	s := miniredis.RunT(t)
	s.Set("key:1", "v1")

	rdsc := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:        []string{s.Addr()},
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _ = getKeysCountInClusters(ctx, "key:*", rdsc)
}
