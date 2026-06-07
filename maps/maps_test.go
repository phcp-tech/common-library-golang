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

package maps

import (
	"sync"
	"testing"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mockAlwaysStrategy is a ReplaceStrategy that always signals replacement.
type mockAlwaysStrategy[V any] struct{ called bool }

func (s *mockAlwaysStrategy[V]) ShouldReplace(_, _ V) bool {
	s.called = true
	return true
}

// neverStrategy is a ReplaceStrategy that never signals replacement.
type neverStrategy[V any] struct{}

func (s neverStrategy[V]) ShouldReplace(_, _ V) bool { return false }

// newStringCmap creates a fresh concurrent map for the fallback-path tests.
func newStringCmap() cmap.ConcurrentMap[string, string] {
	return cmap.New[string]()
}

// ---------------------------------------------------------------------------
// CMap tests  (IMap[string, int64])
// ---------------------------------------------------------------------------

func TestCMap_SetAndGet(t *testing.T) {
	m := NewCMap()
	m.Set("key1", 100)

	val, ok := m.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
}

func TestCMap_GetMissingKey(t *testing.T) {
	m := NewCMap()
	_, ok := m.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestCMap_SetOverwrite(t *testing.T) {
	m := NewCMap()
	m.Set("k", 1)
	m.Set("k", 999)

	val, ok := m.Get("k")
	if !ok || val != 999 {
		t.Fatalf("expected 999, got %d (ok=%v)", val, ok)
	}
}

func TestCMap_Delete(t *testing.T) {
	m := NewCMap()
	m.Set("key", 42)
	m.Delete("key")

	_, ok := m.Get("key")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestCMap_DeleteNonExistent(t *testing.T) {
	m := NewCMap()
	m.Delete("nonexistent") // must not panic
}

func TestCMap_Size_Empty(t *testing.T) {
	m := NewCMap()
	if m.Size() != 0 {
		t.Fatalf("expected size 0, got %d", m.Size())
	}
}

func TestCMap_Size_AfterOperations(t *testing.T) {
	m := NewCMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	if m.Size() != 3 {
		t.Fatalf("expected size 3, got %d", m.Size())
	}

	m.Delete("b")
	if m.Size() != 2 {
		t.Fatalf("expected size 2 after delete, got %d", m.Size())
	}
}

func TestCMap_Clear(t *testing.T) {
	m := NewCMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Clear()

	if m.Size() != 0 {
		t.Fatalf("expected size 0 after Clear, got %d", m.Size())
	}
	_, ok := m.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be cleared")
	}
}

func TestCMap_Range_AllEntries(t *testing.T) {
	m := NewCMap()
	m.Set("x", 10)
	m.Set("y", 20)
	m.Set("z", 30)

	collected := make(map[string]int64)
	m.Range(func(key string, value int64) bool {
		collected[key] = value
		return true
	})

	if len(collected) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(collected))
	}
	for k, want := range map[string]int64{"x": 10, "y": 20, "z": 30} {
		if collected[k] != want {
			t.Fatalf("key %q: expected %d, got %d", k, want, collected[k])
		}
	}
}

func TestCMap_RangeEarlyStop(t *testing.T) {
	m := NewCMap()
	for i := int64(0); i < 10; i++ {
		m.Set(string(rune('a'+i)), i)
	}

	count := 0
	m.Range(func(key string, value int64) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Fatalf("expected Range to stop at 3, got %d", count)
	}
}

