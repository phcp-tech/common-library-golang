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

package component_test

import (
	"fmt"

	dbComp "github.com/phcp-tech/common-library-golang/dbgorm/sqlite/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads the SQLite path from env key db.sqlite.path during Init():
//
//	"file:app.db?cache=shared"          — file-based database
//	"file::memory:?cache=shared"        — in-memory (tests, ephemeral data)
//
// SQLite is an embedded database — Init() never fails due to network issues.
// Schema migration should be registered as a PreReady step immediately after:
//
//	bootstrap.New().
//	    Add(envComp.Component(...)).
//	    Add(logComp.Component()).
//	    Add(dbComp.Component()).
//	    PreReady(func() error {
//	        return dbgorm.AutoMigrate(ctx, dbgorm.Default(), dbgorm.MigrateOptions{Enabled: true}, &MyModel{})
//	    }).
//	    Run()
func ExampleComponent() {
	c := dbComp.Component()
	fmt.Println(c != nil)
	// Output:
	// true
}
