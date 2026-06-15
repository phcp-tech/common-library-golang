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

import "testing"

func TestComponent_ReturnsNonNil(t *testing.T) {
	c := Component("testdata/config.toml")
	if c == nil {
		t.Error("Component() returned nil")
	}
}

func TestComponent_Name(t *testing.T) {
	c := Component("testdata/config.toml")
	if c.Name() != "env" {
		t.Errorf("Name() = %q, want %q", c.Name(), "env")
	}
}

func TestComponent_Init_SuccessWithValidFile(t *testing.T) {
	c := Component("testdata/config.toml")
	if err := c.Init(); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}

// TestComponent_Close_IsNoOp verifies that Close() does not panic
// (the koanf singleton has no resources to release).
func TestComponent_Close_IsNoOp(t *testing.T) {
	c := Component("testdata/config.toml")
	c.Close() // must not panic
}
