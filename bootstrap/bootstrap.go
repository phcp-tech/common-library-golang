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

// Package bootstrap orchestrates the sequential initialization and LIFO cleanup
// of application components at startup.
//
// # Component registration order
//
// The env component MUST be the first Add() call, and the log component MUST
// be the second Add() call. This is a hard convention:
//
//   - Every component reads configuration via env.Env() inside Init(), so env
//     must be initialized before any other component.
//   - go's slog package has a usable default instance from program start, so
//     Init() failures at any stage — including before log.Init() — are captured
//     by slog and written to stderr via the default handler.
//   - env.Close() is a no-op, so LIFO naturally makes log.Close() the last
//     meaningful shutdown operation.
//
// Basic usage:
//
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)). // MUST be first
//	    Add(logComp.Component()).                              // MUST be second
//	    AddParallel(dbComp.Component()).
//	    PreReady(migrate).
//	    PreReady(initServices).
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(func() http.Handler { return router })).
//	    PostReady(func() { slog.Info("server ready") }).
//	    Run()
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/phcp-tech/common-library-golang/shutdown"
	"golang.org/x/sync/errgroup"
)

// IComponent is the lifecycle unit managed by [App].
// Init is called during startup in registration order.
// Close is called during shutdown in LIFO order and must never fail.
type IComponent interface {
	// Name returns the component name used in log messages.
	Name() string

	// Init performs initialization. A non-nil error aborts startup,
	// triggers rollback of already-started components, and exits the process.
	Init() error

	// Close releases resources. Called during shutdown in LIFO order.
	// Must never panic or block indefinitely.
	Close()
}

// funcComp wraps a pair of functions as an IComponent.
type funcComp struct {
	name  string
	init  func() error
	close func()
}

func (f *funcComp) Name() string { return f.name }
func (f *funcComp) Init() error {
	if f.init != nil {
		return f.init()
	}
	return nil
}
func (f *funcComp) Close() {
	if f.close != nil {
		f.close()
	}
}

// Func creates an IComponent from an init/close function pair.
// close may be nil when no cleanup is needed.
func Func(name string, init func() error, close func()) IComponent {
	return &funcComp{name: name, init: init, close: close}
}

// phase holds a group of components with their concurrency mode.
type phase struct {
	comps    []IComponent
	parallel bool
}

// stepKind distinguishes component phases from pre-ready hooks in the step list.
type stepKind int

const (
	stepPhase    stepKind = iota // Add / AddParallel — has Init + Close lifecycle
	stepPreReady                 // PreReady — inline function, Init only, no Close
)

// step is a single entry in the startup sequence.
type step struct {
	kind  stepKind
	phase phase        // valid when kind == stepPhase
	fn    func() error // valid when kind == stepPreReady
}

// App is the startup orchestrator. It manages the ordered initialization
// and LIFO cleanup of application components.
//
// All components participate in the same LIFO close stack.
// Because env.Close() is a no-op, log.Close() is the last meaningful close.
type App struct {
	steps        []step
	postReadyFns []func()
}

// New creates an empty App.
//
// The first Add() call MUST register the env component, and the second Add()
// call MUST register the log component. See package-level documentation for
// the rationale.
func New() *App {
	return &App{}
}

// Add appends one or more components to the sequential startup chain.
// Components are initialized in registration order and closed in LIFO order.
func (a *App) Add(cs ...IComponent) *App {
	a.steps = append(a.steps, step{kind: stepPhase, phase: phase{comps: cs, parallel: false}})
	return a
}

// AddParallel appends a group of components that initialize concurrently.
// All components in the group must succeed before the next phase begins.
// On shutdown, the group's components are closed concurrently as a unit.
func (a *App) AddParallel(cs ...IComponent) *App {
	a.steps = append(a.steps, step{kind: stepPhase, phase: phase{comps: cs, parallel: true}})
	return a
}

// PreReady registers a setup function that runs inline in the startup sequence,
// in registration order relative to Add/AddParallel calls.
// A non-nil error aborts startup identically to a failed Init().
// Multiple calls accumulate; functions run in registration order.
// PreReady steps have no Close and do not participate in LIFO shutdown.
func (a *App) PreReady(fn func() error) *App {
	a.steps = append(a.steps, step{kind: stepPreReady, fn: fn})
	return a
}

// PostReady registers a callback invoked after all steps succeed and before
// the process blocks waiting for an OS signal.
// Multiple calls accumulate; callbacks run in registration order.
func (a *App) PostReady(fn func()) *App {
	a.postReadyFns = append(a.postReadyFns, fn)
	return a
}

// Run initializes all components in registration order, invokes PostReady
// callbacks, waits for a shutdown signal, then closes all components in
// LIFO order.
//
// All Init failures are reported via slog. Go's default slog handler writes
// to stderr, so failures before the log component is initialized are still
// captured. After log.Init() replaces the default handler, subsequent failures
// use the configured output.
//
// On Init failure: rolls back already-initialized components in LIFO order
// and exits with code 1.
// Panics in the main goroutine are recovered, written to stderr, and exit
// with code 2.
func (a *App) Run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "bootstrap: panic in main: %v\nstack: %s\n",
				r, string(debug.Stack()))
			os.Exit(2)
		}
	}()

	var closeStack []phase

	failWith := func(name string, err error) {
		slog.Error("Bootstrap init failed", "component", name, "error", err)
		closeAll(closeStack)
		os.Exit(1)
	}

	for _, s := range a.steps {
		switch s.kind {
		case stepPhase:
			if s.phase.parallel {
				eg, _ := errgroup.WithContext(context.Background())
				for _, c := range s.phase.comps {
					eg.Go(func() error { return c.Init() })
				}
				if err := eg.Wait(); err != nil {
					failWith("parallel-phase", err)
				}
			} else {
				for _, c := range s.phase.comps {
					if err := c.Init(); err != nil {
						failWith(c.Name(), err)
					}
				}
			}
			closeStack = append(closeStack, s.phase)
		case stepPreReady:
			if err := s.fn(); err != nil {
				failWith("pre-ready", err)
			}
		}
	}

	for _, fn := range a.postReadyFns {
		fn()
	}

	shutdown.Wait()

	// Single LIFO close covers all components including env and log.
	// Close order: custom components → log → env (env.Close is a no-op).
	closeAll(closeStack)
}

// closeAll closes phases in LIFO order. Parallel phases close concurrently.
func closeAll(phases []phase) {
	for i := len(phases) - 1; i >= 0; i-- {
		p := phases[i]
		if p.parallel {
			var wg sync.WaitGroup
			for _, c := range p.comps {
				wg.Add(1)
				go func(comp IComponent) {
					defer wg.Done()
					comp.Close()
				}(c)
			}
			wg.Wait()
		} else {
			for j := len(p.comps) - 1; j >= 0; j-- {
				p.comps[j].Close()
			}
		}
	}
}
