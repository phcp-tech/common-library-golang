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

const ringBufCapacity uint64 = 1024

// ProcessFunc defines the function signature for processing messages
// T can be any type: []byte, string, struct, etc.
type ProcessFunc[T any] func(T)

// RingSPSCConfig holds configuration for RingSPSC
// T is the type of messages being processed
type RingSPSCConfig[T any] struct {
	// Capacity is the buffer size of the ring buffer
	// Default: 1024
	Capacity uint64

	// ProcessFunc is the function to process each message
	// Optional: if provided, a consumer goroutine will be automatically started
	// If nil, you need to manually call Pop() to consume messages
	ProcessFunc ProcessFunc[T]
}

// RingSPSC is a lock-free ring buffer for single producer and single consumer
// Implements high-performance SPSC queue using atomic operations, ensuring strict FIFO ordering
type RingSPSC[T any] struct {
	buffer      []T            // offset 0:   24 bytes (slice header: ptr+len+cap)
	capacity    uint64         // offset 24:  8 bytes
	mask        uint64         // offset 32:  8 bytes
	notifyC     chan struct{}  // offset 40:  8 bytes (channel pointer)
	closed      atomic.Bool    // offset 48:  4 bytes (uint32 internally)
	processFunc ProcessFunc[T] // offset 56:  8 bytes (4 bytes padding before for 8-byte alignment)
	wg          sync.WaitGroup // offset 64:  16 bytes
	_padding1   [48]byte       // offset 80:  48 bytes padding to align writePos to 128-byte cache line
	writePos    uint64         // offset 128: write position (separate cache line)
	_padding2   [56]byte       // cache line padding (128+8+56=192)
	readPos     uint64         // offset 192: read position (separate cache line)
	_padding3   [56]byte       // cache line padding (192+8+56=256)
}

// NewRingSPSC creates a new SPSC ring buffer with the given configuration
func NewRingSPSC[T any](cfg RingSPSCConfig[T]) *RingSPSC[T] {
	// Set defaults
	if cfg.Capacity == 0 {
		cfg.Capacity = ringBufCapacity
	}

	// capacity: must be a power of 2, will be automatically rounded up if not
	cfg.Capacity = roundUpToPowerOfTwo(cfg.Capacity)

	rb := &RingSPSC[T]{
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
func (rb *RingSPSC[T]) TryPush(item T) bool {
	// Check if closed first
	if rb.closed.Load() {
		return false
	}

	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)

	// Check if full: write position catches up with read position (difference equals capacity)
	if writePos-readPos >= rb.capacity {
		return false // buffer is full
	}

	// Write data
	rb.buffer[writePos&rb.mask] = item

	// Update write position (atomic operation ensures visibility)
	atomic.StoreUint64(&rb.writePos, writePos+1)

	// Notify consumer of new data (non-blocking)
	select {
	case rb.notifyC <- struct{}{}:
	default:
	}

	return true
}

// Push writes data by producer with blocking
// Blocks and waits until space is available if buffer is full
// Returns false if ring buffer is closed, true on success
// This approach ensures data is not lost, forming a backpressure mechanism
func (rb *RingSPSC[T]) Push(item T) bool {
	for {
		// Check if closed
		if rb.closed.Load() {
			return false
		}

		// Try non-blocking write first
		if rb.TryPush(item) {
			return true
		}

		// Buffer is full, yield CPU time slice to give consumer a chance to execute
		// In SPSC mode, this is more efficient than using an additional notification channel
		// Because the consumer is definitely processing data and space will be available soon
		runtime.Gosched()
	}
}

// TryPop reads data by consumer (non-blocking)
// Returns the data and a flag indicating success
func (rb *RingSPSC[T]) TryPop() (T, bool) {
	readPos := atomic.LoadUint64(&rb.readPos)
	writePos := atomic.LoadUint64(&rb.writePos)

	// Check if empty
	if readPos == writePos {
		var zero T
		return zero, false // buffer is empty
	}

	// Read data
	item := rb.buffer[readPos&rb.mask]

	// Update read position
	atomic.StoreUint64(&rb.readPos, readPos+1)

	return item, true
}

// Pop reads data by consumer with blocking
// Blocks and waits until data arrives if buffer is empty
// Returns the data and success flag (false indicates closed AND buffer is empty)
// When closed, will drain all remaining data before returning false
func (rb *RingSPSC[T]) Pop() (T, bool) {
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
func (rb *RingSPSC[T]) Length() uint64 {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return writePos - readPos
}

// Capacity returns the capacity of the buffer
func (rb *RingSPSC[T]) Capacity() uint64 {
	return rb.capacity
}

// IsEmpty reports whether the ring buffer contains no items.
// It suffers from a race condition where the result may be stale immediately after return.
func (rb *RingSPSC[T]) IsEmpty() bool {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return readPos == writePos
}

// IsFull reports whether the ring buffer is at capacity.
// It suffers from a race condition where the result may be stale immediately after return.
func (rb *RingSPSC[T]) IsFull() bool {
	writePos := atomic.LoadUint64(&rb.writePos)
	readPos := atomic.LoadUint64(&rb.readPos)
	return writePos-readPos >= rb.capacity
}

// Close closes the ring buffer and wakes up blocked consumers
// After close, producers cannot push new data, but consumers will drain remaining data
// Safe to call multiple times
func (rb *RingSPSC[T]) Close() {
	if rb.closed.CompareAndSwap(false, true) {
		close(rb.notifyC)
		rb.wg.Wait()
	}
}

// IsClosed returns whether the ring buffer is closed
func (rb *RingSPSC[T]) IsClosed() bool {
	return rb.closed.Load()
}

// start launches the consumer goroutine
func (rb *RingSPSC[T]) start() {
	rb.wg.Add(1)
	go rb.worker()
}

// worker is the goroutine that processes messages from the ring buffer
func (rb *RingSPSC[T]) worker() {
	defer rb.wg.Done()

	// Recover from potential panic in processing logic to avoid crashing the process
	defer func() {
		if r := recover(); r != nil {
			log.Printf("RingSPSC worker panic: %v", r)
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

// roundUpToPowerOfTwo rounds a number up to the nearest power of 2
func roundUpToPowerOfTwo(n uint64) uint64 {
	if n == 0 {
		return 1
	}

	// Check if already a power of 2
	if n&(n-1) == 0 {
		return n
	}

	// Find the smallest power of 2 greater than n
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++

	return n
}
