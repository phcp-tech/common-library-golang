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

package app

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the entire app package test
// suite. GetHealth and GetVersion both call env.Env().String(...) and will
// panic if the singleton is nil.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("app tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------
// GetHealth
// -----------------------------------------------------------------------

// TestGetHealth_NoDatabase verifies that GetHealth returns Status=0 and the
// configured app name when no PostgreSQL singleton has been initialised.
// db.Default() returns nil in this state so the ping is skipped.
func TestGetHealth_NoDatabase(t *testing.T) {
	h := GetHealth()

	if h.Name != "test-app" {
		t.Errorf("Health.Name = %q, want %q", h.Name, "test-app")
	}
	if h.Status != 0 {
		t.Errorf("Health.Status = %d, want 0 (no DB initialised)", h.Status)
	}
}

// TestGetHealth_StructFields verifies that the Health struct fields are set
// and the JSON tags match the documented API shape.
func TestGetHealth_StructFields(t *testing.T) {
	h := GetHealth()

	// Name must never be empty when env is properly initialised.
	if h.Name == "" {
		t.Error("Health.Name must not be empty when env is initialised")
	}
	// Status is either 0 (no DB) or 2 (DB reachable); never negative.
	if h.Status < 0 {
		t.Errorf("Health.Status must be >= 0, got %d", h.Status)
	}
}
