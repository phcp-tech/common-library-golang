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

package shutdown

import (
	"sync"
	"testing"
	"time"
)

// reset restores package-level state between tests.
// Internal tests (same package) can access unexported vars directly.
func reset() {
	stopCh = make(chan struct{})
	once = sync.Once{}
}

// -----------------------------------------------------------------------
// Trigger
// -----------------------------------------------------------------------

func TestTrigger_UnblocksWait(t *testing.T) {
	reset()

	done := make(chan struct{})
	go func() {
		Wait()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond) // ensure Wait is blocking
	Trigger()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not unblock after Trigger")
	}
}

func TestTrigger_Idempotent(t *testing.T) {
	reset()
	// multiple calls must not panic
	Trigger()
	Trigger()
	Trigger()
}

func TestTrigger_BeforeWait_WaitReturnsImmediately(t *testing.T) {
	reset()
	Trigger() // fire before Wait

	done := make(chan struct{})
	go func() {
		Wait()
		close(done)
	}()

	select {
	case <-done:
		// expected: Wait returns immediately
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return immediately after pre-triggered Trigger")
	}
}

// -----------------------------------------------------------------------
// Wait — concurrent safety
// -----------------------------------------------------------------------

func TestWait_MultipleCallers_AllUnblock(t *testing.T) {
	reset()

	const n = 5
	done := make(chan struct{}, n)
	for range n {
		go func() {
			Wait()
			done <- struct{}{}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	Trigger()

	timeout := time.After(2 * time.Second)
	for range n {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("not all Wait callers unblocked")
		}
	}
}
