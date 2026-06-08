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

package vfs

import (
	"embed"
	"sync"
	"testing"
)

// resetSingleton resets the singleton state for the duration of a single test.
// sync.Once must not be copied — save only the instance pointer.
func resetSingleton(t *testing.T) {
	t.Helper()
	savedInst := instance
	instance = nil
	once = sync.Once{}
	t.Cleanup(func() {
		instance = savedInst
		once = sync.Once{}
		once.Do(func() {}) // mark once as "already run" to match pre-test state
	})
}

// -----------------------------------------------------------------------
// Default — singleton state
// -----------------------------------------------------------------------

func TestDefault_BeforeInit_IsNil(t *testing.T) {
	if Default() != nil {
		t.Skip("singleton already initialised in this process — cannot test pre-init state")
	}
}

func TestDefault_Idempotent(t *testing.T) {
	a := Default()
	b := Default()
	if a != b {
		t.Error("Default() should return the same pointer on every call")
	}
}

// -----------------------------------------------------------------------
// New — error path (no valid embedded SQLite file)
// -----------------------------------------------------------------------

func TestNew_EmptyFS_PingFails(t *testing.T) {
	// sql.Open is lazy: New with an empty FS succeeds (no connection attempt yet).
	// The pool becomes unusable only when it is actually queried.
	var emptyFS embed.FS
	db, err := New(&emptyFS)
	if err != nil {
		return // early error is also acceptable
	}
	defer db.Close()
	if err := db.Ping(); err == nil {
		t.Error("Ping should fail: config/sqlite.db does not exist in the empty VFS")
	}
}

// -----------------------------------------------------------------------
// InitDefault — error path and no-op second call
// -----------------------------------------------------------------------

func TestInitDefault_EmptyFS_PingFails(t *testing.T) {
	resetSingleton(t)
	// sql.Open is lazy: InitDefault with an empty FS succeeds (no connection attempt yet).
	// The singleton is set, but the pool becomes unusable only when queried.
	var emptyFS embed.FS
	_ = InitDefault(&emptyFS)
	db := Default()
	if db == nil {
		return // early error is also acceptable
	}
	if err := db.Ping(); err == nil {
		t.Error("Ping should fail: config/sqlite.db does not exist in the empty VFS")
	}
}

func TestInitDefault_SecondCallIsNoOp(t *testing.T) {
	// When the singleton is already set (or errored), a second call is ignored.
	// We rely on the fact that the first call in this process has already run.
	if Default() == nil {
		t.Skip("singleton not yet initialised — second-call no-op test requires a live instance")
	}
	first := Default()
	_ = InitDefault(nil) // second call must not change the instance
	if Default() != first {
		t.Error("second InitDefault call should not change the singleton")
	}
}
