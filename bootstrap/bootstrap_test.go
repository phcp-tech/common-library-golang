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

	if len(app.phases) != 1 {
		t.Fatalf("len(phases) = %d, want 1", len(app.phases))
	}
	p := app.phases[0]
	if p.parallel {
		t.Error("Add() should produce parallel=false phase")
	}
	if len(p.comps) != 2 {
		t.Errorf("len(comps) = %d, want 2", len(p.comps))
	}
}

func TestApp_AddParallel_AppendsParallelPhase(t *testing.T) {
	app := New(Func("env", nil, nil), Func("log", nil, nil))
	app.AddParallel(Func("a", nil, nil), Func("b", nil, nil), Func("c", nil, nil))

	if len(app.phases) != 1 {
		t.Fatalf("len(phases) = %d, want 1", len(app.phases))
	}
	p := app.phases[0]
	if !p.parallel {
		t.Error("AddParallel() should produce parallel=true phase")
	}
	if len(p.comps) != 3 {
		t.Errorf("len(comps) = %d, want 3", len(p.comps))
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
