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
	"sync"
	"testing"
)

// -----------------------------------------------------------------------
// CMap benchmarks (string→int64, built-in greater-value strategy)
// -----------------------------------------------------------------------

func BenchmarkCMap_Set(b *testing.B) {
	m := NewCMap()
	b.ResetTimer()
	for i := range b.N {
		m.Set("key", int64(i))
	}
}

func BenchmarkCMap_Get(b *testing.B) {
	m := NewCMap()
	m.Set("key", 42)
	b.ResetTimer()
	for range b.N {
		m.Get("key")
	}
}

func BenchmarkCMap_Replace(b *testing.B) {
	m := NewCMap()
	m.Set("key", 0)
	b.ResetTimer()
	for i := range b.N {
		m.Replace("key", int64(i))
	}
}

func BenchmarkCMap_Set_Parallel(b *testing.B) {
	m := NewCMap()
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			m.Set("shared", i)
			i++
		}
	})
}

func BenchmarkCMap_Replace_Parallel(b *testing.B) {
	m := NewCMap()
	m.Set("shared", 0)
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			m.Replace("shared", i)
			i++
		}
	})
}

// -----------------------------------------------------------------------
// CMapGen benchmarks (generic string→int64, NumericGreaterStrategy)
// -----------------------------------------------------------------------

func BenchmarkCMapGen_Set(b *testing.B) {
	m := NewCMapGen[string, int64]()
	b.ResetTimer()
	for i := range b.N {
		m.Set("key", int64(i))
	}
}

func BenchmarkCMapGen_Get(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("key", 42)
	b.ResetTimer()
	for range b.N {
		m.Get("key")
	}
}

func BenchmarkCMapGen_Replace_DefaultStrategy(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("key", 0)
	b.ResetTimer()
	for i := range b.N {
		m.Replace("key", int64(i))
	}
}

func BenchmarkCMapGen_Replace_WithCompare(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("key", 0)
	cmp := func(old, new int64) bool { return new > old }
	b.ResetTimer()
	for i := range b.N {
		m.ReplaceWithCompare("key", int64(i), cmp)
	}
}

func BenchmarkCMapGen_ReplaceAlways(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("key", 0)
	b.ResetTimer()
	for i := range b.N {
		m.ReplaceAlways("key", int64(i))
	}
}

func BenchmarkCMapGen_UpsertWithCallback(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("key", 0)
	cb := func(exists bool, old, new int64) int64 {
		if !exists || new > old {
			return new
		}
		return old
	}
	b.ResetTimer()
	for i := range b.N {
		m.UpsertWithCallback("key", int64(i), cb)
	}
}

func BenchmarkCMapGen_Set_Parallel(b *testing.B) {
	m := NewCMapGen[string, int64]()
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			m.Set("shared", i)
			i++
		}
	})
}

func BenchmarkCMapGen_Replace_Parallel(b *testing.B) {
	m := NewCMapGen[string, int64]()
	m.Set("shared", 0)
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			m.Replace("shared", i)
			i++
		}
	})
}

// -----------------------------------------------------------------------
// Concurrent read/write mixed — realistic workload
// -----------------------------------------------------------------------

// BenchmarkCMapGen_MixedReadWrite simulates a realistic workload where
// multiple goroutines read and write different keys concurrently.
func BenchmarkCMapGen_MixedReadWrite(b *testing.B) {
	const keyCount = 100
	m := NewCMapGen[int, int64]()
	for i := range keyCount {
		m.Set(i, int64(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			key := int(i) % keyCount
			if i%3 == 0 {
				m.Replace(key, i)
			} else {
				m.Get(key)
			}
			i++
		}
	})
}

// -----------------------------------------------------------------------
// stdlib sync.Map — reference baseline
// -----------------------------------------------------------------------

func BenchmarkSyncMap_Set(b *testing.B) {
	var m sync.Map
	b.ResetTimer()
	for i := range b.N {
		m.Store("key", int64(i))
	}
}

func BenchmarkSyncMap_Get(b *testing.B) {
	var m sync.Map
	m.Store("key", int64(42))
	b.ResetTimer()
	for range b.N {
		m.Load("key")
	}
}

func BenchmarkSyncMap_Set_Parallel(b *testing.B) {
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			m.Store("shared", i)
			i++
		}
	})
}
