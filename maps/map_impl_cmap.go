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
	cmap "github.com/orcaman/concurrent-map/v2"
)

// CMap is a thread-safe map with string keys and int64 values, backed by a concurrent map implementation.
// It satisfies the IMap[string, int64] interface.
type CMap struct {
	maps cmap.ConcurrentMap[string, int64]
}

// NewCMap creates and returns a new CMap instance that satisfies the IMap[string, int64] interface.
func NewCMap() IMap[string, int64] {
	return &CMap{
		maps: cmap.New[int64](),
	}
}

// Set stores the given value under the specified key, overwriting any existing value.
func (m *CMap) Set(key string, value int64) {
	m.maps.Set(key, value)
}

// Get retrieves the value for the given key and a boolean indicating whether the key was found.
func (m *CMap) Get(key string) (int64, bool) {
	return m.maps.Get(key)
}

// Replace updates the value for the given key only when the new value is greater than the existing value.
// If the key does not exist, it is created with the new value. Returns true if the value was stored.
func (m *CMap) Replace(key string, value int64) bool {
	// // way 1: replace value anytime, so there is no quotes delay and false alarms
	//dao.tickMap.Set(key, value)
	//return true

	// // way 2: replace value only when the new value is greater than the old value
	// Upsert only return int64, not bool
	// dao.tickMap.Upsert(key, value, existTickCB)
	// return true

	// way 3: query the value first, then compare the value, last replace it
	// if the same value, do not replace it, call Get() method only once at least, but lock twice if replace it.
	val, ok := m.maps.Get(key)
	if ok {
		if val < value {
			m.maps.Set(key, value)
			return true
		}
	} else {
		m.maps.Set(key, value)
		return true
	}
	return false
}

// Delete removes the entry associated with the given key from the map.
func (m *CMap) Delete(key string) {
	m.maps.Remove(key)
}

// Range iterates over all key-value pairs in the map, calling f for each one.
// Iteration stops early if f returns false.
func (m *CMap) Range(f func(key string, value int64) bool) {
	for item := range m.maps.IterBuffered() {
		if !f(item.Key, item.Val) {
			break
		}
	}
}

// Clear removes all key-value pairs from the map.
func (m *CMap) Clear() {
	m.maps.Clear()
}

// Size returns the number of key-value pairs currently stored in the map.
func (m *CMap) Size() int {
	return m.maps.Count()
}

// SetDefaultStrategy is a no-op for CMap; the replacement strategy is fixed (greater-value wins)
// and cannot be customised on this implementation.
func (m *CMap) SetDefaultStrategy(strategy ReplaceStrategy[int64]) {
	// CMap does not support custom replace strategy, so do nothing here
}

// SetDefaultCompare is a no-op for CMap; custom comparison functions are not supported
// on this implementation.
func (m *CMap) SetDefaultCompare(compare CompareFunc[int64]) {
	// CMap does not support custom compare function, so do nothing here
}

// need return: int64, bool
func existTickCB_camp(exists bool, valueInMap int64, newValue int64) int64 {
	if exists {
		// only new value is greater than value in map, then update
		if newValue > valueInMap {
			return newValue
		} else {
			return valueInMap
		}
	} else {
		return newValue
	}
}
