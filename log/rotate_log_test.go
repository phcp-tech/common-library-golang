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
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestMain initialises the logger with a temporary log file so that all tests
// exercise the file-logging code path (asyncWriter + lumberjack).
// CloseLogFile is called after the test run to drain the ring buffer cleanly.
func TestMain(m *testing.M) {
	f, err := os.CreateTemp("", "logtest-*.log")
	if err != nil {
		os.Exit(2)
	}
	f.Close()
	defer os.Remove(f.Name())

	InitLog(&Config{
		Level:    "debug",
		FilePath: f.Name(),
	})

	code := m.Run()

	// Flush the async ring buffer and release the log file before exit.
	CloseLogFile()
	os.Exit(code)
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
	cases := []struct {
		input string
		want  string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"error", "ERROR"},
	}
	for _, c := range cases {
		if err := SetLevel(c.input); err != nil {
			t.Fatalf("SetLevel(%q): %v", c.input, err)
		}
		got := Instance().logLevel.Level().String()
		if got != c.want {
			t.Errorf("logLevel after SetLevel(%q): want %s, got %s", c.input, c.want, got)
		}
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
// With variants – must not panic
// -----------------------------------------------------------------------

func TestInfoWith_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InfoWith panicked: %v", r)
		}
	}()
	InfoWith("structured message", "key", "value", "count", 1)
}

func TestErrorWith_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ErrorWith panicked: %v", r)
		}
	}()
	ErrorWith("structured error", "error", "something went wrong", "code", 500)
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
// InitLog – instance is initialised correctly
// -----------------------------------------------------------------------

// resetLog resets the singleton state for the duration of a single test.
// It saves and restores instance and once so other tests are unaffected.
func resetLog(t *testing.T) {
	t.Helper()
	savedInst := instance // sync.Once must not be copied — save only the instance pointer
	instance = nil
	once = sync.Once{}
	t.Cleanup(func() {
		instance = savedInst
		once = sync.Once{}
		once.Do(func() {}) // mark once as "already run" to match pre-test state
	})
}

// TestInitLog_WithFileConfig verifies that InitLog applies every Config field
// to the initialised logger instance.
func TestInitLog_WithFileConfig(t *testing.T) {
	resetLog(t)

	cfg := Config{
		Level:      "debug",
		FilePath:   filepath.Join(t.TempDir(), "test.log"),
		MaxSizeMB:  50,
		MaxBackups: 5,
		MaxAgeDays: 14,
		Compress:   true,
	}
	InitLog(&cfg)
	t.Cleanup(func() {
		if instance != nil && instance.asyncWriter != nil {
			instance.asyncWriter.Close()
		}
		if instance != nil && instance.logFile != nil {
			instance.logFile.Close() //nolint:errcheck
		}
	})

	l := Instance()
	if l.logFile == nil {
		t.Fatal("logFile should not be nil when FilePath is set")
	}
	if l.logFile.Filename != cfg.FilePath {
		t.Errorf("FilePath: want %q, got %q", cfg.FilePath, l.logFile.Filename)
	}
	if l.logFile.MaxSize != 50 {
		t.Errorf("MaxSize: want 50, got %d", l.logFile.MaxSize)
	}
	if l.logFile.MaxBackups != 5 {
		t.Errorf("MaxBackups: want 5, got %d", l.logFile.MaxBackups)
	}
	if l.logFile.MaxAge != 14 {
		t.Errorf("MaxAgeDays: want 14, got %d", l.logFile.MaxAge)
	}
	if !l.logFile.Compress {
		t.Error("Compress: want true, got false")
	}
	if l.logLevel.Level() != slog.LevelDebug {
		t.Errorf("Level: want DEBUG, got %v", l.logLevel.Level())
	}
}

// TestInstance_DefaultProducesStdout verifies that InitLog() with no arguments
// produces a logger in stdout mode at INFO level.
func TestInstance_DefaultProducesStdout(t *testing.T) {
	l := newLog(Config{}) // equivalent to what InitLog() configures internally
	if l.logFile != nil {
		t.Error("logFile should be nil when FilePath is empty (stdout mode)")
	}
	if l.asyncWriter != nil {
		t.Error("asyncWriter should be nil when FilePath is empty (stdout mode)")
	}
	if l.logLevel.Level() != slog.LevelInfo {
		t.Errorf("default level: want INFO, got %v", l.logLevel.Level())
	}
}

// -----------------------------------------------------------------------
// newLog – stdout path
// -----------------------------------------------------------------------

