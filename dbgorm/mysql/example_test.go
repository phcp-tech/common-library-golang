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

// package mysql_test demonstrates the public API from a caller's perspective.
package mysql_test

import (
	"fmt"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/mysql"
)

// ExampleNewMySQL shows how to open a MySQL GORM database with default pool settings.
// GORM's MySQL driver pings the server during Open — a non-nil error is returned
// when the server is unreachable.
func ExampleNewMySQL() {
	_, err := mysql.NewMySQL(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable, GORM pings on Open
	// Output:
	// true
}

// ExampleInitDefault shows the default-instance pattern.
// Call InitDefault once at application startup; dbgorm.Default() returns
// the shared *gorm.DB for the process. Unlike sync.Once, SetDefault can
// be called again if the connection needs to be replaced.
func ExampleInitDefault() {
	err := mysql.InitDefault(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err == nil)
	fmt.Println(dbgorm.Default() != nil)
}

