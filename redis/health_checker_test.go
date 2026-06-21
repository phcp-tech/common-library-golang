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

package redis

import (
	"context"
	"testing"

	"github.com/phcp-tech/common-library-golang/health"
)

func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

// TestHealthChecker_NoDefault verifies StatusUnhealthy when Default() is nil.
func TestHealthChecker_NoDefault(t *testing.T) {
	resetDefault(t)

	result := HealthChecker()(context.Background())

	if result.Name != "redis" {
		t.Errorf("result.Name = %q, want %q", result.Name, "redis")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_UnreachableServer verifies StatusUnhealthy when the
// default client exists but the server cannot be reached.
func TestHealthChecker_UnreachableServer(t *testing.T) {
	resetDefault(t)
	InitDefault(unreachableConf) //nolint:errcheck

	ctx, cancel := ctxShort()
	defer cancel()

	result := HealthChecker()(ctx)

	if result.Name != "redis" {
		t.Errorf("result.Name = %q, want %q", result.Name, "redis")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_Success verifies StatusHealthy when the default client
// successfully pings a live Redis node (miniredis).
func TestHealthChecker_Success(t *testing.T) {
	resetDefault(t)
	InitDefault(startMini(t)) //nolint:errcheck

	result := HealthChecker()(context.Background())

	if result.Name != "redis" {
		t.Errorf("result.Name = %q, want %q", result.Name, "redis")
	}
	if result.Status != health.StatusHealthy {
		t.Errorf("result.Status = %d, want StatusHealthy (%d)", result.Status, health.StatusHealthy)
	}
}