func TestCMap_Replace_GreaterValue(t *testing.T) {
	m := NewCMap()
	m.Set("k", 50)

	if !m.Replace("k", 100) {
		t.Fatal("expected Replace true when new > old")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
}

func TestCMap_Replace_SmallerValue(t *testing.T) {
	m := NewCMap()
	m.Set("k", 100)

	if m.Replace("k", 50) {
		t.Fatal("expected Replace false when new < old")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("value should remain 100, got %d", val)
	}
}

func TestCMap_Replace_EqualValue(t *testing.T) {
	m := NewCMap()
	m.Set("k", 100)

	if m.Replace("k", 100) {
		t.Fatal("expected Replace false when new == old")
	}
}

func TestCMap_Replace_KeyNotExist(t *testing.T) {
	m := NewCMap()

	if !m.Replace("newkey", 42) {
		t.Fatal("expected Replace true for new key")
	}
	val, ok := m.Get("newkey")
	if !ok || val != 42 {
		t.Fatalf("expected 42, got %d (ok=%v)", val, ok)
	}
}

// SetDefaultStrategy and SetDefaultCompare are no-ops for CMap – just verify they don't panic.
func TestCMap_SetDefaultStrategy_NoOp(t *testing.T) {
	m := NewCMap()
	m.SetDefaultStrategy(neverStrategy[int64]{})
	m.Set("k", 1)
	if !m.Replace("k", 10) {
		t.Fatal("CMap Replace should still follow built-in logic regardless of SetDefaultStrategy")
	}
}

func TestCMap_SetDefaultCompare_NoOp(t *testing.T) {
	m := NewCMap()
	m.SetDefaultCompare(func(old, new int64) bool { return false })
	m.Set("k", 1)
	if !m.Replace("k", 10) {
		t.Fatal("CMap Replace should still follow built-in logic regardless of SetDefaultCompare")
	}
}

// Concurrent safety
func TestCMap_ConcurrentSetGet(t *testing.T) {
	m := NewCMap()
	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				m.Set("key", int64(id*iterations+i))
				m.Get("key")
			}
		}(g)
	}
	wg.Wait()

	if m.Size() != 1 {
		t.Fatalf("expected size 1, got %d", m.Size())
	}
}

func TestCMap_ConcurrentReplace(t *testing.T) {
	m := NewCMap()
	m.Set("stock", 0)

	var wg sync.WaitGroup
	for g := 1; g <= 20; g++ {
		wg.Add(1)
		go func(v int64) {
			defer wg.Done()
			m.Replace("stock", v)
		}(int64(g * 100))
	}
	wg.Wait()

	val, ok := m.Get("stock")
	if !ok {
		t.Fatal("key 'stock' must still exist")
	}
	if val <= 0 {
		t.Fatalf("expected positive value, got %d", val)
	}
}

// ---------------------------------------------------------------------------
// CMapGen tests (generic map)
// ---------------------------------------------------------------------------

func TestCMapGen_SetAndGet_StringKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("hello", 999)

	val, ok := m.Get("hello")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != 999 {
		t.Fatalf("expected 999, got %d", val)
	}
}

func TestCMapGen_SetAndGet_IntKey(t *testing.T) {
	m := NewCMapGen[int, string]()
	m.Set(42, "forty-two")

	val, ok := m.Get(42)
	if !ok {
		t.Fatal("expected key 42 to exist")
	}
	if val != "forty-two" {
		t.Fatalf("expected 'forty-two', got %q", val)
	}
}

