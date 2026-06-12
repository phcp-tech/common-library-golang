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

package loader_test

import (
	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres/loader"
)

// ExampleLoadFromEnv shows the typical call-site pattern for LoadFromEnv.
// It reads PostgreSQL connection parameters from the koanf env singleton
// (keys: db.host, db.port, db.name, db.schema, db.username, db.password, …)
// and initialises the package-level default pool via postgres.InitDefault.
//
// LoadFromEnv performs an eager connectivity check: if the database is
// unreachable the error is returned immediately so the application can fail
// fast at startup rather than on the first real query.
//
// postgres.InitDefault uses sync.Once — only the first call in a process takes
// effect. Subsequent calls (including calls from other tests in this binary)
// are no-ops and return nil. For this reason the example does not assert the
// return value: its primary purpose is to document the call site pattern.
func ExampleLoadFromEnv() {
	err := loader.LoadFromEnv()
	// err is non-nil on first call with unreachable DB;
	// nil on subsequent calls (sync.Once no-op).
	_ = err
}
