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

// package ringbuf_test demonstrates the public API from a caller's perspective.
package ringbuf_test

import (
	"fmt"
	"sync"

	"github.com/phcp-tech/common-library-golang/ringbuf"
)

// -----------------------------------------------------------------------
// RingSPSC – Single Producer Single Consumer
// -----------------------------------------------------------------------

// ExampleNewRingSPSC_processFunc shows the most common usage pattern: supply a
// ProcessFunc so the ring buffer starts a consumer goroutine automatically.
// The producer calls Push; the consumer goroutine calls ProcessFunc for each item.
// Close drains all remaining items and waits for the consumer to finish.
func ExampleNewRingSPSC_processFunc() {
	var wg sync.WaitGroup
	wg.Add(3)

	rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[int]{
		Capacity: 8,
		ProcessFunc: func(v int) {
			fmt.Println(v)
			wg.Done()
		},
	})

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	wg.Wait() // wait until all items have been processed
	rb.Close()
	// Output:
	// 1
	// 2
	// 3
}

// ExampleNewRingSPSC_manualPop shows manual consumer mode: omit ProcessFunc and
// drive the Pop loop yourself. Useful when the consumer needs full control over
// scheduling or batching.
func ExampleNewRingSPSC_manualPop() {
	rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[string]{
		Capacity: 4,
		// ProcessFunc is nil: no automatic consumer goroutine
	})

	// Producer goroutine
	go func() {
		rb.Push("hello")
		rb.Push("world")
		rb.Close() // signal that no more items will be pushed
	}()

	// Consumer in the current goroutine: drain until closed and empty
	for {
		item, ok := rb.Pop()
		if !ok {
			break
		}
		fmt.Println(item)
	}
	// Output:
	// hello
	// world
}

// ExampleRingSPSC_Push shows blocking push. Unlike TryPush, Push waits until
// space is available when the buffer is full, providing natural backpressure
// to the single producer. Must only be called from one goroutine at a time.
func ExampleRingSPSC_Push() {
	var wg sync.WaitGroup
	wg.Add(3)

	rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[string]{
		Capacity: 4,
		ProcessFunc: func(s string) {
			fmt.Println(s)
			wg.Done()
		},
	})

	// Push from the single producer goroutine; blocks only when the buffer is full.
	rb.Push("first")
	rb.Push("second")
	rb.Push("third")

	wg.Wait()
	rb.Close()
	// Output:
	// first
	// second
	// third
}

// ExampleRingSPSC_TryPush shows non-blocking push. TryPush returns false
// immediately when the buffer is full instead of blocking the caller.
func ExampleRingSPSC_TryPush() {
	rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[int]{Capacity: 2})
	defer rb.Close()

	fmt.Println(rb.TryPush(10)) // true  – slot available
	fmt.Println(rb.TryPush(20)) // true  – slot available
	fmt.Println(rb.TryPush(30)) // false – buffer full (capacity 2)
	// Output:
	// true
	// true
	// false
}

// ExampleRingSPSC_TryPop shows non-blocking pop. TryPop returns false
// immediately when the buffer is empty instead of blocking the caller.
func ExampleRingSPSC_TryPop() {
	rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[int]{Capacity: 4})
	defer rb.Close()

	rb.TryPush(42)

	v, ok := rb.TryPop()
	fmt.Println(v, ok) // 42 true

	_, ok = rb.TryPop()
	fmt.Println(ok) // false – buffer now empty
	// Output:
	// 42 true
	// false
}

// -----------------------------------------------------------------------
// RingMPSC – Multiple Producers Single Consumer
// -----------------------------------------------------------------------

// ExampleNewRingMPSC_processFunc shows the typical MPSC pattern: multiple
// producer goroutines push concurrently while a single consumer goroutine
// (started automatically via ProcessFunc) processes each item sequentially.
func ExampleNewRingMPSC_processFunc() {
	const producers = 3
	var (
		producerWg sync.WaitGroup
		consumerWg sync.WaitGroup
	)
	consumerWg.Add(producers)

	rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[string]{
		Capacity: 16,
		ProcessFunc: func(msg string) {
			// consumer goroutine: called once per message, sequentially
			_ = msg
			consumerWg.Done()
		},
	})

	// Launch multiple producer goroutines
	for i := range producers {
		producerWg.Add(1)
		go func(id int) {
			defer producerWg.Done()
			rb.Push(fmt.Sprintf("producer-%d", id))
		}(i)
	}

	producerWg.Wait()  // wait for all producers to finish pushing
	consumerWg.Wait()  // wait for all items to be processed
	rb.Close()
}

// ExampleRingMPSC_TryPush shows non-blocking push for the MPSC buffer.
// Safe to call from multiple goroutines simultaneously.
func ExampleRingMPSC_TryPush() {
	rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[int]{Capacity: 2})
	defer rb.Close()

	fmt.Println(rb.TryPush(10)) // true  – slot available
	fmt.Println(rb.TryPush(20)) // true  – slot available
	fmt.Println(rb.TryPush(30)) // false – buffer full
	// Output:
	// true
	// true
	// false
}

// ExampleRingMPSC_TryPop shows non-blocking pop for the MPSC buffer.
// TryPop is intended for use by the single consumer only.
func ExampleRingMPSC_TryPop() {
	rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[int]{Capacity: 4})
	defer rb.Close()

	rb.TryPush(42)

	v, ok := rb.TryPop()
	fmt.Println(v, ok) // 42 true

	_, ok = rb.TryPop()
	fmt.Println(ok) // false – buffer now empty
	// Output:
	// 42 true
	// false
}

// ExampleNewRingMPSC_manualPop shows manual consumer mode without ProcessFunc.
// Multiple producers push concurrently while the caller drives the Pop loop.
func ExampleNewRingMPSC_manualPop() {
	rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[int]{
		Capacity: 16,
		// ProcessFunc is nil: no automatic consumer goroutine
	})

	const n = 5
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			rb.Push(v)
		}(i)
	}

	// Close after all producers finish so the Pop loop can terminate.
	go func() {
		wg.Wait()
		rb.Close()
	}()

	count := 0
	for {
		_, ok := rb.Pop()
		if !ok {
			break
		}
		count++
	}
	fmt.Println(count)
	// Output:
	// 5
}

// ExampleRingMPSC_Push shows blocking push. Unlike TryPush, Push waits until
// space is available when the buffer is full, providing natural backpressure
// to the producer. It is safe to call from multiple goroutines concurrently.
func ExampleRingMPSC_Push() {
	var wg sync.WaitGroup
	wg.Add(3)

	rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[string]{
		Capacity: 4,
		ProcessFunc: func(s string) {
			fmt.Println(s)
			wg.Done()
		},
	})

	// Push from the current goroutine; blocks only when the buffer is full.
	rb.Push("first")
	rb.Push("second")
	rb.Push("third")

	wg.Wait()
	rb.Close()
	// Output:
	// first
	// second
	// third
}
