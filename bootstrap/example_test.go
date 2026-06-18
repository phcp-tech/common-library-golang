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

package bootstrap_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	envComp "github.com/phcp-tech/common-library-golang/env/component"
	logComp "github.com/phcp-tech/common-library-golang/log/component"
)

// ExampleNew shows the complete bootstrap chain.
//
// Registration order convention — the first two Add() calls are reserved:
//   - 1st Add: env component  (MUST be first — all other Init() calls read env)
//   - 2nd Add: log component  (MUST be second — log.Close() is last via LIFO)
//
// All subsequent Add / AddParallel / PreReady calls are order-independent
// relative to each other (subject to their own logical dependencies).
//
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)). // 1st — env
//	    Add(logComp.Component()).                              // 2nd — log
//	    AddParallel(dbComp.Component()).
//	    PreReady(migrate).
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(handler)).
//	    PostReady(func() { slog.Info("server ready") }).
//	    Run()
func ExampleNew() {
	app := bootstrap.New().
		Add(envComp.Component("config/app.toml")). // 1st — env
		Add(logComp.Component())                   // 2nd — log
	fmt.Println(app != nil)
	// Output:
	// true
}

// ExampleApp_PreReady shows how to register inline setup functions that run
// in the startup sequence before the HTTP server starts.
// Each PreReady call appends a step that runs in registration order relative
// to Add/AddParallel calls. A non-nil error aborts startup identically to a
// failed component Init().
//
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)). // 1st — env
//	    Add(logComp.Component()).                              // 2nd — log
//	    AddParallel(dbComp.Component()).
//	    PreReady(migrate).          // runs after DB is ready
//	    PreReady(initServices).     // runs after migrate
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(handler)).
//	    Run()
func ExampleApp_PreReady() {
	app := bootstrap.New().
		Add(envComp.Component("config/app.toml")). // 1st — env
		Add(logComp.Component()).                   // 2nd — log
		PreReady(func() error { return nil }).
		PreReady(func() error { return nil })
	fmt.Println(app != nil)
	// Output:
	// true
}

// ExampleApp_PostReady shows how to register multiple post-startup callbacks.
// Each PostReady call appends a callback; all run in registration order after
// every step succeeds and before the process blocks waiting for an OS signal.
//
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)). // 1st — env
//	    Add(logComp.Component()).                              // 2nd — log
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(handler)).
//	    PostReady(func() { slog.Info("server ready", "addr", ":8080") }).
//	    PostReady(func() { discovery.Register(serviceID) }).
//	    Run()
func ExampleApp_PostReady() {
	app := bootstrap.New().
		Add(envComp.Component("config/app.toml")). // 1st — env
		Add(logComp.Component()).                   // 2nd — log
		PostReady(func() { fmt.Println("first") }).
		PostReady(func() { fmt.Println("second") }).
		PostReady(func() { fmt.Println("third") })
	fmt.Println(app != nil)
	// Output:
	// true
}

// ExampleFunc shows how to create a simple IComponent from a pair of functions.
// close may be nil when no cleanup is needed.
func ExampleFunc() {
	c := bootstrap.Func("my-service",
		func() error {
			// initialise service
			return nil
		},
		func() {
			// release resources
		},
	)
	fmt.Println(c.Name())
	fmt.Println(c.Init())
	// Output:
	// my-service
	// <nil>
}
