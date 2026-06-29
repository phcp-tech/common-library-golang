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

	dbComp "github.com/phcp-tech/common-library-golang/dbsqlc/mysql/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads MySQL connection parameters from env during Init():
//
//	db.host, db.port, db.name, db.username, db.password
//
// Unlike PostgreSQL, sql.Open is lazy — Init() always returns nil regardless of
// whether the server is reachable. Use a PreReady step for an eager ping if needed:
//
//	bootstrap.New().
//	    Add(envComp.Component(...)).
//	    Add(logComp.Component()).
//	    Add(dbComp.Component()).
//	    PreReady(func() error {
//	        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	        defer cancel()
//	        return mysql.Default().PingContext(ctx)
//	    }).
//	    Run()
//
// mysql.InitDefault uses sync.Once — only the first Init call per process takes effect.
func ExampleComponent() {
	c := dbComp.Component()
	fmt.Println(c != nil)
	// Output:
	// true
}
