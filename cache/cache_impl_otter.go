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

package cache

import (
	"time"

	"github.com/maypok86/otter/v2"
)

// NoExpiry, passed as Config.DefaultTTL, means the cache never expires
// entries: NewOtterCache constructs the underlying Otter cache without any
// expiry policy at all (Otter's native no-expiration mode), rather than
// simulating "permanent" with some very long duration.
//
// In this mode, Set's expire parameter has no effect on any key, silently
// (not an error): Otter's SetExpiresAfter is a no-op whenever the cache
// itself wasn't configured with an expiry policy, regardless of what's
// passed for an individual key. Otter's "which keys can have a TTL at all"
// is an all-or-nothing, cache-wide setting, not a per-key one — so a single
// cache can't mix "most keys are permanent" with "this one key really does
// expire". Construct a second OtterCache with a real DefaultTTL if both are
// needed side by side.
const NoExpiry time.Duration = -1

const (
	// defaultMaxSize is the default maximum number of entries.
	defaultMaxSize = 10_000
	// defaultTTL is the default TTL applied to a key when it's first
	// created via Set with expire <= 0.
	defaultTTL = time.Hour
)

// Config holds NewOtterCache's configuration. Zero-value fields fall back to
// the package defaults above.
type Config struct {
	// MaxSize is the maximum number of entries before the W-TinyLFU
	// admission policy starts evicting to make room for new keys — this
	// applies regardless of DefaultTTL; even a NoExpiry cache can still
	// evict entries under capacity pressure, just never due to TTL.
	// Zero-value: 10,000.
	MaxSize int

	// DefaultTTL is the TTL applied the first time a key is created via Set
	// with expire <= 0 (re-Set/Update on an already-existing key never
	// resets it - see the package doc). Zero-value: 1 hour.
	//
	// Set to NoExpiry to disable expiry entirely for this cache instance —
	// entries are then only ever removed by Delete/Clear/capacity eviction,
	// never by TTL, and Set's expire parameter stops having any effect for
	// every key (see NoExpiry's doc comment).
	DefaultTTL time.Duration
}

// resolve returns a copy of c with zero-value fields replaced by defaults.
func (c Config) resolve() Config {
	if c.MaxSize == 0 {
		c.MaxSize = defaultMaxSize
	}
	if c.DefaultTTL == 0 {
		c.DefaultTTL = defaultTTL
	}
	return c
}

// OtterCache is a high-performance in-process cache backed by the Otter library.
// It implements the ICache interface and supports TTL-based expiry.
type OtterCache struct {
	cache *otter.Cache[string, interface{}] // underlying Otter cache instance
}

// NewOtterCache creates a new OtterCache instance. cfg is optional: omit it
// for the package defaults (capacity 10,000, default TTL 1 hour), or pass
// one Config to customize capacity and/or the default TTL — including
// Config.DefaultTTL: NoExpiry for a cache that never expires entries.
func NewOtterCache(cfg ...Config) ICache {
	c := Config{}
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c = c.resolve()

	opts := &otter.Options[string, interface{}]{
		MaximumSize: c.MaxSize,
	}
	if c.DefaultTTL != NoExpiry {
		opts.ExpiryCalculator = otter.ExpiryCreating[string, interface{}](c.DefaultTTL)
	}
	// c.DefaultTTL == NoExpiry: leave ExpiryCalculator unset — Otter's own
	// native "no expiration policy configured" behavior.

	return &OtterCache{cache: otter.Must(opts)}
}

// Get retrieves the value associated with key from the cache.
// It returns the value and true if the key exists, or nil and false otherwise.
func (c *OtterCache) Get(key string) (interface{}, bool) {
	return c.cache.GetIfPresent(key)
}

// Set stores a value in the cache with an optional custom TTL.
//
// Behavior:
//   - If expire > 0: the entry's TTL is set to expire and overrides the default
//     expiry configured by ExpiryCalculator.
//   - If expire <= 0: no custom TTL is applied and the entry uses the default
//     expiry from ExpiryCalculator (currently 1 hour). Note that this differs
//     from some cache implementations where a zero TTL means "no expiration".
func (c *OtterCache) Set(key string, value interface{}, expire time.Duration) error {
	c.cache.Set(key, value)
	// Apply a custom TTL only when expire > 0; otherwise the default expiry from
	// ExpiryCalculator (1 hour) remains in effect.
	if expire > 0 {
		c.cache.SetExpiresAfter(key, expire)
	}
	return nil
}

// Update updates the value of the key in the cache without changing its TTL
func (c *OtterCache) Update(key string, value interface{}) error {
	c.cache.Set(key, value)
	return nil
}

// Keys returns a snapshot of all keys currently present in the cache.
func (c *OtterCache) Keys() []interface{} {
	keys := make([]interface{}, 0, c.cache.EstimatedSize())
	for key := range c.cache.Keys() {
		keys = append(keys, key)
	}
	return keys
}

// Values returns a snapshot of all values currently stored in the cache.
func (c *OtterCache) Values() []interface{} {
	values := make([]interface{}, 0, c.cache.EstimatedSize())
	for value := range c.cache.Values() {
		values = append(values, value)
	}
	return values
}

// Delete removes the entry identified by key from the cache.
// It always returns nil.
func (c *OtterCache) Delete(key string) error {
	c.cache.Invalidate(key)
	return nil
}

// Size returns an estimated number of entries currently held in the cache.
func (c *OtterCache) Size() int {
	return c.cache.EstimatedSize()
}

// Clear removes all entries from the cache.
// It always returns nil.
func (c *OtterCache) Clear() error {
	c.cache.InvalidateAll()
	return nil
}