func TestCMapGen_GetMissing(t *testing.T) {
	m := NewCMapGen[string, int64]()
	_, ok := m.Get("missing")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestCMapGen_Delete(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("del", 1)
	m.Delete("del")

	_, ok := m.Get("del")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestCMapGen_Size(t *testing.T) {
	m := NewCMapGen[string, int64]()
	if m.Size() != 0 {
		t.Fatalf("expected size 0, got %d", m.Size())
	}
	m.Set("a", 1)
	m.Set("b", 2)
	if m.Size() != 2 {
		t.Fatalf("expected size 2, got %d", m.Size())
	}
}

func TestCMapGen_Clear(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Clear()
	if m.Size() != 0 {
		t.Fatalf("expected size 0 after Clear, got %d", m.Size())
	}
}

func TestCMapGen_Range(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("p", 10)
	m.Set("q", 20)

	sum := int64(0)
	m.Range(func(key string, value int64) bool {
		sum += value
		return true
	})
	if sum != 30 {
		t.Fatalf("expected sum 30, got %d", sum)
	}
}

func TestCMapGen_RangeEarlyStop(t *testing.T) {
	m := NewCMapGen[string, int64]()
	for i := int64(0); i < 10; i++ {
		m.Set(string(rune('a'+i)), i)
	}
	count := 0
	m.Range(func(_ string, _ int64) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Fatalf("expected early stop at 2, got %d", count)
	}
}

func TestCMapGen_Replace_WithDefaultCompare_Greater(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.SetDefaultCompare(func(old, new int64) bool { return new > old })
	m.Set("k", 50)

	if !m.Replace("k", 100) {
		t.Fatal("expected Replace true when new > old")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
}

func TestCMapGen_Replace_WithDefaultCompare_Lesser(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.SetDefaultCompare(func(old, new int64) bool { return new > old })
	m.Set("k", 100)

	if m.Replace("k", 10) {
		t.Fatal("expected Replace false when new < old")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("value should remain 100, got %d", val)
	}
}

func TestCMapGen_Replace_WithDefaultStrategy(t *testing.T) {
	m := NewCMapGen[string, int64]()
	ms := &mockAlwaysStrategy[int64]{}
	m.SetDefaultStrategy(ms)
	m.Set("k", 10)

	if !m.Replace("k", 99) {
		t.Fatal("expected Replace true with AlwaysReplace strategy")
	}
	if !ms.called {
		t.Fatal("expected ShouldReplace to be called")
	}
	val, _ := m.Get("k")
	if val != 99 {
		t.Fatalf("expected 99, got %d", val)
	}
}

func TestCMapGen_Replace_Fallback_StringComparison(t *testing.T) {
	// Create a CMapGen with no strategy and no compare set,
	// and bypass NewCMapGen (which installs NumericGreaterStrategy).
	m := &CMapGen[string, string]{
		maps: newStringCmap(),
	}
	m.Set("k", "abc")

	if !m.Replace("k", "xyz") { // "xyz" > "abc"
		t.Fatal("expected Replace true for 'xyz' > 'abc'")
	}
	val, _ := m.Get("k")
	if val != "xyz" {
		t.Fatalf("expected 'xyz', got %q", val)
	}

	if m.Replace("k", "aaa") { // "aaa" < "xyz"
		t.Fatal("expected Replace false for 'aaa' < 'xyz'")
	}
}

func TestCMapGen_SetDefaultCompare_ClearsStrategy(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.SetDefaultStrategy(AlwaysReplaceStrategy[int64]{})
	m.SetDefaultCompare(func(old, new int64) bool { return false })
	m.Set("k", 5)

	if m.Replace("k", 99) {
		t.Fatal("expected Replace false because compare always returns false")
	}
	if m.defaultStrategy != nil {
		t.Fatal("expected defaultStrategy to be nil after SetDefaultCompare")
	}
}

func TestCMapGen_SetDefaultStrategy_ClearsCompare(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.SetDefaultCompare(func(old, new int64) bool { return true })
	m.SetDefaultStrategy(neverStrategy[int64]{})
	m.Set("k", 5)

	if m.Replace("k", 99) {
		t.Fatal("expected Replace false because strategy never replaces")
	}
	if m.defaultCompare != nil {
		t.Fatal("expected defaultCompare to be nil after SetDefaultStrategy")
	}
}

func TestCMapGen_ReplaceWithCompare_ExistingKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("k", 10)
	compare := func(old, new int64) bool { return new > old }

	if !m.ReplaceWithCompare("k", 20, compare) {
		t.Fatal("expected true: 20 > 10")
	}
	val, _ := m.Get("k")
	if val != 20 {
		t.Fatalf("expected 20, got %d", val)
	}

	if m.ReplaceWithCompare("k", 5, compare) {
		t.Fatal("expected false: 5 < 20")
	}
}

func TestCMapGen_ReplaceWithCompare_NewKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	compare := func(old, new int64) bool { return new > old }

	if !m.ReplaceWithCompare("brand-new", 100, compare) {
		t.Fatal("expected true for brand-new key")
	}
	val, ok := m.Get("brand-new")
	if !ok || val != 100 {
		t.Fatalf("expected 100, got %d (ok=%v)", val, ok)
	}
}

func TestCMapGen_ReplaceWithStrategy_Always(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("k", 50)

	if !m.ReplaceWithStrategy("k", 100, AlwaysReplaceStrategy[int64]{}) {
		t.Fatal("expected true with AlwaysReplaceStrategy")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
}

func TestCMapGen_ReplaceWithStrategy_Never(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("k", 100)

	if m.ReplaceWithStrategy("k", 999, neverStrategy[int64]{}) {
		t.Fatal("expected false with neverStrategy")
	}
	val, _ := m.Get("k")
	if val != 100 {
		t.Fatalf("value should remain 100, got %d", val)
	}
}

func TestCMapGen_ReplaceWithStrategy_NewKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	// New key should always be inserted regardless of strategy
	if !m.ReplaceWithStrategy("fresh", 42, neverStrategy[int64]{}) {
		t.Fatal("expected true for new key even with neverStrategy")
	}
	val, ok := m.Get("fresh")
	if !ok || val != 42 {
		t.Fatalf("expected 42, got %d (ok=%v)", val, ok)
	}
}

func TestCMapGen_ReplaceAlways(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("k", 1)

	if !m.ReplaceAlways("k", 9999) {
		t.Fatal("expected ReplaceAlways to return true")
	}
	val, _ := m.Get("k")
	if val != 9999 {
		t.Fatalf("expected 9999, got %d", val)
	}
}

func TestCMapGen_ReplaceIfNotExists_NewKey(t *testing.T) {
	m := NewCMapGen[string, int64]()

	if !m.ReplaceIfNotExists("newkey", 42) {
		t.Fatal("expected true for non-existent key")
	}
	val, _ := m.Get("newkey")
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestCMapGen_ReplaceIfNotExists_ExistingKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("newkey", 42)

	if m.ReplaceIfNotExists("newkey", 100) {
		t.Fatal("expected false for existing key")
	}
	val, _ := m.Get("newkey")
	if val != 42 {
		t.Fatalf("value should remain 42, got %d", val)
	}
}

func TestCMapGen_UpsertWithCallback_NewKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	called := false

	m.UpsertWithCallback("k", 5, func(exists bool, old, new int64) int64 {
		called = true
		if exists {
			t.Error("expected exists=false for new key")
		}
		return new * 2
	})
	if !called {
		t.Fatal("callback was not called")
	}
	val, _ := m.Get("k")
	if val != 10 {
		t.Fatalf("expected 10, got %d", val)
	}
}

