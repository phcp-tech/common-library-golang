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

package component

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the entire component test suite.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("token/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	if Component() == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name.
func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "token" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "token")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() succeeds when
// jwt.access.secretcode is configured with a real value (testdata/config.toml).
func TestComponent_Init_ReturnsNil(t *testing.T) {
	if err := Component().Init(); err != nil {
		t.Errorf("Component().Init() = %v, want nil", err)
	}
}

// TestComponent_Close_DoesNotPanic verifies Close() is a safe no-op.
func TestComponent_Close_DoesNotPanic(t *testing.T) {
	Component().Close()
}
