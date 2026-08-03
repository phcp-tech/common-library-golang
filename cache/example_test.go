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

// package cache_test demonstrates the public API from a caller's perspective.
package cache_test

import (
	"fmt"
	"time"

	"github.com/phcp-tech/common-library-golang/cache"
)

// ExampleNewOtterCache shows how to create an in-process cache and perform a
// basic Set / Get round trip. A zero (or negative) expire means "use the
// cache's default TTL" rather than "never expire" — see ExampleOtterCache_Set.
func ExampleNewOtterCache() {
	c := cache.NewOtterCache()
	_ = c.Set("EURUSD", 10500, 0)

	val, ok := c.Get("EURUSD")
	fmt.Println(val, ok)
	// Output:
	// 10500 true
}

// ExampleNewOtterCache_config shows customizing capacity and the default
// TTL. Config is optional — omit it entirely (as in ExampleNewOtterCache)
// for capacity 10,000 and a 1-hour default TTL.
func ExampleNewOtterCache_config() {
	c := cache.NewOtterCache(cache.Config{
		MaxSize:    500,
		DefaultTTL: 10 * time.Minute,
	})
	_ = c.Set("EURUSD", 10500, 0) // expire<=0 -> uses this cache's 10-minute default

	val, ok := c.Get("EURUSD")
	fmt.Println(val, ok)
	// Output:
	// 10500 true
}

// ExampleNewOtterCache_noExpiry shows building a cache that never expires
// entries. In this mode Set's expire parameter has no effect on any key —
// see cache.NoExpiry's doc comment for why a single cache can't mix
// permanent keys with keys that really do expire.
func ExampleNewOtterCache_noExpiry() {
	c := cache.NewOtterCache(cache.Config{DefaultTTL: cache.NoExpiry})

	_ = c.Set("config:feature-flags", "enabled", 0)     // never expires
	_ = c.Set("also-permanent", "value", 5*time.Minute) // expire is ignored: also never expires

	val, ok := c.Get("config:feature-flags")
	fmt.Println(val, ok)
	// Output:
	// enabled true
}

// ExampleOtterCache_Get shows Get on a key that was never set: it returns
// (nil, false) rather than panicking or returning a zero value of some
// concrete type — the cache stores values as interface{}.
func ExampleOtterCache_Get() {
	c := cache.NewOtterCache()
	val, ok := c.Get("missing")
	fmt.Println(val, ok)
	// Output:
	// <nil> false
}

// ExampleOtterCache_Set shows the default expire behavior: 0 (or any
// non-positive duration) does not mean "never expires" — it means "use
// whatever default TTL this cache was constructed with" (NewOtterCache's
// default is 1 hour). Pass a positive duration for a custom, per-entry TTL.
func ExampleOtterCache_Set() {
	c := cache.NewOtterCache()
	_ = c.Set("session-token", "abc123", 5*time.Minute)

	val, ok := c.Get("session-token")
	fmt.Println(val, ok)
	// Output:
	// abc123 true
}

// ExampleOtterCache_Update shows replacing an existing key's value without
// touching its expiry — unlike Set, Update never resets the TTL clock.
func ExampleOtterCache_Update() {
	c := cache.NewOtterCache()
	_ = c.Set("counter", 1, 0)
	_ = c.Update("counter", 2)

	val, _ := c.Get("counter")
	fmt.Println(val)
	// Output:
	// 2
}

// ExampleOtterCache_Delete shows removing an entry; Get on the deleted key
// afterwards reports a miss.
func ExampleOtterCache_Delete() {
	c := cache.NewOtterCache()
	_ = c.Set("temp", "value", 0)
	_ = c.Delete("temp")

	_, ok := c.Get("temp")
	fmt.Println(ok)
	// Output:
	// false
}

// ExampleOtterCache_Size shows Size reflecting the number of entries
// currently held.
func ExampleOtterCache_Size() {
	c := cache.NewOtterCache()
	_ = c.Set("a", 1, 0)
	_ = c.Set("b", 2, 0)

	fmt.Println(c.Size())
	// Output:
	// 2
}

// ExampleOtterCache_Clear shows removing every entry at once.
func ExampleOtterCache_Clear() {
	c := cache.NewOtterCache()
	_ = c.Set("a", 1, 0)
	_ = c.Set("b", 2, 0)
	_ = c.Clear()

	fmt.Println(c.Size())
	// Output:
	// 0
}

// ExampleOtterCache_Keys shows listing every key currently stored. Only one
// key is inserted here because Keys' traversal order over multiple entries
// is not guaranteed (like Go's own map iteration) — sort the result first if
// you need a stable order for more than one key.
func ExampleOtterCache_Keys() {
	c := cache.NewOtterCache()
	_ = c.Set("only-key", "value", 0)

	fmt.Println(c.Keys())
	// Output:
	// [only-key]
}

// ExampleOtterCache_Values shows listing every value currently stored — see
// ExampleOtterCache_Keys about traversal order with more than one entry.
func ExampleOtterCache_Values() {
	c := cache.NewOtterCache()
	_ = c.Set("only-key", 42, 0)

	fmt.Println(c.Values())
	// Output:
	// [42]
}