func TestCMapGen_UpsertWithCallback_ExistingKey(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.Set("k", 3)

	m.UpsertWithCallback("k", 7, func(exists bool, old, new int64) int64 {
		if !exists {
			t.Error("expected exists=true")
		}
		return old + new
	})
	val, _ := m.Get("k")
	if val != 10 {
		t.Fatalf("expected 10, got %d", val)
	}
}

// ---------------------------------------------------------------------------
// CMapGen.Replace — fallback new-key path (no strategy, no compare)
// ---------------------------------------------------------------------------

func TestCMapGen_Replace_NewKey_Fallback(t *testing.T) {
	m := &CMapGen[string, int64]{maps: newStringIntCmap()}
	// key does not exist → always stored
	if !m.Replace("fresh", 42) {
		t.Fatal("expected true for new key in fallback path")
	}
	val, ok := m.Get("fresh")
	if !ok || val != 42 {
		t.Fatalf("expected 42, got %d (ok=%v)", val, ok)
	}
}

// newStringIntCmap creates a fresh concurrent map for the fallback-path tests.
func newStringIntCmap() cmap.ConcurrentMap[string, int64] {
	return cmap.New[int64]()
}

// ---------------------------------------------------------------------------
// Strategy / Strategy helpers
// ---------------------------------------------------------------------------

func TestNumericGreaterStrategy(t *testing.T) {
	s := NumericGreaterStrategy[int64]{}
	if !s.ShouldReplace(1, 2) {
		t.Fatal("expected true: 2 > 1")
	}
	if s.ShouldReplace(2, 1) {
		t.Fatal("expected false: 1 < 2")
	}
	if s.ShouldReplace(5, 5) {
		t.Fatal("expected false: equal values")
	}
}

func TestAlwaysReplaceStrategy(t *testing.T) {
	s := AlwaysReplaceStrategy[int64]{}
	if !s.ShouldReplace(100, 1) {
		t.Fatal("expected always true")
	}
	if !s.ShouldReplace(0, 0) {
		t.Fatal("expected always true for equal")
	}
}

func TestTimestampStrategy(t *testing.T) {
	s := TimestampStrategy{MinInterval: 10}

	if !s.ShouldReplace(100, 110) {
		t.Fatal("expected true: difference 10 == MinInterval")
	}
	if !s.ShouldReplace(100, 120) {
		t.Fatal("expected true: difference 20 > MinInterval")
	}
	if s.ShouldReplace(100, 109) {
		t.Fatal("expected false: difference 9 < MinInterval")
	}
	if s.ShouldReplace(100, 100) {
		t.Fatal("expected false: difference 0 < MinInterval")
	}
	// Regression: older timestamp should not replace.
	if s.ShouldReplace(100, 90) {
		t.Fatal("expected false: new value is older")
	}
}

// ---------------------------------------------------------------------------
// keyToString — all type branches
// ---------------------------------------------------------------------------

