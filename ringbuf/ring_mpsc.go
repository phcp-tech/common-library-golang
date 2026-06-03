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
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

// RingMPSCConfig holds configuration for RingMPSC
// T is the type of messages being processed
type RingMPSCConfig[T any] struct {
	// Capacity is the buffer size of the ring buffer
	// Default: 1024
	Capacity uint64

	// ProcessFunc is the function to process each message
	// Optional: if provided, a consumer goroutine will be automatically started
	// If nil, you need to manually call Pop() to consume messages
	ProcessFunc ProcessFunc[T]
}

// RingMPSC is a high-performance ring buffer for multiple producers and a single consumer
// Uses a mutex to serialize concurrent writes from multiple producers
type RingMPSC[T any] struct {
	buffer      []T            // offset 0:   24 bytes (slice header: ptr+len+cap)
	capacity    uint64         // offset 24:  8 bytes
	mask        uint64         // offset 32:  8 bytes
	notifyC     chan struct{}  // offset 40:  8 bytes (channel pointer)
	mu          sync.Mutex     // offset 48:  8 bytes
	closed      atomic.Bool    // offset 56:  4 bytes (uint32 internally)
	processFunc ProcessFunc[T] // offset 64:  8 bytes (4 bytes padding before for 8-byte alignment)
	wg          sync.WaitGroup // offset 72:  16 bytes
	_padding1   [40]byte       // offset 88:  40 bytes padding to align writePos to 128-byte cache line
	writePos    uint64         // offset 128: write position (separate cache line)
	_padding2   [56]byte       // cache line padding (128+8+56=192)
	readPos     uint64         // offset 192: read position (separate cache line)
	_padding3   [56]byte       // cache line padding (192+8+56=256)
}

// NewRingMPSC creates a new MPSC ring buffer with the given configuration
func NewRingMPSC[T any](cfg RingMPSCConfig[T]) *RingMPSC[T] {
	// Set defaults
	if cfg.Capacity == 0 {
		cfg.Capacity = ringBufCapacity
	}

	// capacity: must be a power of 2, will be automatically rounded up if not
	cfg.Capacity = roundUpToPowerOfTwo(cfg.Capacity)

	rb := &RingMPSC[T]{
		buffer:      make([]T, cfg.Capacity),
		capacity:    cfg.Capacity,
		mask:        cfg.Capacity - 1,
		notifyC:     make(chan struct{}, 1), // buffered to avoid blocking producer
		processFunc: cfg.ProcessFunc,
		writePos:    0,
		readPos:     0,
	}

	// If ProcessFunc is provided, start consumer goroutine
	if rb.processFunc != nil {
		rb.start()
	}

	return rb
}

// TryPush writes data by producer (non-blocking)
// Returns true on success, false if buffer is full or closed
func (rb *RingMPSC[T]) TryPush(item T) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// Check if closed first
	if rb.closed.Load() {
		return false
	}

	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)

	if writePos-readPos >= rb.capacity {
		return false // buffer is full
	}

	rb.buffer[writePos&rb.mask] = item

	atomic.StoreUint64(&rb.writePos, writePos+1)

	select {
	case rb.notifyC <- struct{}{}:
	default:
	}

	return true
}

// Push writes data by producer with blocking
// Blocks and waits until space is available if buffer is full
// Returns false if ring buffer is closed, true on success
func (rb *RingMPSC[T]) Push(item T) bool {
	for {
		// Check if closed first to avoid unnecessary lock acquisition
		if rb.closed.Load() {
			return false
		}

		// Try non-blocking write
		if rb.TryPush(item) {
			return true
		}

		// Buffer is full, yield CPU time slice to give consumer a chance to execute
		runtime.Gosched()
	}
}

// TryPop reads data by consumer (non-blocking)
// Returns the data and a flag indicating success
func (rb *RingMPSC[T]) TryPop() (T, bool) {
	readPos := atomic.LoadUint64(&rb.readPos)
	writePos := atomic.LoadUint64(&rb.writePos)

	if readPos == writePos {
		var zero T
		return zero, false // buffer is empty
	}

	item := rb.buffer[readPos&rb.mask]

	atomic.StoreUint64(&rb.readPos, readPos+1)

	return item, true
}

// Pop reads data by consumer with blocking
// Blocks and waits until data arrives if buffer is empty
// Returns the data and success flag (false indicates closed AND buffer is empty)
// When closed, will drain all remaining data before returning false
func (rb *RingMPSC[T]) Pop() (T, bool) {
	for {
		// Try non-blocking read first
		if item, ok := rb.TryPop(); ok {
			return item, true
		}

		// Buffer is empty, check if closed
		if rb.closed.Load() {
			// Double check: try one more time to ensure no data was added
			// between TryPop and closed check (race condition protection)
			if item, ok := rb.TryPop(); ok {
				return item, true
			}
			// Closed and buffer is truly empty, we're done
			var zero T
			return zero, false
		}

		// Buffer is empty and not closed, wait for notification
		// When notifyC is closed, this returns immediately and loop continues
		<-rb.notifyC
	}
}

// Length returns the approximate number of items currently in the buffer.
// The returned length is approximate and may be stale by the time it's used.
func (rb *RingMPSC[T]) Length() uint64 {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return writePos - readPos
}

// Capacity returns the capacity of the buffer
func (rb *RingMPSC[T]) Capacity() uint64 {
	return rb.capacity
}

// IsEmpty reports whether the ring buffer contains no items.
// It suffers from a race condition where the result may be stale immediately after return.
func (rb *RingMPSC[T]) IsEmpty() bool {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return readPos == writePos
}

// IsFull reports whether the ring buffer is at capacity.
// It suffers from a race condition where the result may be stale immediately after return.
func (rb *RingMPSC[T]) IsFull() bool {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return writePos-readPos >= rb.capacity
}

// Close closes the ring buffer and wakes up blocked consumers
// After close, producers cannot push new data, but consumers will drain remaining data
// Safe to call multiple times
func (rb *RingMPSC[T]) Close() {
	if rb.closed.CompareAndSwap(false, true) {
		// Acquire lock to prevent race with TryPush
		rb.mu.Lock()
		close(rb.notifyC)
		rb.mu.Unlock()
		rb.wg.Wait()
	}
}

// IsClosed returns whether the ring buffer is closed
func (rb *RingMPSC[T]) IsClosed() bool {
	return rb.closed.Load()
}

// start launches the consumer goroutine
func (rb *RingMPSC[T]) start() {
	rb.wg.Add(1)
	go rb.worker()
}

// worker is the goroutine that processes messages from the ring buffer
func (rb *RingMPSC[T]) worker() {
	defer rb.wg.Done()

	// Recover from potential panic in processing logic to avoid crashing the process
	defer func() {
		if r := recover(); r != nil {
			log.Printf("RingMPSC worker panic: %v", r)
		}
	}()

	for {
		// Block and read data, process immediately when data is available
		msg, ok := rb.Pop()
		if !ok {
			// ring buffer is closed and empty
			return
		}

		// Process message
		rb.processFunc(msg)
	}
}
