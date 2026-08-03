package cache

// Benchmark suite test OtterCache multiple scenarios.
//
// Run all benchmarks:
//
//	go test -bench=. -benchmem -benchtime=3s ./cache/
//
// Run with CPU profiling:
//
//	go test -bench=. -benchmem -cpuprofile=cpu.prof ./cache/
//
// Scenarios covered:
//   - Set             : write a new key every iteration (no reuse)
//   - Get/Hit         : read a pre-populated key (100% hit rate)
//   - Get/Miss        : read a key that was never written (0% hit rate)
//   - Mixed (80/20)   : 80% reads, 20% writes, simulating typical workload
//   - Parallel        : concurrent reads/writes (GOMAXPROCS goroutines)
//   - Delete          : delete an existing key every iteration
//   - LargeValue      : store and retrieve a 1 KB value

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// prePopulate writes n keys into the cache so Get benchmarks always hit.
func prePopulate(c ICache, n int) {
	for i := range n {
		_ = c.Set(fmt.Sprintf("key-%d", i), i, 0)
	}
	// OtterCache buffers writes asynchronously; give it a moment to settle.
	time.Sleep(10 * time.Millisecond)
}

const (
	preloadSize = 10_000 // keys pre-loaded for hit benchmarks
	keySpace    = 10_000 // key space for mixed / parallel benchmarks
)

var largeValue = make([]byte, 1024) // 1 KB payload

// ── Set ──────────────────────────────────────────────────────────────────────

func BenchmarkOtterSet(b *testing.B) {
	c := NewOtterCache()
	i := 0
	for b.Loop() {
		_ = c.Set(fmt.Sprintf("key-%d", i), i, 0)
		i++
	}
}

// ── Set with TTL ──────────────────────────────────────────────────────────────

func BenchmarkOtterSetWithTTL(b *testing.B) {
	c := NewOtterCache()
	ttl := 5 * time.Minute
	i := 0
	for b.Loop() {
		_ = c.Set(fmt.Sprintf("key-%d", i), i, ttl)
		i++
	}
}

// ── Get / Hit ─────────────────────────────────────────────────────────────────

func BenchmarkOtterGetHit(b *testing.B) {
	c := NewOtterCache()
	prePopulate(c, preloadSize)
	i := 0
	for b.Loop() {
		c.Get(fmt.Sprintf("key-%d", i%preloadSize))
		i++
	}
}

// ── Get / Miss ────────────────────────────────────────────────────────────────

func BenchmarkOtterGetMiss(b *testing.B) {
	c := NewOtterCache()
	i := 0
	for b.Loop() {
		c.Get(fmt.Sprintf("miss-%d", i))
		i++
	}
}

// ── Mixed 80 % Read / 20 % Write ─────────────────────────────────────────────

func BenchmarkOtterMixed(b *testing.B) {
	c := NewOtterCache()
	prePopulate(c, keySpace)
	i := 0
	for b.Loop() {
		key := fmt.Sprintf("key-%d", i%keySpace)
		if i%5 == 0 {
			_ = c.Set(key, i, 0)
		} else {
			c.Get(key)
		}
		i++
	}
}

// ── Parallel (concurrent reads + writes) ─────────────────────────────────────

func BenchmarkOtterParallel(b *testing.B) {
	c := NewOtterCache()
	prePopulate(c, keySpace)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%keySpace)
			if i%5 == 0 {
				_ = c.Set(key, i, 0)
			} else {
				c.Get(key)
			}
			i++
		}
	})
}

// ── Delete ────────────────────────────────────────────────────────────────────

func BenchmarkOtterDelete(b *testing.B) {
	c := NewOtterCache()
	prePopulate(c, preloadSize)
	i := 0
	for b.Loop() {
		_ = c.Delete(fmt.Sprintf("key-%d", i%preloadSize))
		i++
	}
}

// ── Large Value (1 KB) ────────────────────────────────────────────────────────

func BenchmarkOtterSetLargeValue(b *testing.B) {
	c := NewOtterCache()
	b.SetBytes(int64(len(largeValue)))
	i := 0
	for b.Loop() {
		_ = c.Set(fmt.Sprintf("key-%d", i%keySpace), largeValue, 0)
		i++
	}
}

func BenchmarkOtterGetLargeValue(b *testing.B) {
	c := NewOtterCache()
	for i := range keySpace {
		_ = c.Set(fmt.Sprintf("key-%d", i), largeValue, 0)
	}
	time.Sleep(10 * time.Millisecond)
	b.SetBytes(int64(len(largeValue)))
	i := 0
	for b.Loop() {
		c.Get(fmt.Sprintf("key-%d", i%keySpace))
		i++
	}
}

// ── Zipf / skewed access pattern ─────────────────────────────────────────────
// Simulates real-world "hot key" workloads where a small fraction of keys
// receive the majority of requests (Zipf distribution, s=1.1).

func BenchmarkOtterZipf(b *testing.B) {
	c := NewOtterCache()
	prePopulate(c, keySpace)
	z := rand.NewZipf(rand.New(rand.NewPCG(42, 0)), 1.1, 1, uint64(keySpace-1))
	for b.Loop() {
		c.Get(fmt.Sprintf("key-%d", z.Uint64()))
	}
}
