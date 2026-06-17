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

package bootstrap

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/phcp-tech/common-library-golang/shutdown"
)

// -----------------------------------------------------------------------
// Func
// -----------------------------------------------------------------------

func TestFunc_NilInitAndClose(t *testing.T) {
	c := Func("x", nil, nil)
	if c.Name() != "x" {
		t.Errorf("Name() = %q, want %q", c.Name(), "x")
	}
	if err := c.Init(); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
	c.Close() // must not panic
}

func TestFunc_InitError(t *testing.T) {
	want := errors.New("boom")
	c := Func("fail", func() error { return want }, nil)
	if got := c.Init(); !errors.Is(got, want) {
		t.Errorf("Init() = %v, want %v", got, want)
	}
}

func TestFunc_CloseIsCalled(t *testing.T) {
	var called bool
	c := Func("c", nil, func() { called = true })
	c.Close()
	if !called {
		t.Error("Close() was not called")
	}
}

// -----------------------------------------------------------------------
// New / Add / AddParallel
// -----------------------------------------------------------------------

func TestNew_NotNil(t *testing.T) {
	env := Func("env", nil, nil)
	log := Func("log", nil, nil)
	app := New(env, log)
	if app == nil {
		t.Error("New() returned nil")
	}
}

func TestApp_Add_AppendsSequentialPhase(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	c1 := Func("a", nil, nil)
	c2 := Func("b", nil, nil)
	app.Add(c1, c2)

	if len(app.steps) != 1 {
		t.Fatalf("len(steps) = %d, want 1", len(app.steps))
	}
	s := app.steps[0]
	if s.kind != stepPhase {
		t.Error("Add() should produce stepPhase step")
	}
	if s.phase.parallel {
		t.Error("Add() should produce parallel=false phase")
	}
	if len(s.phase.comps) != 2 {
		t.Errorf("len(comps) = %d, want 2", len(s.phase.comps))
	}
}

func TestApp_AddParallel_AppendsParallelPhase(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	app.AddParallel(Func("a", nil, nil), Func("b", nil, nil), Func("c", nil, nil))

	if len(app.steps) != 1 {
		t.Fatalf("len(steps) = %d, want 1", len(app.steps))
	}
	s := app.steps[0]
	if s.kind != stepPhase {
		t.Error("AddParallel() should produce stepPhase step")
	}
	if !s.phase.parallel {
		t.Error("AddParallel() should produce parallel=true phase")
	}
	if len(s.phase.comps) != 3 {
		t.Errorf("len(comps) = %d, want 3", len(s.phase.comps))
	}
}

// -----------------------------------------------------------------------
// closeAll
// -----------------------------------------------------------------------

