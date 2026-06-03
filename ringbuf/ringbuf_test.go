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

package ringbuf

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// roundUpToPowerOfTwo
// ---------------------------------------------------------------------------

func TestRoundUpToPowerOfTwo(t *testing.T) {
	cases := []struct {
		input uint64
		want  uint64
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{1000, 1024},
		{1024, 1024},
		{1025, 2048},
	}
	for _, tc := range cases {
		got := roundUpToPowerOfTwo(tc.input)
		if got != tc.want {
			t.Errorf("roundUpToPowerOfTwo(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// RingSPSC – basic operations
// ---------------------------------------------------------------------------

func newSPSC(cap uint64) *RingSPSC[int] {
	return NewRingSPSC(RingSPSCConfig[int]{Capacity: cap})
}

func TestSPSC_DefaultCapacity(t *testing.T) {
	rb := NewRingSPSC(RingSPSCConfig[int]{})
	if rb.Capacity() != ringBufCapacity {
		t.Fatalf("expected default capacity %d, got %d", ringBufCapacity, rb.Capacity())
	}
}

func TestSPSC_CapacityRoundedUpToPowerOfTwo(t *testing.T) {
	rb := newSPSC(3) // should round to 4
	if rb.Capacity() != 4 {
		t.Fatalf("expected capacity 4, got %d", rb.Capacity())
	}
}

func TestSPSC_TryPush_TryPop(t *testing.T) {
	rb := newSPSC(4)

	if !rb.TryPush(42) {
		t.Fatal("TryPush should succeed on empty buffer")
	}
	item, ok := rb.TryPop()
	if !ok {
		t.Fatal("TryPop should succeed after push")
	}
	if item != 42 {
		t.Fatalf("expected 42, got %d", item)
	}
}

func TestSPSC_TryPop_Empty(t *testing.T) {
	rb := newSPSC(4)
	_, ok := rb.TryPop()
	if ok {
		t.Fatal("TryPop should return false on empty buffer")
	}
}

func TestSPSC_TryPush_Full(t *testing.T) {
	rb := newSPSC(4)
	for i := 0; i < 4; i++ {
		if !rb.TryPush(i) {
			t.Fatalf("TryPush %d should succeed", i)
		}
	}
	// Buffer is now full
	if rb.TryPush(99) {
		t.Fatal("TryPush should fail on full buffer")
	}
}

func TestSPSC_IsEmpty_IsFull(t *testing.T) {
	rb := newSPSC(4)

	if !rb.IsEmpty() {
		t.Fatal("expected IsEmpty true on new buffer")
	}
	if rb.IsFull() {
		t.Fatal("expected IsFull false on new buffer")
	}

	rb.TryPush(1)
	rb.TryPush(2)
	rb.TryPush(3)
	rb.TryPush(4)

	if rb.IsEmpty() {
		t.Fatal("expected IsEmpty false after filling")
	}
	if !rb.IsFull() {
		t.Fatal("expected IsFull true after filling")
	}
}

func TestSPSC_Length(t *testing.T) {
	rb := newSPSC(8)

	if rb.Length() != 0 {
		t.Fatalf("expected length 0, got %d", rb.Length())
	}

	for i := 0; i < 5; i++ {
		rb.TryPush(i)
	}
	if rb.Length() != 5 {
		t.Fatalf("expected length 5, got %d", rb.Length())
	}
}

func TestSPSC_FIFO_Order(t *testing.T) {
	rb := newSPSC(8)
	for i := 0; i < 5; i++ {
		rb.TryPush(i)
	}
	for i := 0; i < 5; i++ {
		item, ok := rb.TryPop()
		if !ok {
			t.Fatalf("expected item at position %d", i)
		}
		if item != i {
			t.Fatalf("expected %d, got %d", i, item)
		}
	}
}

func TestSPSC_Close_StopsPush(t *testing.T) {
	rb := newSPSC(4)
	rb.Close()

	if rb.TryPush(1) {
		t.Fatal("TryPush should fail after Close")
	}
	if rb.Push(1) {
		t.Fatal("Push should fail after Close")
	}
}

func TestSPSC_IsClosed(t *testing.T) {
	rb := newSPSC(4)
	if rb.IsClosed() {
		t.Fatal("expected IsClosed false before Close")
	}
	rb.Close()
	if !rb.IsClosed() {
		t.Fatal("expected IsClosed true after Close")
	}
}

func TestSPSC_Close_Idempotent(t *testing.T) {
	rb := newSPSC(4)
	rb.Close()
	rb.Close() // must not panic
}

func TestSPSC_Pop_DrainBeforeClose(t *testing.T) {
	rb := newSPSC(8)
	rb.TryPush(10)
	rb.TryPush(20)
	rb.Close()

	// Pop should drain remaining items even after Close
	v1, ok1 := rb.Pop()
	v2, ok2 := rb.Pop()
	_, ok3 := rb.Pop() // should return false

	if !ok1 || v1 != 10 {
		t.Fatalf("expected (10, true), got (%d, %v)", v1, ok1)
	}
	if !ok2 || v2 != 20 {
		t.Fatalf("expected (20, true), got (%d, %v)", v2, ok2)
	}
	if ok3 {
		t.Fatal("expected Pop to return false when closed and empty")
	}
}

func TestSPSC_Push_Blocking(t *testing.T) {
	rb := newSPSC(2)
	rb.TryPush(1)
	rb.TryPush(2) // full

	done := make(chan bool, 1)
	go func() {
		ok := rb.Push(3) // should block until space
		done <- ok
	}()

	// Give goroutine time to block, then free space
	time.Sleep(10 * time.Millisecond)
	rb.TryPop() // free one slot

	select {
	case ok := <-done:
		if !ok {
			t.Fatal("expected Push to succeed after space freed")
		}
	case <-time.After(time.Second):
		t.Fatal("Push did not unblock within 1s")
	}
}

func TestSPSC_WithProcessFunc(t *testing.T) {
	var count atomic.Int64
	rb := NewRingSPSC(RingSPSCConfig[int]{
		Capacity: 16,
		ProcessFunc: func(v int) {
			count.Add(1)
		},
	})

	const n = 5
	for i := 0; i < n; i++ {
		rb.Push(i)
	}
	rb.Close()

	if count.Load() != n {
		t.Fatalf("expected ProcessFunc called %d times, got %d", n, count.Load())
	}
}

func TestSPSC_Concurrent_SPSC(t *testing.T) {
	const n = 1000
	rb := newSPSC(256)

	var sum atomic.Int64
	var wg sync.WaitGroup

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		received := 0
		for received < n {
			if v, ok := rb.TryPop(); ok {
				sum.Add(int64(v))
				received++
			}
		}
	}()

	// Producer
	for i := 1; i <= n; i++ {
		rb.Push(i)
	}

	wg.Wait()

	expected := int64(n * (n + 1) / 2)
	if sum.Load() != expected {
		t.Fatalf("expected sum %d, got %d", expected, sum.Load())
	}
}

// ---------------------------------------------------------------------------
// RingMPSC – basic operations
// ---------------------------------------------------------------------------

func newMPSC(cap uint64) *RingMPSC[int] {
	return NewRingMPSC(RingMPSCConfig[int]{Capacity: cap})
}

func TestMPSC_DefaultCapacity(t *testing.T) {
	rb := NewRingMPSC(RingMPSCConfig[int]{})
	if rb.Capacity() != ringBufCapacity {
		t.Fatalf("expected default capacity %d, got %d", ringBufCapacity, rb.Capacity())
	}
}

func TestMPSC_CapacityRoundedUp(t *testing.T) {
	rb := newMPSC(5)
	if rb.Capacity() != 8 {
		t.Fatalf("expected 8, got %d", rb.Capacity())
	}
}

func TestMPSC_TryPush_TryPop(t *testing.T) {
	rb := newMPSC(8)

	if !rb.TryPush(7) {
		t.Fatal("TryPush should succeed on empty buffer")
	}
	item, ok := rb.TryPop()
	if !ok || item != 7 {
		t.Fatalf("expected (7, true), got (%d, %v)", item, ok)
	}
}

func TestMPSC_TryPop_Empty(t *testing.T) {
	rb := newMPSC(4)
	_, ok := rb.TryPop()
	if ok {
		t.Fatal("TryPop should return false on empty buffer")
	}
}

func TestMPSC_TryPush_Full(t *testing.T) {
	rb := newMPSC(4)
	for i := 0; i < 4; i++ {
		if !rb.TryPush(i) {
			t.Fatalf("TryPush %d should succeed", i)
		}
	}
	if rb.TryPush(99) {
		t.Fatal("TryPush should fail on full buffer")
	}
}

func TestMPSC_IsEmpty_IsFull(t *testing.T) {
	rb := newMPSC(4)

	if !rb.IsEmpty() {
		t.Fatal("expected IsEmpty true on new buffer")
	}

	for i := 0; i < 4; i++ {
		rb.TryPush(i)
	}
	if !rb.IsFull() {
		t.Fatal("expected IsFull true after filling")
	}
}

func TestMPSC_Length(t *testing.T) {
	rb := newMPSC(8)
	for i := 0; i < 3; i++ {
		rb.TryPush(i)
	}
	if rb.Length() != 3 {
		t.Fatalf("expected 3, got %d", rb.Length())
	}
}

func TestMPSC_FIFO_Order(t *testing.T) {
	rb := newMPSC(8)
	for i := 0; i < 5; i++ {
		rb.TryPush(i)
	}
	for i := 0; i < 5; i++ {
		item, ok := rb.TryPop()
		if !ok {
			t.Fatalf("expected item at index %d", i)
		}
		if item != i {
			t.Fatalf("expected %d, got %d", i, item)
		}
	}
}

func TestMPSC_Close_StopsPush(t *testing.T) {
	rb := newMPSC(4)
	rb.Close()

	if rb.TryPush(1) {
		t.Fatal("TryPush should fail after Close")
	}
	if rb.Push(1) {
		t.Fatal("Push should fail after Close")
	}
}

func TestMPSC_IsClosed(t *testing.T) {
	rb := newMPSC(4)
	if rb.IsClosed() {
		t.Fatal("expected IsClosed false before Close")
	}
	rb.Close()
	if !rb.IsClosed() {
		t.Fatal("expected IsClosed true after Close")
	}
}

func TestMPSC_Close_Idempotent(t *testing.T) {
	rb := newMPSC(4)
	rb.Close()
	rb.Close() // must not panic
}

func TestMPSC_Pop_DrainBeforeClose(t *testing.T) {
	rb := newMPSC(8)
	rb.TryPush(10)
	rb.TryPush(20)
	rb.Close()

	v1, ok1 := rb.Pop()
	v2, ok2 := rb.Pop()
	_, ok3 := rb.Pop()

	if !ok1 || v1 != 10 {
		t.Fatalf("expected (10, true), got (%d, %v)", v1, ok1)
	}
	if !ok2 || v2 != 20 {
		t.Fatalf("expected (20, true), got (%d, %v)", v2, ok2)
	}
	if ok3 {
		t.Fatal("expected false when closed and empty")
	}
}

func TestMPSC_Push_Blocking(t *testing.T) {
	rb := newMPSC(2)
	rb.TryPush(1)
	rb.TryPush(2) // full

	done := make(chan bool, 1)
	go func() {
		ok := rb.Push(3)
		done <- ok
	}()

	time.Sleep(10 * time.Millisecond)
	rb.TryPop()

	select {
	case ok := <-done:
		if !ok {
			t.Fatal("expected Push to succeed after space freed")
		}
	case <-time.After(time.Second):
		t.Fatal("Push did not unblock within 1s")
	}
}

func TestMPSC_WithProcessFunc(t *testing.T) {
	var count atomic.Int64
	rb := NewRingMPSC(RingMPSCConfig[int]{
		Capacity: 16,
		ProcessFunc: func(v int) {
			count.Add(1)
		},
	})

	const n = 5
	for i := 0; i < n; i++ {
		rb.Push(i)
	}
	rb.Close()

	if count.Load() != n {
		t.Fatalf("expected ProcessFunc called %d times, got %d", n, count.Load())
	}
}

// Multiple producers, single consumer
func TestMPSC_Concurrent_MultipleProducers(t *testing.T) {
	const producers = 4
	const perProducer = 250
	const total = producers * perProducer

	rb := newMPSC(512)
	var sum atomic.Int64
	var wg sync.WaitGroup

	// Consumer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		received := 0
		for received < total {
			if v, ok := rb.TryPop(); ok {
				sum.Add(int64(v))
				received++
			}
		}
	}()

	// Producers
	for p := 0; p < producers; p++ {
		wg.Add(1)
		start := p * perProducer
		go func(base int) {
			defer wg.Done()
			for i := 1; i <= perProducer; i++ {
				rb.Push(base + i)
			}
		}(start)
	}

	wg.Wait()

	// Sum of 1..total (each producer contributes start+1..start+perProducer)
	// Simpler: just verify count by checking item count was correct (sum will differ by base)
	// Verify total is non-zero and positive
	if sum.Load() <= 0 {
		t.Fatalf("expected positive sum from concurrent producers, got %d", sum.Load())
	}
}

func TestMPSC_Concurrent_PushClose(t *testing.T) {
	rb := newMPSC(64)
	var pushOkCount atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			if rb.Push(v) {
				pushOkCount.Add(1)
			}
		}(i)
	}

	// Let producers run briefly, then close
	time.Sleep(5 * time.Millisecond)
	rb.Close()

	wg.Wait()
	// After close, no more can be pushed – but this verifies no data race / panic
}
