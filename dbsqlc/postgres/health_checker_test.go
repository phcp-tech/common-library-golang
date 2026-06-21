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

package postgres_test

import (
	"context"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
	"github.com/phcp-tech/common-library-golang/health"
)

// TestHealthChecker_ReturnsChecker verifies that HealthChecker returns a non-nil Checker.
func TestHealthChecker_ReturnsChecker(t *testing.T) {
	checker := postgres.HealthChecker()
	if checker == nil {
		t.Error("HealthChecker() returned nil")
	}
}

// TestHealthChecker_NoPool verifies that the Checker reports StatusUnhealthy when
// the default pool has not been initialised (Default() returns nil).
func TestHealthChecker_NoPool(t *testing.T) {
	checker := postgres.HealthChecker()
	result := checker(context.Background())

	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}
