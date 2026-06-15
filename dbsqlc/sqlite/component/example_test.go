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

	sqliteComp "github.com/phcp-tech/common-library-golang/dbsqlc/sqlite/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// The SQLite database path is read from env key db.sqlite.path during Init():
//   - ":memory:" for in-memory databases (tests, ephemeral data)
//   - "file:app.db?_journal_mode=WAL&_foreign_keys=on" for file-based databases
//
// Close() calls db.Close() on the default connection.
//
//	bootstrap.New(envComp.Component(...), logComp.Component()).
//	    AddParallel(sqliteComp.Component()).
//	    Run()
func ExampleComponent() {
	c := sqliteComp.Component()
	fmt.Println(c != nil)
	// Output:
	// true
}