// TestNewLog_StdoutPath verifies that an empty FilePath produces a logger
// that writes to stdout with no asyncWriter or logFile.
func TestNewLog_StdoutPath(t *testing.T) {
	l := newLog(Config{})
	if l.asyncWriter != nil {
		t.Error("asyncWriter should be nil for stdout logging")
	}
	if l.logFile != nil {
		t.Error("logFile should be nil for stdout logging")
	}
	if l.Logger == nil {
		t.Error("Logger should not be nil")
	}
}

// -----------------------------------------------------------------------
// newLog – file path: asyncWriter and logFile are created
// -----------------------------------------------------------------------

// TestNewLog_FilePathCreatesAsyncWriter verifies that a non-empty FilePath
// creates both an asyncWriter and a logFile.
func TestNewLog_FilePathCreatesAsyncWriter(t *testing.T) {
	l := newLog(Config{FilePath: filepath.Join(t.TempDir(), "test.log")})
	t.Cleanup(func() {
		l.asyncWriter.Close()
		l.logFile.Close()
	})

	if l.asyncWriter == nil {
		t.Error("asyncWriter should not be nil for file logging")
	}
	if l.logFile == nil {
		t.Error("logFile should not be nil for file logging")
	}
}

// -----------------------------------------------------------------------
// newLog – Config fields are applied correctly
// -----------------------------------------------------------------------

func newLogWithTempFile(t *testing.T, cfg Config) *Log {
	t.Helper()
	cfg.FilePath = filepath.Join(t.TempDir(), "test.log")
	l := newLog(cfg)
	t.Cleanup(func() {
		l.asyncWriter.Close()
		l.logFile.Close()
	})
	return l
}

func TestNewLog_MaxSizeMB_Default(t *testing.T) {
	l := newLogWithTempFile(t, Config{})
	if l.logFile.MaxSize != defaultLogFileMaxSize {
		t.Errorf("MaxSize: want %d (default), got %d", defaultLogFileMaxSize, l.logFile.MaxSize)
	}
}

func TestNewLog_MaxSizeMB_Custom(t *testing.T) {
	l := newLogWithTempFile(t, Config{MaxSizeMB: 50})
	if l.logFile.MaxSize != 50 {
		t.Errorf("MaxSize: want 50, got %d", l.logFile.MaxSize)
	}
}

func TestNewLog_MaxBackups_Default(t *testing.T) {
	l := newLogWithTempFile(t, Config{})
	if l.logFile.MaxBackups != defaultLogFileMaxBackups {
		t.Errorf("MaxBackups: want %d (default), got %d", defaultLogFileMaxBackups, l.logFile.MaxBackups)
	}
}

func TestNewLog_MaxBackups_Custom(t *testing.T) {
	l := newLogWithTempFile(t, Config{MaxBackups: 5})
	if l.logFile.MaxBackups != 5 {
		t.Errorf("MaxBackups: want 5, got %d", l.logFile.MaxBackups)
	}
}

func TestNewLog_MaxAgeDays_Default(t *testing.T) {
	l := newLogWithTempFile(t, Config{})
	if l.logFile.MaxAge != defaultLogFileMaxAge {
		t.Errorf("MaxAge: want %d (default), got %d", defaultLogFileMaxAge, l.logFile.MaxAge)
	}
}

func TestNewLog_MaxAgeDays_Custom(t *testing.T) {
	l := newLogWithTempFile(t, Config{MaxAgeDays: 30})
	if l.logFile.MaxAge != 30 {
		t.Errorf("MaxAge: want 30, got %d", l.logFile.MaxAge)
	}
}

func TestNewLog_Compress_DefaultFalse(t *testing.T) {
	l := newLogWithTempFile(t, Config{})
	if l.logFile.Compress != false {
		t.Error("Compress: want false (default), got true")
	}
}

func TestNewLog_Compress_True(t *testing.T) {
	l := newLogWithTempFile(t, Config{Compress: true})
	if !l.logFile.Compress {
		t.Error("Compress: want true, got false")
	}
}

// -----------------------------------------------------------------------
// newLog – Level is resolved correctly
// -----------------------------------------------------------------------

func TestNewLog_Level(t *testing.T) {
	cases := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"DEBUG", slog.LevelDebug},  // case-insensitive
		{"INFO", slog.LevelInfo},
		{"",       slog.LevelInfo},  // empty → default info
		{"invalid", slog.LevelInfo}, // unrecognised → default info
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l := newLog(Config{Level: c.input})
			if got := l.logLevel.Level(); got != c.want {
				t.Errorf("Level(%q): want %v, got %v", c.input, c.want, got)
			}
		})
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
	// CloseLogFile is idempotent; calling it mid-test must not panic.
	CloseLogFile()
}
