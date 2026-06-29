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

// package clickhouse_test demonstrates the public API from a caller's perspective.
package clickhouse_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbgorm/clickhouse"
)

// ExampleNewClickHouse shows how to open a ClickHouse GORM database.
// GORM's ClickHouse driver pings the server during Open — a non-nil error
// is returned when the server is unreachable.
func ExampleNewClickHouse() {
	_, err := clickhouse.NewClickHouse(&clickhouse.Config{
		Host:     "127.0.0.1",
		Port:     "1",
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
// the shared *gorm.DB for the process.
func ExampleInitDefault() {
	err := clickhouse.InitDefault(&clickhouse.Config{
		Host:     "127.0.0.1",
		Port:     "1",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable
}
