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

	dbComp "github.com/phcp-tech/common-library-golang/dbsqlc/postgres/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads PostgreSQL connection parameters from env during Init():
//
//	db.host, db.port, db.name, db.schema, db.username, db.password
//
// NewPostgres performs an eager connectivity check — if the database is
// unreachable Init() returns an error immediately so bootstrap aborts startup.
//
// Schema migration should be registered as a separate step immediately after:
//
//	bootstrap.New(envComp.Component(...), logComp.Component()).
//	    Add(dbComp.Component()).
//	    Add(bootstrap.Func("migrate", func() error {
//	        db.AutoMigrate(...)
//	        return nil
//	    }, nil)).
//	    Run()
//
// postgres.InitDefault uses sync.Once — only the first Init call per process
// takes effect. In this example the database is unreachable (testdata points
// to 127.0.0.1:1) so Init returns an error.
func ExampleComponent() {
	c := dbComp.Component()
	fmt.Println(c != nil)
	// Output:
	// true
}
