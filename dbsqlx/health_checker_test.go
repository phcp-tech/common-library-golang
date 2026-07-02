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

package dbsqlx_test

import (
	"context"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/health"
)

func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if dbsqlx.HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

func TestHealthChecker_NoDefault(t *testing.T) {
	prev := dbsqlx.Default()
	dbsqlx.SetDefault(nil)
	defer dbsqlx.SetDefault(prev)

	result := dbsqlx.HealthChecker()(context.Background())
	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

func TestHealthChecker_WithLiveDB(t *testing.T) {
	prev := dbsqlx.Default()
	dbsqlx.SetDefault(openTestDB(t))
	defer dbsqlx.SetDefault(prev)

	result := dbsqlx.HealthChecker()(context.Background())
	if result.Status != health.StatusHealthy {
		t.Errorf("result.Status = %d, want StatusHealthy (%d)", result.Status, health.StatusHealthy)
	}
}

func TestHealthChecker_AfterClose(t *testing.T) {
	db := openTestDB(t)
	prev := dbsqlx.Default()
	dbsqlx.SetDefault(db)
	defer dbsqlx.SetDefault(prev)

	dbsqlx.Close(db) //nolint:errcheck

	result := dbsqlx.HealthChecker()(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status after Close = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}
