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

package dbgorm

import (
	"context"

	"github.com/phcp-tech/common-library-golang/health"
)

// HealthChecker returns a [health.Checker] that verifies database connectivity
// by pinging the default GORM database.
//
// The Checker is dialect-agnostic — it works with any driver (MySQL, PostgreSQL,
// SQLite, …) that was used to initialise [Default].
// Reports [health.StatusUnhealthy] when Default() is nil or when the ping fails;
// [health.StatusHealthy] when reachable.
func HealthChecker() health.Checker {
	return func(ctx context.Context) health.Result {
		status := health.StatusUnhealthy
		if db := Default(); db != nil {
			if sqlDB, err := db.DB(); err == nil {
				if sqlDB.PingContext(ctx) == nil {
					status = health.StatusHealthy
				}
			}
		}
		return health.Result{Name: "database", Status: status}
	}
}
