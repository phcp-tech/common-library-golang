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

package log

import (
	"os"
	"strings"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises a minimal environment so that the sync.OnceValue
// initialisation closure in Instance takes the "env.Env() != nil" branch,
// covering the log-level and write-to-file configuration paths.
func TestMain(m *testing.M) {
	dir := strings.ReplaceAll(os.TempDir(), `\`, "/")
	toml := "[app]\nname = \"logtest\"\n[app.env]\nprefix = \"TEST_\"\n[log]\nlevel = \"debug\"\nwritefile = true\npath = \"" +
		dir + "/logtest.log\"\n"

	f, err := os.CreateTemp("", "logtest-*.toml")
	if err != nil {
		os.Exit(2)
	}
	f.WriteString(toml) //nolint:errcheck
	f.Close()
	defer os.Remove(f.Name())

	if err := env.InitEnv(f.Name()); err != nil {
		os.Exit(2)
	}

	os.Exit(m.Run())
}

// -----------------------------------------------------------------------
// SetLevel – valid levels
// -----------------------------------------------------------------------

func TestSetLevel_ValidLevels(t *testing.T) {
	validLevels := []string{"error", "warn", "info", "debug", "ERROR", "WARN", "INFO", "DEBUG", "Error", "Warn"}
	for _, lvl := range validLevels {
		if err := SetLevel(lvl); err != nil {
			t.Errorf("SetLevel(%q) returned unexpected error: %v", lvl, err)
		}
	}
}

func TestSetLevel_InvalidLevel(t *testing.T) {
	invalidLevels := []string{"", "trace", "verbose", "critical", "fatal", "123"}
	for _, lvl := range invalidLevels {
		if err := SetLevel(lvl); err == nil {
			t.Errorf("SetLevel(%q) expected error, got nil", lvl)
		}
	}
}

func TestSetLevel_InvalidLevel_ErrorMessage(t *testing.T) {
	err := SetLevel("bogus")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// -----------------------------------------------------------------------
// SetLevel – level actually changes (verify via Instance().logLevel)
// -----------------------------------------------------------------------

func TestSetLevel_ChangesLevel(t *testing.T) {
	// After calling SetLevel("debug") the internal logLevel should reflect Debug.
	if err := SetLevel("debug"); err != nil {
		t.Fatalf("SetLevel(debug): %v", err)
	}
	want := "DEBUG"
	got := Instance().logLevel.Level().String()
	if got != want {
		t.Errorf("logLevel after SetLevel(debug): want %s, got %s", want, got)
	}

	if err := SetLevel("error"); err != nil {
		t.Fatalf("SetLevel(error): %v", err)
	}
	want = "ERROR"
	got = Instance().logLevel.Level().String()
	if got != want {
		t.Errorf("logLevel after SetLevel(error): want %s, got %s", want, got)
	}

	if err := SetLevel("warn"); err != nil {
		t.Fatalf("SetLevel(warn): %v", err)
	}
	want = "WARN"
	got = Instance().logLevel.Level().String()
	if got != want {
		t.Errorf("logLevel after SetLevel(warn): want %s, got %s", want, got)
	}

	if err := SetLevel("info"); err != nil {
		t.Fatalf("SetLevel(info): %v", err)
	}
	want = "INFO"
	got = Instance().logLevel.Level().String()
	if got != want {
		t.Errorf("logLevel after SetLevel(info): want %s, got %s", want, got)
	}
}

// -----------------------------------------------------------------------
// Logging functions – must not panic
// -----------------------------------------------------------------------

func TestInfo_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Info panicked: %v", r)
		}
	}()
	Info("info message")
}

func TestWarn_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Warn panicked: %v", r)
		}
	}()
	Warn("warn message")
}

func TestError_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error panicked: %v", r)
		}
	}()
	Error("error message")
}

func TestDebug_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Debug panicked: %v", r)
		}
	}()
	_ = SetLevel("debug")
	Debug("debug message")
}

func TestInfof_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Infof panicked: %v", r)
		}
	}()
	Infof("formatted %s %d", "msg", 42)
}

func TestWarnf_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Warnf panicked: %v", r)
		}
	}()
	Warnf("warn %v", "x")
}

func TestErrorf_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Errorf panicked: %v", r)
		}
	}()
	Errorf("error %v", "x")
}

func TestDebugf_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Debugf panicked: %v", r)
		}
	}()
	_ = SetLevel("debug")
	Debugf("debug %d", 1)
}

// -----------------------------------------------------------------------
// Instance – singleton is non-nil and has a non-nil Logger
// -----------------------------------------------------------------------

func TestInstance_NonNil(t *testing.T) {
	inst := Instance()
	if inst == nil {
		t.Fatal("Instance() returned nil")
	}
	if inst.Logger == nil {
		t.Fatal("Instance().Logger is nil")
	}
	if inst.logLevel == nil {
		t.Fatal("Instance().logLevel is nil")
	}
}

func TestInstance_Idempotent(t *testing.T) {
	a := Instance()
	b := Instance()
	if a != b {
		t.Error("Instance() should return the same pointer on every call")
	}
}

// -----------------------------------------------------------------------
// CloseLogFile – safe to call (must not panic)
// -----------------------------------------------------------------------

func TestCloseLogFile_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CloseLogFile panicked: %v", r)
		}
	}()
	// When no file is configured the lumberjack.Logger.Close() is a no-op; must not panic.
	CloseLogFile()
}
