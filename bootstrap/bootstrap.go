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
// Env is always initialized first; Log is always initialized second and
// closed absolutely last so that all cleanup log messages are captured before
// the process exits. All other components are registered via [App.Add] or
// [App.AddParallel] and participate in LIFO cleanup.
//
// Basic usage:
//
//	bootstrap.New(
//	    env.Component("config/app.toml", &configFS),
//	    log.Component(),
//	).
//	    AddParallel(dbComp.Component()).
//	    Add(bootstrap.Func("services", initServices, nil)).
//	    Add(httpComp.Component(func() http.Handler { return router })).
//	    Run()
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/phcp-tech/common-library-golang/network"
	"github.com/phcp-tech/common-library-golang/shutdown"
	"golang.org/x/sync/errgroup"
)

// Component is the lifecycle unit managed by [App].
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

// App is the startup orchestrator. It manages the ordered initialization
// and LIFO cleanup of application components.
//
// Close order guarantee:
//  1. Add/AddParallel components in LIFO order
//  2. envComp.Close() (usually no-op)
//  3. logComp.Close() (absolutely last — flushes the async log buffer)
type App struct {
	envComp IComponent
	logComp IComponent
	phases  []phase
}

// New creates a new App with mandatory foundation components.
//
// envComp is initialized first (all other components depend on configuration).
// logComp is initialized second and closed absolutely last so that all cleanup
// log messages are captured before the process exits.
//
// Both are passed as constructor parameters rather than Add() calls to enforce
// the ordering constraint at compile time.
func New(envComp IComponent, logComp IComponent) *App {
	return &App{envComp: envComp, logComp: logComp}
}

// Add appends one or more components to the sequential startup chain.
// Components are initialized in registration order and closed in LIFO order.
func (a *App) Add(cs ...IComponent) *App {
	a.phases = append(a.phases, phase{comps: cs, parallel: false})
	return a
}

// AddParallel appends a group of components that initialize concurrently.
// All components in the group must succeed before the next phase begins.
// On shutdown, the group's components are closed concurrently as a unit.
func (a *App) AddParallel(cs ...IComponent) *App {
	a.phases = append(a.phases, phase{comps: cs, parallel: true})
	return a
}

// Run initializes all components, waits for a shutdown signal, then closes
// all components in reverse order.
//
// On Init failure: rolls back already-initialized components in LIFO order,
// flushes the log buffer, then calls os.Exit(1).
// Panics in the main goroutine are recovered, written to stderr, and exit with code 2.
func (a *App) Run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "bootstrap: panic in main: %v\nstack: %s\n",
				r, string(debug.Stack()))
			os.Exit(2)
		}
	}()

	// closeStack tracks initialized custom components for LIFO rollback/shutdown.
	var closeStack []phase

	// failWith logs the error, rolls back all started custom components,
	// then closes env and log before exiting.
	// Must only be called after logComp.Init() has succeeded.
	failWith := func(name string, err error) {
		slog.Error("Bootstrap init failed", "component", name, "error", err)
		closeAll(closeStack)
		a.envComp.Close()
		a.logComp.Close()
		os.Exit(1)
	}

	// 1. Init env — always first; use stderr because log is not yet available.
	if err := a.envComp.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap: init [%s] failed: %v\n", a.envComp.Name(), err)
		os.Exit(1)
	}

	// 2. Init log — always second; all subsequent errors go through slog.
	if err := a.logComp.Init(); err != nil {
		a.envComp.Close()
		fmt.Fprintf(os.Stderr, "bootstrap: init [%s] failed: %v\n", a.logComp.Name(), err)
		os.Exit(1)
	}

	// 3. Init custom phases in registration order.
	for _, p := range a.phases {
		if p.parallel {
			eg, _ := errgroup.WithContext(context.Background())
			for _, c := range p.comps {
				eg.Go(func() error { return c.Init() })
			}
			if err := eg.Wait(); err != nil {
				failWith("parallel-phase", err)
			}
		} else {
			for _, c := range p.comps {
				if err := c.Init(); err != nil {
					failWith(c.Name(), err)
				}
			}
		}
		closeStack = append(closeStack, p)
	}

	// 4. Log startup success.
	slog.Info("Application started", "ip", network.GetLocalIpAddress())

	// 5. Block until OS signal or shutdown.Trigger().
	shutdown.Wait()

	// 6. Close custom components in LIFO order.
	closeAll(closeStack)

	// 7. Close env (no-op for koanf singleton).
	a.envComp.Close()

	// 8. Close log absolutely last — flushes async write buffer.
	a.logComp.Close()
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
