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

// Tests must be run with: go test -gcflags=all=-l ./dbsqlx/clickhouse-native/...
package clickhousenative

import (
	"context"
	"errors"
	"testing"

	"github.com/phcp-tech/common-library-golang/health"
)

// TestHealthChecker_ReturnsNonNil verifies that HealthChecker returns a non-nil Checker.
func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

// TestHealthChecker_NoDefault verifies StatusUnhealthy when Default() is nil.
func TestHealthChecker_NoDefault(t *testing.T) {
	resetSingleton(t)

	result := HealthChecker()(context.Background())

	if result.Name != "clickhouse-native" {
		t.Errorf("result.Name = %q, want %q", result.Name, "clickhouse-native")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_PingSuccess verifies StatusHealthy when Default() pings successfully.
func TestHealthChecker_PingSuccess(t *testing.T) {
	resetSingleton(t)
	instance = &mockConn{pingErr: nil}

	result := HealthChecker()(context.Background())

	if result.Status != health.StatusHealthy {
		t.Errorf("result.Status = %d, want StatusHealthy (%d)", result.Status, health.StatusHealthy)
	}
}

// TestHealthChecker_PingFailure verifies StatusUnhealthy when Ping returns an error.
func TestHealthChecker_PingFailure(t *testing.T) {
	resetSingleton(t)
	instance = &mockConn{pingErr: errors.New("connection refused")}

	result := HealthChecker()(context.Background())

	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}
