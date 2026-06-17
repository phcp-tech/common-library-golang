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
// log.level and log.file.* keys via env.Env().
// The testdata config sets log.file.path so the file-logging branch in Init()
// is exercised. The temporary log file is removed after all tests complete.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("log/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	code := m.Run()
	os.Remove("testdata/testlog.tmp")
	os.Exit(code)
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
// The testdata config sets log.file.path, so this call also exercises the
// file-logging branch (cfg.FilePath, cfg.MaxSizeMB, cfg.MaxBackups, etc.).
func TestComponent_Init_ReturnsNil(t *testing.T) {
	c := Component()
	if err := c.Init(); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}

// TestComponent_Close_FlushesLog verifies that Close() executes its full body
// (log.Info + log.Close) without panicking.
//
// Init() is called first to ensure the log singleton is initialised.
// log.InitLog uses a singleton, so the call is idempotent if Init was
// already invoked by a prior test.
func TestComponent_Close_FlushesLog(t *testing.T) {
	c := Component()
	_ = c.Init() // idempotent — ensures log is initialised before closing
	c.Close()    // must not panic; covers log.Info + log.Close
}
