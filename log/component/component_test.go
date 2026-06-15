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

// TestMain initialises the env singleton so that logComponent.Init() can read
// log.level and related keys via env.Env().
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("log/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestComponent_ReturnsNonNil(t *testing.T) {
	c := Component()
	if c == nil {
		t.Error("Component() returned nil")
	}
}

func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "log" {
		t.Errorf("Name() = %q, want %q", c.Name(), "log")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() reads log config from env
// and calls log.InitLog without error.
func TestComponent_Init_ReturnsNil(t *testing.T) {
	c := Component()
	if err := c.Init(); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}
