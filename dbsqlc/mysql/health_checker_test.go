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

package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/phcp-tech/common-library-golang/dbsqlc/mysql"
	"github.com/phcp-tech/common-library-golang/health"
)

// TestHealthChecker_ReturnsNonNil verifies that HealthChecker returns a non-nil Checker.
func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if mysql.HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

// TestHealthChecker_NoDefault verifies StatusUnhealthy when Default() is nil.
// Skipped if the singleton was already initialised by an earlier test in this binary.
func TestHealthChecker_NoDefault(t *testing.T) {
	if mysql.Default() != nil {
		t.Skip("singleton already initialised — cannot test pre-init state")
	}

	result := mysql.HealthChecker()(context.Background())

	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_UnreachableServer verifies StatusUnhealthy when the default
// connection exists but the server cannot be reached.
// Skipped if the singleton has not been initialised yet in this binary.
func TestHealthChecker_UnreachableServer(t *testing.T) {
	if mysql.Default() == nil {
		t.Skip("singleton not yet initialised — skipping unreachable-server test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := mysql.HealthChecker()(ctx)

	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}
