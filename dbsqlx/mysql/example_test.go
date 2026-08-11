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

	"github.com/phcp-tech/common-library-golang/dbsqlx/mysql"
	"github.com/phcp-tech/common-library-golang/dto"
)

// ExampleNewMySQL shows how to open a MySQL-backed *sqlx.DB.
// Open verifies connectivity with an eager ping — a non-nil error is returned
// immediately when the server is unreachable.
func ExampleNewMySQL() {
	_, err := mysql.NewMySQL(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable, Open pings eagerly
	// Output:
	// true
}

// ExampleInitDefault shows the default-instance pattern.
// Call InitDefault once at application startup; dbsqlx.Default() returns
// the shared *sqlx.DB for the process.
func ExampleInitDefault() {
	err := mysql.InitDefault(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable
}

// ExampleDSN shows how to build a go-sql-driver/mysql connection string.
func ExampleDSN() {
	dsn, err := mysql.DSN(&mysql.Config{
		Host:     "localhost",
		Port:     "3306",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err == nil)
	fmt.Println(dsn)
	// Output:
	// true
	// user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=UTC
}

// ExampleZhSortSql shows how to build an ORDER BY clause that
// approximates pinyin ordering for a Chinese-text column, then append it
// directly to a raw SELECT statement alongside dbsqlx.PageSql for pagination.
// This is MySQL-specific — see ZhSortSql's doc comment for why.
func ExampleZhSortSql() {
	para := dto.PageParameter{Sort: "name", Direction: "ASC"}
	query := "SELECT * FROM products" + mysql.ZhSortSql(&para)
	fmt.Println(query)
	// Output:
	// SELECT * FROM products ORDER BY CONVERT(name USING gbk) ASC
}