func Test_keyToString(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"string", keyToString("hello"), "hello"},
		{"int", keyToString(42), "42"},
		{"int8", keyToString(int8(8)), "8"},
		{"int16", keyToString(int16(16)), "16"},
		{"int32", keyToString(int32(32)), "32"},
		{"int64", keyToString(int64(64)), "64"},
		{"uint", keyToString(uint(1)), "1"},
		{"uint8", keyToString(uint8(8)), "8"},
		{"uint16", keyToString(uint16(16)), "16"},
		{"uint32", keyToString(uint32(32)), "32"},
		{"uint64", keyToString(uint64(64)), "64"},
		{"float32", keyToString(float32(1.5)), "1.5"},
		{"float64", keyToString(float64(2.5)), "2.5"},
		{"bool true", keyToString(true), "true"},
		{"bool false", keyToString(false), "false"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("keyToString: want %q, got %q", c.want, c.got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// stringToKey — all type branches
// ---------------------------------------------------------------------------

func Test_stringToKey(t *testing.T) {
	if v := stringToKey[string]("hello"); v != "hello" {
		t.Errorf("string: want %q, got %q", "hello", v)
	}
	if v := stringToKey[int]("42"); v != 42 {
		t.Errorf("int: want 42, got %d", v)
	}
	if v := stringToKey[int8]("8"); v != 8 {
		t.Errorf("int8: want 8, got %d", v)
	}
	if v := stringToKey[int16]("16"); v != 16 {
		t.Errorf("int16: want 16, got %d", v)
	}
	if v := stringToKey[int32]("32"); v != 32 {
		t.Errorf("int32: want 32, got %d", v)
	}
	if v := stringToKey[int64]("64"); v != 64 {
		t.Errorf("int64: want 64, got %d", v)
	}
	if v := stringToKey[uint]("1"); v != 1 {
		t.Errorf("uint: want 1, got %d", v)
	}
	if v := stringToKey[uint8]("8"); v != 8 {
		t.Errorf("uint8: want 8, got %d", v)
	}
	if v := stringToKey[uint16]("16"); v != 16 {
		t.Errorf("uint16: want 16, got %d", v)
	}
	if v := stringToKey[uint32]("32"); v != 32 {
		t.Errorf("uint32: want 32, got %d", v)
	}
	if v := stringToKey[uint64]("64"); v != 64 {
		t.Errorf("uint64: want 64, got %d", v)
	}
	if v := stringToKey[float32]("1.5"); v != float32(1.5) {
		t.Errorf("float32: want 1.5, got %v", v)
	}
	if v := stringToKey[float64]("2.5"); v != 2.5 {
		t.Errorf("float64: want 2.5, got %v", v)
	}
	if v := stringToKey[bool]("true"); v != true {
		t.Errorf("bool true: want true, got %v", v)
	}
	if v := stringToKey[bool]("false"); v != false {
		t.Errorf("bool false: want false, got %v", v)
	}
}

// Test_stringToKey_InvalidInput verifies that malformed input returns the zero value without panic.
func Test_stringToKey_InvalidInput(t *testing.T) {
	if v := stringToKey[int]("not-a-number"); v != 0 {
		t.Errorf("invalid int input: want zero, got %d", v)
	}
	if v := stringToKey[float64]("nan-value"); v != 0 {
		t.Errorf("invalid float64 input: want zero, got %v", v)
	}
	if v := stringToKey[bool]("maybe"); v != false {
		t.Errorf("invalid bool input: want false, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Concurrent safety for CMapGen
// ---------------------------------------------------------------------------

func TestCMapGen_ConcurrentSetGet(t *testing.T) {
	m := NewCMapGen[int, int64]()
	const goroutines = 50
	const keysPerG = 20

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < keysPerG; i++ {
				key := id*keysPerG + i
				m.Set(key, int64(key))
				v, ok := m.Get(key)
				if !ok {
					return
				}
				if v != int64(key) {
					t.Errorf("key %d: expected %d, got %d", key, key, v)
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestCMapGen_ConcurrentReplace(t *testing.T) {
	m := NewCMapGen[string, int64]()
	m.SetDefaultCompare(func(old, new int64) bool { return new > old })
	m.Set("sym", 0)

	var wg sync.WaitGroup
	for g := 1; g <= 20; g++ {
		wg.Add(1)
		go func(v int64) {
			defer wg.Done()
			m.Replace("sym", v)
		}(int64(g * 100))
	}
	wg.Wait()

	val, ok := m.Get("sym")
	if !ok {
		t.Fatal("key 'sym' must still exist")
	}
	if val <= 0 {
		t.Fatalf("expected positive value, got %d", val)
	}
}
