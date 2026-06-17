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
)

// ExampleApp_PreReady shows how to register inline setup functions that run
// in the startup sequence before the HTTP server starts.
// Each PreReady call appends a step; steps run in registration order relative
// to Add/AddParallel calls. A non-nil error aborts startup identically to a
// failed component Init().
//
//	bootstrap.New(envComp.Component(...), logComp.Component()).
//	    AddParallel(dbComp.Component()).
//	    PreReady(migrate).        // runs after DB is ready
//	    PreReady(initServices).   // runs after migrate
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(handler)).
//	    Run()
func ExampleApp_PreReady() {
	app := bootstrap.New(
		bootstrap.Func("env", nil, nil),
		bootstrap.Func("log", nil, nil),
	).
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
// Multiple PostReady calls are useful when separate concerns each need a
// post-startup hook independently:
//
//	bootstrap.New(envComp.Component(...), logComp.Component()).
//	    AddParallel(dbComp.Component()).
//	    Add(ginComp.Component(mount)).
//	    Add(httpComp.Component(handler)).
//	    PostReady(func() { slog.Info("server ready", "addr", ":8080") }).
//	    PostReady(func() { discovery.Register(serviceID) }).
//	    Run()
func ExampleApp_PostReady() {
	app := bootstrap.New(
		bootstrap.Func("env", nil, nil),
		bootstrap.Func("log", nil, nil),
	).
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
