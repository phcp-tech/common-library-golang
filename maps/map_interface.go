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

package maps

// IMap defines a generic, thread-safe key-value map interface with pluggable replacement
// strategies. Implementations must support conditional replacement via Replace as well as
// configurable strategies and comparison functions.
type IMap[K comparable, V any] interface {
	// Set stores value under key, overwriting any existing entry.
	Set(key K, value V)

	// Get retrieves the value associated with key and reports whether the key was found.
	Get(key K) (V, bool)

	// Replace conditionally updates the value for key according to the configured strategy
	// or comparison function. Returns true if the value was stored.
	Replace(key K, value V) bool

	// Delete removes the entry associated with key from the map.
	Delete(key K)

	// Range iterates over all key-value pairs, calling f for each one.
	// Iteration stops early if f returns false.
	Range(f func(key K, value V) bool)

	// Size returns the number of key-value pairs currently stored in the map.
	Size() int

	// Clear removes all key-value pairs from the map.
	Clear()

	// SetDefaultStrategy sets the replacement strategy used by Replace when no compare
	// function is configured.
	SetDefaultStrategy(strategy ReplaceStrategy[V])

	// SetDefaultCompare sets the comparison function used by Replace to decide whether
	// to overwrite an existing value.
	SetDefaultCompare(compare CompareFunc[V])
}

// ReplaceStrategy defines the interface for a pluggable replacement strategy used by map implementations.
type ReplaceStrategy[V any] interface {
	// ShouldReplace returns true when the old value should be overwritten by the new value.
	ShouldReplace(oldValue, newValue V) bool
}

// CompareFunc is a function type used to compare an existing (old) map value against a candidate
// (new) value. It returns true when the new value should overwrite the old one.
type CompareFunc[V any] func(oldValue, newValue V) bool
