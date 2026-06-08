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

package ringbuf

import (
	"sync"
	"testing"
)

// -----------------------------------------------------------------------
// RingSPSC benchmarks
// -----------------------------------------------------------------------

// BenchmarkSPSC_Push measures raw push throughput with a concurrent consumer
// draining items as fast as possible (capacity 4096, int64 items).
func BenchmarkSPSC_Push(b *testing.B) {
	rb := NewRingSPSC(RingSPSCConfig[int64]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	b.ResetTimer()
	for i := range b.N {
		rb.Push(int64(i))
	}
	b.StopTimer()

	rb.Close()
	<-done
}

// BenchmarkSPSC_TryPush measures non-blocking push throughput when the buffer
// has guaranteed space (consumer goroutine keeps it drained).
func BenchmarkSPSC_TryPush(b *testing.B) {
	rb := NewRingSPSC(RingSPSCConfig[int64]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	b.ResetTimer()
	for i := range b.N {
		for !rb.TryPush(int64(i)) {
		}
	}
	b.StopTimer()

	rb.Close()
	<-done
}

// BenchmarkSPSC_ProducerConsumer measures end-to-end throughput: one producer
// goroutine pushes items and one consumer goroutine processes them via ProcessFunc.
func BenchmarkSPSC_ProducerConsumer(b *testing.B) {
	var count int64
	var wg sync.WaitGroup
	wg.Add(1)

	rb := NewRingSPSC(RingSPSCConfig[int64]{
		Capacity: 4096,
		ProcessFunc: func(v int64) {
			count++
			if count == int64(b.N) {
				wg.Done()
			}
		},
	})

	b.ResetTimer()
	for i := range b.N {
		rb.Push(int64(i))
	}
	wg.Wait()
	b.StopTimer()

	rb.Close()
}

// BenchmarkSPSC_String benchmarks SPSC with string payloads to reflect a
// realistic log-forwarding or message-passing use case.
func BenchmarkSPSC_String(b *testing.B) {
	rb := NewRingSPSC(RingSPSCConfig[string]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	msg := "benchmark-message-payload"
	b.ResetTimer()
	for range b.N {
		rb.Push(msg)
	}
	b.StopTimer()

	rb.Close()
	<-done
}

// -----------------------------------------------------------------------
// RingMPSC benchmarks
// -----------------------------------------------------------------------

// BenchmarkMPSC_Push_1Producer measures MPSC push throughput from a single
// producer (comparable to SPSC for same-scenario overhead analysis).
func BenchmarkMPSC_Push_1Producer(b *testing.B) {
	rb := NewRingMPSC(RingMPSCConfig[int64]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	b.ResetTimer()
	for i := range b.N {
		rb.Push(int64(i))
	}
	b.StopTimer()

	rb.Close()
	<-done
}

// BenchmarkMPSC_Push_4Producers measures MPSC aggregate throughput from 4
// concurrent producer goroutines sharing one consumer.
func BenchmarkMPSC_Push_4Producers(b *testing.B) {
	const producers = 4
	rb := NewRingMPSC(RingMPSCConfig[int64]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	b.ResetTimer()
	var wg sync.WaitGroup
	each := b.N / producers
	for p := range producers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range each {
				rb.Push(int64(id*each + i))
			}
		}(p)
	}
	wg.Wait()
	b.StopTimer()

	rb.Close()
	<-done
}

// BenchmarkMPSC_Push_8Producers measures MPSC aggregate throughput from 8
// concurrent producer goroutines — typical for a heavily concurrent service.
func BenchmarkMPSC_Push_8Producers(b *testing.B) {
	const producers = 8
	rb := NewRingMPSC(RingMPSCConfig[int64]{Capacity: 4096})
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			if _, ok := rb.Pop(); !ok {
				return
			}
		}
	}()

	b.ResetTimer()
	var wg sync.WaitGroup
	each := b.N / producers
	for p := range producers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range each {
				rb.Push(int64(id*each + i))
			}
		}(p)
	}
	wg.Wait()
	b.StopTimer()

	rb.Close()
	<-done
}

// BenchmarkMPSC_ProducerConsumer measures end-to-end MPSC throughput via
// ProcessFunc (4 producers, auto consumer goroutine).
func BenchmarkMPSC_ProducerConsumer(b *testing.B) {
	const producers = 4
	var mu sync.Mutex
	var count int
	done := make(chan struct{})

	rb := NewRingMPSC(RingMPSCConfig[int64]{
		Capacity: 4096,
		ProcessFunc: func(v int64) {
			mu.Lock()
			count++
			if count == b.N {
				close(done)
			}
			mu.Unlock()
		},
	})

	b.ResetTimer()
	var wg sync.WaitGroup
	each := b.N / producers
	remainder := b.N - each*producers
	for p := range producers {
		n := each
		if p == 0 {
			n += remainder
		}
		wg.Add(1)
		go func(id, n int) {
			defer wg.Done()
			for i := range n {
				rb.Push(int64(id*each + i))
			}
		}(p, n)
	}
	wg.Wait()
	<-done
	b.StopTimer()

	rb.Close()
}

// -----------------------------------------------------------------------
// Go channel — reference baseline
// -----------------------------------------------------------------------

// BenchmarkChannel_Unbuffered_Push provides a channel baseline for comparison.
func BenchmarkChannel_Buffered4096_Push(b *testing.B) {
	ch := make(chan int64, 4096)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range ch {
		}
	}()

	b.ResetTimer()
	for i := range b.N {
		ch <- int64(i)
	}
	b.StopTimer()

	close(ch)
	<-done
}
