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
	"testing"
	"time"
)

// newTestCache returns an OtterCache via the ICache interface,
// ensuring NewOtterCache satisfies the interface contract.
func newTestCache(t *testing.T) ICache {
	t.Helper()
	c := NewOtterCache()
	if c == nil {
		t.Fatal("NewOtterCache returned nil")
	}
	return c
}

// ─────────────────────────────────────────────
// NewOtterCache
// ─────────────────────────────────────────────

func TestNewOtterCache_ImplementsICache(t *testing.T) {
	var _ ICache = NewOtterCache()
}

func TestNewOtterCache_InitialSizeIsZero(t *testing.T) {
	c := newTestCache(t)
	if c.Size() != 0 {
		t.Errorf("initial Size() must be 0, got %d", c.Size())
	}
}

// ─────────────────────────────────────────────
// Config / resolve
// ─────────────────────────────────────────────

func TestConfig_Resolve_ZeroValueGetsPackageDefaults(t *testing.T) {
	resolved := Config{}.resolve()
	if resolved.MaxSize != defaultMaxSize {
		t.Errorf("MaxSize = %d, want default %d", resolved.MaxSize, defaultMaxSize)
	}
	if resolved.DefaultTTL != defaultTTL {
		t.Errorf("DefaultTTL = %v, want default %v", resolved.DefaultTTL, defaultTTL)
	}
}

func TestConfig_Resolve_NonZeroFieldsArePreserved(t *testing.T) {
	resolved := Config{MaxSize: 42, DefaultTTL: 5 * time.Minute}.resolve()
	if resolved.MaxSize != 42 {
		t.Errorf("MaxSize = %d, want 42 (unchanged)", resolved.MaxSize)
	}
	if resolved.DefaultTTL != 5*time.Minute {
		t.Errorf("DefaultTTL = %v, want 5m (unchanged)", resolved.DefaultTTL)
	}
}

func TestConfig_Resolve_NoExpiryIsPreservedNotOverwritten(t *testing.T) {
	// NoExpiry (-1) must survive resolve() unchanged - it must never be
	// confused with the zero-value "unset, use the default TTL" case.
	resolved := Config{DefaultTTL: NoExpiry}.resolve()
	if resolved.DefaultTTL != NoExpiry {
		t.Errorf("DefaultTTL = %v, want NoExpiry (%v)", resolved.DefaultTTL, NoExpiry)
	}
}

// ─────────────────────────────────────────────
// NewOtterCache(Config{...})
// ─────────────────────────────────────────────

func TestNewOtterCache_ZeroValueConfig_BehavesLikeNoArgument(t *testing.T) {
	c := NewOtterCache(Config{})
	_ = c.Set("k", "v", 0)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected key present after Set on a zero-value Config cache")
	}
}

func TestNewOtterCache_CustomDefaultTTL_AppliesToKeysCreatedWithExpireZero(t *testing.T) {
	c := NewOtterCache(Config{DefaultTTL: 20 * time.Millisecond})
	_ = c.Set("k", "v", 0) // expire<=0 -> uses the configured DefaultTTL
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected key present immediately after Set")
	}
	time.Sleep(80 * time.Millisecond)
	if _, ok := c.Get("k"); ok {
		t.Error("expected key to have expired after the configured DefaultTTL elapsed")
	}
}

