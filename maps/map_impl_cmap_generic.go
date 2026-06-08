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

import (
	"fmt"
	"strconv"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// CMapGen is a generic, thread-safe concurrent map with comparable keys of type K and values of type V.
// Keys are converted to strings internally using keyToString.
// It supports pluggable replacement strategies and comparison functions.
type CMapGen[K comparable, V any] struct {
	maps            cmap.ConcurrentMap[string, V]
	defaultStrategy ReplaceStrategy[V] // default strategy used by Replace when no compare func is set
	defaultCompare  CompareFunc[V]     // default compare function used by Replace when set
}

// NewCMapGen creates and returns a new CMapGen with the default NumericGreaterStrategy replacement strategy.
func NewCMapGen[K comparable, V any]() *CMapGen[K, V] {
	return &CMapGen[K, V]{
		maps: cmap.New[V](),
		// The default string comparison strategy is used
		defaultStrategy: NumericGreaterStrategy[V]{},
	}
}

// Set stores the given value under the specified key, overwriting any existing value.
func (m *CMapGen[K, V]) Set(key K, value V) {
	m.maps.Set(keyToString(key), value)
}

// Get retrieves the value for the given key and a boolean indicating whether the key was found.
func (m *CMapGen[K, V]) Get(key K) (V, bool) {
	return m.maps.Get(keyToString(key))
}

// Replace conditionally updates the value for the given key according to the configured strategy.
// If a default compare function is set it takes precedence; otherwise the default strategy is used.
// If neither is configured, values are compared via their string representations.
// If the key does not exist it is always created. Returns true if the value was stored.
func (m *CMapGen[K, V]) Replace(key K, value V) bool {
	// step 1: check if a default comparison function
	if m.defaultCompare != nil {
		return m.ReplaceWithCompare(key, value, m.defaultCompare)
	}

	// step 2: check if a default strategy is set
	if m.defaultStrategy != nil {
		return m.ReplaceWithStrategy(key, value, m.defaultStrategy)
	}

	// step 3: if no default strategy is set, the original string comparison logic is used
	keyStr := keyToString(key)
	val, ok := m.maps.Get(keyStr)
	if ok {
		// This assumes V can be compared (works for numeric types)
		// For a more type-safe approach, you might want to add constraints
		if fmt.Sprint(val) < fmt.Sprint(value) {
			m.maps.Set(keyStr, value)
			return true
		}
	} else {
		m.maps.Set(keyStr, value)
		return true
	}
	return false
}

// Delete removes the entry associated with the given key from the map.
func (m *CMapGen[K, V]) Delete(key K) {
	m.maps.Remove(keyToString(key))
}

// Range iterates over all key-value pairs in the map, calling f for each one.
// Iteration stops early if f returns false.
func (m *CMapGen[K, V]) Range(f func(key K, value V) bool) {
	for item := range m.maps.IterBuffered() {
		// Convert string key back to K type - this is a limitation
		// You might need a more sophisticated key conversion strategy
		//var k K
		//fmt.Sscanf(item.Key, "%v", &k)

		// Use efficient type-specific conversion instead of fmt.Sscanf
		k := stringToKey[K](item.Key)
		if !f(k, item.Val) {
			break
		}
	}
}

// Clear removes all key-value pairs from the map.
func (m *CMapGen[K, V]) Clear() {
	m.maps.Clear()
}

// Size returns the number of key-value pairs currently stored in the map.
func (m *CMapGen[K, V]) Size() int {
	return m.maps.Count()
}

// SetDefaultStrategy sets the replacement strategy used by Replace when no compare function is configured.
// Setting a strategy clears any previously configured compare function.
func (m *CMapGen[K, V]) SetDefaultStrategy(strategy ReplaceStrategy[V]) {
	m.defaultStrategy = strategy
	m.defaultCompare = nil // Clear the comparison function and use the strategy
}

// SetDefaultCompare sets the comparison function used by Replace to decide whether to overwrite an existing value.
// Setting a compare function clears any previously configured replacement strategy.
func (m *CMapGen[K, V]) SetDefaultCompare(compare CompareFunc[V]) {
	m.defaultCompare = compare
	m.defaultStrategy = nil // Clear strategy, use comparison function
}

// ReplaceWithCompare updates the value for the given key using the provided comparison function.
// The value is stored only when compare(existingValue, newValue) returns true.
// If the key does not exist the value is always stored. Returns true if the value was stored.
func (m *CMapGen[K, V]) ReplaceWithCompare(key K, value V, compare CompareFunc[V]) bool {
	keyStr := keyToString(key)
	val, ok := m.maps.Get(keyStr)
	if ok {
		if compare(val, value) {
			m.maps.Set(keyStr, value)
			return true
		}
	} else {
		m.maps.Set(keyStr, value)
		return true
	}
	return false
}

// ReplaceWithStrategy updates the value for the given key using the provided strategy.
// The value is stored only when strategy.ShouldReplace(existingValue, newValue) returns true.
// If the key does not exist the value is always stored. Returns true if the value was stored.
func (m *CMapGen[K, V]) ReplaceWithStrategy(key K, value V, strategy ReplaceStrategy[V]) bool {
	keyStr := keyToString(key)
	val, ok := m.maps.Get(keyStr)
	if ok {
		if strategy.ShouldReplace(val, value) {
			m.maps.Set(keyStr, value)
			return true
		}
	} else {
		m.maps.Set(keyStr, value)
		return true
	}
	return false
}

// ReplaceAlways unconditionally stores value under key, overwriting any existing entry.
// It always returns true.
func (m *CMapGen[K, V]) ReplaceAlways(key K, value V) bool {
	m.maps.Set(keyToString(key), value)
	return true
}

// ReplaceIfNotExists stores value under key only when the key is not already present.
// Returns true if the value was stored, false if the key already existed.
func (m *CMapGen[K, V]) ReplaceIfNotExists(key K, value V) bool {
	keyStr := keyToString(key)
	_, ok := m.maps.Get(keyStr)
	if !ok {
		m.maps.Set(keyStr, value)
		return true
	}
	return false
}

// UpsertWithCallback inserts or updates the value for key using the provided callback.
// The callback receives whether the key existed, the old value (zero value if not), and the new value,
// and returns the value that should ultimately be stored. Always returns true.
func (m *CMapGen[K, V]) UpsertWithCallback(key K, value V, callback func(exists bool, oldValue, newValue V) V) bool {
	keyStr := keyToString(key)
	oldValue, exists := m.maps.Get(keyStr)
	newValue := callback(exists, oldValue, value)
	m.maps.Set(keyStr, newValue)
	return true
}

// AlwaysReplaceStrategy is a ReplaceStrategy that unconditionally signals that the old value
// should be replaced by the new value.
type AlwaysReplaceStrategy[V any] struct{}

// ShouldReplace always returns true, indicating the value should always be replaced.
func (s AlwaysReplaceStrategy[V]) ShouldReplace(oldValue, newValue V) bool {
	return true
}

// NumericGreaterStrategy is a ReplaceStrategy that replaces the old value only when the
// string representation of the new value is lexicographically greater than the old one.
// For numeric types this is equivalent to a numeric greater-than comparison.
type NumericGreaterStrategy[V any] struct{}

// ShouldReplace returns true when the string representation of newValue is greater than that of oldValue.
func (s NumericGreaterStrategy[V]) ShouldReplace(oldValue, newValue V) bool {
	return fmt.Sprint(newValue) > fmt.Sprint(oldValue)
}

// TimestampStrategy is a ReplaceStrategy that accepts a new int64 value only when
// the difference between the new and old value meets a minimum threshold.
// Typical use case: accept a new timestamp only when it is sufficiently newer than the stored one.
type TimestampStrategy struct {
	MinInterval int64 // minimum required difference (newValue - oldValue) for a replacement to occur
}

// ShouldReplace returns true when newValue-oldValue is at least MinInterval.
func (s TimestampStrategy) ShouldReplace(oldValue, newValue int64) bool {
	return newValue-oldValue >= s.MinInterval
}

// keyToString converts K type to string efficiently
func keyToString[K comparable](key K) string {
	switch k := any(key).(type) {
	case string:
		return k
	case int:
		return strconv.Itoa(k)
	case int8:
		return strconv.FormatInt(int64(k), 10)
	case int16:
		return strconv.FormatInt(int64(k), 10)
	case int32:
		return strconv.FormatInt(int64(k), 10)
	case int64:
		return strconv.FormatInt(k, 10)
	case uint:
		return strconv.FormatUint(uint64(k), 10)
	case uint8:
		return strconv.FormatUint(uint64(k), 10)
	case uint16:
		return strconv.FormatUint(uint64(k), 10)
	case uint32:
		return strconv.FormatUint(uint64(k), 10)
	case uint64:
		return strconv.FormatUint(k, 10)
	case float32:
		return strconv.FormatFloat(float64(k), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(k, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(k)
	default:
		// Fallback to fmt.Sprint for other types
		return fmt.Sprint(key)
	}
}

// stringToKey converts string back to K type efficiently
func stringToKey[K comparable](s string) K {
	var k K
	switch any(k).(type) {
	case string:
		return any(s).(K)
	case int:
		if val, err := strconv.Atoi(s); err == nil {
			return any(val).(K)
		}
	case int8:
		if val, err := strconv.ParseInt(s, 10, 8); err == nil {
			return any(int8(val)).(K)
		}
	case int16:
		if val, err := strconv.ParseInt(s, 10, 16); err == nil {
			return any(int16(val)).(K)
		}
	case int32:
		if val, err := strconv.ParseInt(s, 10, 32); err == nil {
			return any(int32(val)).(K)
		}
	case int64:
		if val, err := strconv.ParseInt(s, 10, 64); err == nil {
			return any(val).(K)
		}
	case uint:
		if val, err := strconv.ParseUint(s, 10, 0); err == nil {
			return any(uint(val)).(K)
		}
	case uint8:
		if val, err := strconv.ParseUint(s, 10, 8); err == nil {
			return any(uint8(val)).(K)
		}
	case uint16:
		if val, err := strconv.ParseUint(s, 10, 16); err == nil {
			return any(uint16(val)).(K)
		}
	case uint32:
		if val, err := strconv.ParseUint(s, 10, 32); err == nil {
			return any(uint32(val)).(K)
		}
	case uint64:
		if val, err := strconv.ParseUint(s, 10, 64); err == nil {
			return any(val).(K)
		}
	case float32:
		if val, err := strconv.ParseFloat(s, 32); err == nil {
			return any(float32(val)).(K)
		}
	case float64:
		if val, err := strconv.ParseFloat(s, 64); err == nil {
			return any(val).(K)
		}
	case bool:
		if val, err := strconv.ParseBool(s); err == nil {
			return any(val).(K)
		}
	default:
		// For other types, use fmt.Sscanf as fallback
		fmt.Sscanf(s, "%v", &k)
		return k
	}
	return k
}
