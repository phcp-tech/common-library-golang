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

// package maps_test demonstrates the public API from a caller's perspective.
package maps_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/maps"
)

// -----------------------------------------------------------------------
// CMap — fixed string→int64 map
// -----------------------------------------------------------------------

// ExampleNewCMap shows how to create a thread-safe string→int64 map and
// perform basic Set / Get / Replace / Delete operations.
func ExampleNewCMap() {
	m := maps.NewCMap()
	m.Set("EURUSD", 10500)
	m.Set("GBPUSD", 12800)

	val, ok := m.Get("EURUSD")
	fmt.Println(val, ok)

	// Replace only stores the new value when it is greater than the stored one.
	fmt.Println(m.Replace("EURUSD", 10600)) // true  — 10600 > 10500
	fmt.Println(m.Replace("EURUSD", 10400)) // false — 10400 < 10600

	m.Delete("GBPUSD")
	fmt.Println(m.Size())
	// Output:
	// 10500 true
	// true
	// false
	// 1
}

// -----------------------------------------------------------------------
// CMapGen — generic map with pluggable replacement strategy
// -----------------------------------------------------------------------

// ExampleNewCMapGen shows how to create a generic concurrent map with string
// keys and int64 values, and configure a default comparison function so that
// Replace only updates when the new value is strictly greater.
func ExampleNewCMapGen() {
	m := maps.NewCMapGen[string, int64]()
	m.SetDefaultCompare(func(old, new int64) bool { return new > old })

	m.Set("USDJPY", 15000)
	fmt.Println(m.Replace("USDJPY", 15100)) // true
	fmt.Println(m.Replace("USDJPY", 14900)) // false
	fmt.Println(m.Size())
	// Output:
	// true
	// false
	// 1
}

// ExampleCMapGen_ReplaceWithStrategy shows how to supply a strategy ad-hoc
// without modifying the map's default configuration.
func ExampleCMapGen_ReplaceWithStrategy() {
	m := maps.NewCMapGen[string, int64]()
	m.Set("AUDUSD", 7000)

	// NumericGreaterStrategy replaces only when the new value is greater.
	replaced := m.ReplaceWithStrategy("AUDUSD", 7200, maps.NumericGreaterStrategy[int64]{})
	fmt.Println(replaced)

	val, _ := m.Get("AUDUSD")
	fmt.Println(val)
	// Output:
	// true
	// 7200
}

// ExampleCMapGen_ReplaceAlways shows unconditional overwrite regardless of value.
func ExampleCMapGen_ReplaceAlways() {
	m := maps.NewCMapGen[string, int64]()
	m.Set("NZDUSD", 6500)

	m.ReplaceAlways("NZDUSD", 6000) // always stored, even though 6000 < 6500
	val, _ := m.Get("NZDUSD")
	fmt.Println(val)
	// Output:
	// 6000
}

// ExampleCMapGen_ReplaceIfNotExists shows set-once semantics: value is stored
// only when the key is not already present.
func ExampleCMapGen_ReplaceIfNotExists() {
	m := maps.NewCMapGen[string, int64]()

	fmt.Println(m.ReplaceIfNotExists("CADUSD", 7500)) // true  — key absent
	fmt.Println(m.ReplaceIfNotExists("CADUSD", 8000)) // false — key exists

	val, _ := m.Get("CADUSD")
	fmt.Println(val)
	// Output:
	// true
	// false
	// 7500
}

// ExampleCMapGen_UpsertWithCallback shows how to perform an atomic
// read-modify-write using a callback. The callback receives whether the key
// existed, the old value, and the candidate new value, and returns the value
// to store.
func ExampleCMapGen_UpsertWithCallback() {
	m := maps.NewCMapGen[string, int64]()
	m.Set("CHFUSD", 1000)

	// Keep the larger of the two values.
	m.UpsertWithCallback("CHFUSD", 1200, func(exists bool, old, new int64) int64 {
		if !exists || new > old {
			return new
		}
		return old
	})
	val, _ := m.Get("CHFUSD")
	fmt.Println(val)
	// Output:
	// 1200
}

// ExampleCMapGen_Range shows how to iterate over all entries. Returning false
// from the callback stops iteration early.
func ExampleCMapGen_Range() {
	m := maps.NewCMapGen[string, int64]()
	m.Set("A", 1)

	m.Range(func(key string, val int64) bool {
		fmt.Println(key, val)
		return true
	})
	// Output:
	// A 1
}

// -----------------------------------------------------------------------
// TimestampStrategy
// -----------------------------------------------------------------------

// ExampleTimestampStrategy shows how to accept a new value only when it
// exceeds the stored value by at least a minimum interval.
func ExampleTimestampStrategy() {
	m := maps.NewCMapGen[string, int64]()
	m.Set("tick", 1000)

	s := maps.TimestampStrategy{MinInterval: 10}

	fmt.Println(m.ReplaceWithStrategy("tick", 1015, s)) // true  — diff 15 >= 10
	fmt.Println(m.ReplaceWithStrategy("tick", 1020, s)) // true  — diff 5 < 10? no, 1020-1015=5
	val, _ := m.Get("tick")
	fmt.Println(val)
	// Output:
	// true
	// false
	// 1015
}

