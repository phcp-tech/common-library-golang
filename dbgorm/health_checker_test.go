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

package dbgorm_test

import (
	"context"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/health"
)

func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if dbgorm.HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

func TestHealthChecker_NoDefault(t *testing.T) {
	prev := dbgorm.Default()
	dbgorm.SetDefault(nil)
	defer dbgorm.SetDefault(prev)

	result := dbgorm.HealthChecker()(context.Background())
	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_WithLiveDB covers the StatusHealthy path via a real SQLite DB.
func TestHealthChecker_WithLiveDB(t *testing.T) {
	prev := dbgorm.Default()
	dbgorm.SetDefault(openLocalDB(t))
	defer dbgorm.SetDefault(prev)

	result := dbgorm.HealthChecker()(context.Background())
	if result.Status != health.StatusHealthy {
		t.Errorf("result.Status = %d, want StatusHealthy (%d)", result.Status, health.StatusHealthy)
	}
}