func TestCloseAll_LIFO(t *testing.T) {
	var order []string

	makeComp := func(name string) IComponent {
		return Func(name, nil, func() { order = append(order, name) })
	}

	phases := []phase{
		{comps: []IComponent{makeComp("a")}, parallel: false},
		{comps: []IComponent{makeComp("b")}, parallel: false},
		{comps: []IComponent{makeComp("c")}, parallel: false},
	}

	closeAll(phases)

	want := []string{"c", "b", "a"}
	if len(order) != len(want) {
		t.Fatalf("close order len = %d, want %d", len(order), len(want))
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

// -----------------------------------------------------------------------
// PreReady
// -----------------------------------------------------------------------

func TestApp_PreReady_AppendsStep(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	app.PreReady(func() error { return nil })
	app.PreReady(func() error { return nil })

	if len(app.steps) != 2 {
		t.Fatalf("len(steps) = %d, want 2", len(app.steps))
	}
	for i, s := range app.steps {
		if s.kind != stepPreReady {
			t.Errorf("steps[%d].kind = %v, want stepPreReady", i, s.kind)
		}
		if s.fn == nil {
			t.Errorf("steps[%d].fn is nil", i)
		}
	}
}

func TestApp_PreReady_ReturnsApp(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	got := app.PreReady(func() error { return nil })
	if got != app {
		t.Error("PreReady() should return the same *App for chaining")
	}
}

func TestApp_PreReady_OrderedWithPhases(t *testing.T) {
	// Verifies that PreReady and Add steps are interleaved in registration order.
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	app.AddParallel(Func("db", nil, nil))
	app.PreReady(func() error { return nil })
	app.Add(Func("gin", nil, nil))

	if len(app.steps) != 3 {
		t.Fatalf("len(steps) = %d, want 3", len(app.steps))
	}
	if app.steps[0].kind != stepPhase {
		t.Error("steps[0] should be stepPhase (AddParallel)")
	}
	if app.steps[1].kind != stepPreReady {
		t.Error("steps[1] should be stepPreReady (PreReady)")
	}
	if app.steps[2].kind != stepPhase {
		t.Error("steps[2] should be stepPhase (Add)")
	}
}

// -----------------------------------------------------------------------
// PostReady
// -----------------------------------------------------------------------

func TestApp_PostReady_AppendsFn(t *testing.T) {
	var order []int
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	app.PostReady(func() { order = append(order, 1) })
	app.PostReady(func() { order = append(order, 2) })
	app.PostReady(func() { order = append(order, 3) })

	if len(app.postReadyFns) != 3 {
		t.Fatalf("len(postReadyFns) = %d, want 3", len(app.postReadyFns))
	}
	for _, fn := range app.postReadyFns {
		fn()
	}
	for i, v := range []int{1, 2, 3} {
		if order[i] != v {
			t.Errorf("order[%d] = %d, want %d", i, order[i], v)
		}
	}
}

func TestApp_PostReady_ReturnsApp(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	got := app.PostReady(func() {})
	if got != app {
		t.Error("PostReady() should return the same *App for chaining")
	}
}

// -----------------------------------------------------------------------
// Run — happy path
//
// shutdown.Trigger() is called inside a PostReady callback to unblock
// shutdown.Wait() immediately after startup, allowing Run() to complete
// within the test. Once Trigger() fires, subsequent Wait() calls in the
// same test binary return at once; tests are written so that this does
// not affect their correctness.
// -----------------------------------------------------------------------

func TestRun_SequentialPhaseAndPreReady(t *testing.T) {
	var order []string

	New(
		Func("env",
			func() error { order = append(order, "env-init"); return nil },
			func() { order = append(order, "env-close") }),
		Func("log",
			func() error { order = append(order, "log-init"); return nil },
			func() { order = append(order, "log-close") }),
	).
		Add(Func("svc",
			func() error { order = append(order, "svc-init"); return nil },
			func() { order = append(order, "svc-close") })).
		PreReady(func() error { order = append(order, "pre-ready"); return nil }).
		PostReady(func() {
			order = append(order, "post-ready")
			shutdown.Trigger()
		}).
		Run()

	want := []string{
		"env-init", "log-init",
		"svc-init", "pre-ready", "post-ready",
		"svc-close", "env-close", "log-close",
	}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestRun_LIFOCloseOrder(t *testing.T) {
	var closed []string

	New(
		Func("env", nil, func() { closed = append(closed, "env") }),
		Func("log", nil, func() { closed = append(closed, "log") }),
	).
		Add(Func("a", nil, func() { closed = append(closed, "a") })).
		Add(Func("b", nil, func() { closed = append(closed, "b") })).
		Add(Func("c", nil, func() { closed = append(closed, "c") })).
		PostReady(func() { shutdown.Trigger() }).
		Run()

	want := []string{"c", "b", "a", "env", "log"}
	if len(closed) != len(want) {
		t.Fatalf("close order = %v, want %v", closed, want)
	}
	for i, v := range want {
		if closed[i] != v {
			t.Errorf("closed[%d] = %q, want %q", i, closed[i], v)
		}
	}
}

func TestRun_ParallelPhaseAllInited(t *testing.T) {
	var count atomic.Int32

	New(Func("env", nil, nil), Func("log", nil, nil)).
		AddParallel(
			Func("a", func() error { count.Add(1); return nil }, nil),
			Func("b", func() error { count.Add(1); return nil }, nil),
			Func("c", func() error { count.Add(1); return nil }, nil),
		).
		PostReady(func() { shutdown.Trigger() }).
		Run()

	if count.Load() != 3 {
		t.Errorf("parallel init count = %d, want 3", count.Load())
	}
}

func TestRun_MultiplePostReadyCallbacks(t *testing.T) {
	var order []string

	New(Func("env", nil, nil), Func("log", nil, nil)).
		PostReady(func() { order = append(order, "first") }).
		PostReady(func() { order = append(order, "second") }).
		PostReady(func() {
			order = append(order, "third")
			shutdown.Trigger()
		}).
		Run()

	want := []string{"first", "second", "third"}
	if len(order) != len(want) {
		t.Fatalf("post-ready order = %v, want %v", order, want)
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestCloseAll_ParallelGroup_AllClosed(t *testing.T) {
	var count atomic.Int32

	makeComp := func() IComponent {
		return Func("p", nil, func() { count.Add(1) })
	}

	phases := []phase{
		{comps: []IComponent{makeComp(), makeComp(), makeComp()}, parallel: true},
	}

	closeAll(phases)

	if count.Load() != 3 {
		t.Errorf("closed count = %d, want 3", count.Load())
	}
}
