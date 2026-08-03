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

import "time"

// ICache defines a generic key-value cache interface with optional TTL support.
// All implementations must be safe for concurrent use.
type ICache interface {
	// Get retrieves the value for key. Returns the value and true if found, otherwise nil and false.
	Get(key string) (interface{}, bool)
	// Set stores value under key with the given TTL. A non-positive expire means use the default TTL.
	Set(key string, value interface{}, expire time.Duration) error
	// Update replaces the value of an existing key without altering its expiry.
	Update(key string, value interface{}) error
	// Keys returns all keys currently present in the cache.
	Keys() []interface{}
	// Values returns all values currently stored in the cache.
	Values() []interface{}
	// Delete removes the entry identified by key from the cache.
	Delete(key string) error
	// Size returns the number of entries currently held in the cache.
	Size() int
	// Clear removes all entries from the cache.
	Clear() error
}