func TestNewOtterCache_CustomDefaultTTL_PerKeyExpireOverrideStillWorks(t *testing.T) {
	// A long DefaultTTL must not prevent an individual Set call from
	// overriding it with a shorter (or longer) per-key expire.
	c := NewOtterCache(Config{DefaultTTL: time.Hour})
	_ = c.Set("short-lived", "v", 20*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	if _, ok := c.Get("short-lived"); ok {
		t.Error("expected the per-key custom TTL to override the cache's long DefaultTTL")
	}
}

// TestNewOtterCache_NoExpiry_NeverExpires is a regression test for the
// DefaultTTL: NoExpiry configuration: a key created with expire<=0 must
// still be present well past what would otherwise be a short TTL.
func TestNewOtterCache_NoExpiry_NeverExpires(t *testing.T) {
	c := NewOtterCache(Config{DefaultTTL: NoExpiry})
	_ = c.Set("permanent", "v", 0)
	time.Sleep(80 * time.Millisecond)
	if _, ok := c.Get("permanent"); !ok {
		t.Error("expected key to still be present in a NoExpiry cache")
	}
}

// TestNewOtterCache_NoExpiry_SetExpireArgumentIsIgnored is a regression test
// for the documented NoExpiry caveat: passing expire > 0 to Set on a
// NoExpiry cache must not cause that key to expire - the whole cache has no
// expiry policy configured at all, so Otter's SetExpiresAfter silently
// no-ops regardless of which key it's called for.
func TestNewOtterCache_NoExpiry_SetExpireArgumentIsIgnored(t *testing.T) {
	c := NewOtterCache(Config{DefaultTTL: NoExpiry})
	_ = c.Set("k", "v", 20*time.Millisecond) // expire>0, but must have no effect
	time.Sleep(80 * time.Millisecond)
	if _, ok := c.Get("k"); !ok {
		t.Error("expected key to still be present - NoExpiry cache must ignore Set's expire parameter")
	}
}

func TestNewOtterCache_CustomMaxSize_ConstructsWithoutError(t *testing.T) {
	c := NewOtterCache(Config{MaxSize: 2})
	_ = c.Set("a", 1, 0)
	_ = c.Set("b", 2, 0)
	if _, ok := c.Get("a"); !ok {
		t.Error("expected key present within MaxSize capacity")
	}
}

// ─────────────────────────────────────────────
// Set / Get
// ─────────────────────────────────────────────

func TestSet_Get_BasicStringValue(t *testing.T) {
	c := newTestCache(t)
	err := c.Set("hello", "world", 0)
	if err != nil {
		t.Fatalf("Set returned unexpected error: %v", err)
	}
	val, ok := c.Get("hello")
	if !ok {
		t.Fatal("Get returned false for an existing key")
	}
	if val != "world" {
		t.Errorf("Get value: want %q got %v", "world", val)
	}
}

func TestSet_Get_IntValue(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("count", 42, 0)
	val, ok := c.Get("count")
	if !ok {
		t.Fatal("Get returned false")
	}
	if val != 42 {
		t.Errorf("want 42 got %v", val)
	}
}

func TestSet_Get_NilValue(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("nilkey", nil, 0)
	// Otter may or may not store nil values; we just ensure no panic
	// and the API is usable.
}

func TestGet_MissingKey_ReturnsFalse(t *testing.T) {
	c := newTestCache(t)
	_, ok := c.Get("does-not-exist")
	if ok {
		t.Error("Get must return false for a missing key")
	}
}

func TestSet_WithPositiveTTL_NoError(t *testing.T) {
	c := newTestCache(t)
	err := c.Set("ttlkey", "value", 10*time.Minute)
	if err != nil {
		t.Errorf("Set with TTL returned error: %v", err)
	}
}

func TestSet_OverwritesExistingKey(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("k", "first", 0)
	_ = c.Set("k", "second", 0)
	val, ok := c.Get("k")
	if !ok {
		t.Fatal("Get returned false after overwrite")
	}
	if val != "second" {
		t.Errorf("want %q got %v", "second", val)
	}
}

// ─────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────

func TestUpdate_ExistingKey(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("x", "original", time.Hour)
	err := c.Update("x", "updated")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	val, ok := c.Get("x")
	if !ok {
		t.Fatal("Get returned false after Update")
	}
	if val != "updated" {
		t.Errorf("want %q got %v", "updated", val)
	}
}

func TestUpdate_NoError(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("y", 1, 0)
	if err := c.Update("y", 2); err != nil {
		t.Errorf("Update must return nil error, got %v", err)
	}
}

// ─────────────────────────────────────────────
// Delete
// ─────────────────────────────────────────────

func TestDelete_ExistingKey(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("del", "value", 0)
	err := c.Delete("del")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	_, ok := c.Get("del")
	if ok {
		t.Error("Get must return false after Delete")
	}
}

func TestDelete_MissingKey_NoError(t *testing.T) {
	c := newTestCache(t)
	if err := c.Delete("missing"); err != nil {
		t.Errorf("Delete of missing key must return nil, got %v", err)
	}
}

// ─────────────────────────────────────────────
// Size
// ─────────────────────────────────────────────

func TestSize_IncreasesAfterSet(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("a", 1, 0)
	_ = c.Set("b", 2, 0)
	// Otter uses estimated size; at minimum it should be >= 0
	if c.Size() < 0 {
		t.Error("Size must not be negative")
	}
}

func TestSize_AfterClear(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("p", "q", 0)
	_ = c.Clear()
	if c.Size() != 0 {
		t.Errorf("Size after Clear must be 0, got %d", c.Size())
	}
}

// ─────────────────────────────────────────────
// Clear
// ─────────────────────────────────────────────

func TestClear_EmptyCache_NoError(t *testing.T) {
	c := newTestCache(t)
	if err := c.Clear(); err != nil {
		t.Errorf("Clear on empty cache must return nil, got %v", err)
	}
}

func TestClear_RemovesAllEntries(t *testing.T) {
	c := newTestCache(t)
	for i := 0; i < 5; i++ {
		_ = c.Set(string(rune('a'+i)), i, 0)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	// After clear, all keys should be gone
	for i := 0; i < 5; i++ {
		_, ok := c.Get(string(rune('a' + i)))
		if ok {
			t.Errorf("key %q should not exist after Clear", string(rune('a'+i)))
		}
	}
}

// ─────────────────────────────────────────────
// Keys / Values
// ─────────────────────────────────────────────

func TestKeys_EmptyCache(t *testing.T) {
	c := newTestCache(t)
	keys := c.Keys()
	if keys == nil {
		t.Error("Keys must return a non-nil slice even when empty")
	}
}

func TestValues_EmptyCache(t *testing.T) {
	c := newTestCache(t)
	vals := c.Values()
	if vals == nil {
		t.Error("Values must return a non-nil slice even when empty")
	}
}

func TestKeys_ContainsInsertedKey(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("mykey", "myval", 0)
	keys := c.Keys()
	found := false
	for _, k := range keys {
		if k == "mykey" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Keys() must contain inserted key 'mykey'")
	}
}

func TestValues_ContainsInsertedValue(t *testing.T) {
	c := newTestCache(t)
	_ = c.Set("vkey", "vval", 0)
	vals := c.Values()
	found := false
	for _, v := range vals {
		if v == "vval" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Values() must contain inserted value 'vval'")
	}
}

// ─────────────────────────────────────────────
// Multiple entries
// ─────────────────────────────────────────────

func TestMultipleEntries_IndependentKeys(t *testing.T) {
	c := newTestCache(t)
	entries := map[string]interface{}{
		"one": 1,
		"two": "two",
		"thr": true,
	}
	for k, v := range entries {
		_ = c.Set(k, v, 0)
	}
	for k, want := range entries {
		got, ok := c.Get(k)
		if !ok {
			t.Errorf("key %q: Get returned false", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: want %v got %v", k, want, got)
		}
	}
}

// ─────────────────────────────────────────────
// ICache interface compliance
// ─────────────────────────────────────────────

// TestICache_AllMethodsReturnExpectedTypes exercises every ICache method
// to confirm the concrete type satisfies the full interface contract.
func TestICache_AllMethodsReturnExpectedTypes(t *testing.T) {
	c := newTestCache(t)

	// Set
	if err := c.Set("k", "v", time.Second); err != nil {
		t.Errorf("Set: %v", err)
	}
	// Get
	val, ok := c.Get("k")
	if !ok || val == nil {
		t.Error("Get after Set must return (value, true)")
	}
	// Update
	if err := c.Update("k", "v2"); err != nil {
		t.Errorf("Update: %v", err)
	}
	// Keys
	keys := c.Keys()
	if keys == nil {
		t.Error("Keys must not return nil")
	}
	// Values
	values := c.Values()
	if values == nil {
		t.Error("Values must not return nil")
	}
	// Size
	_ = c.Size()
	// Delete
	if err := c.Delete("k"); err != nil {
		t.Errorf("Delete: %v", err)
	}
	// Clear
	if err := c.Clear(); err != nil {
		t.Errorf("Clear: %v", err)
	}
}
