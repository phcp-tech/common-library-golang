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

package env

import (
	"embed"
	"os"
	"sync"
	"testing"
)


// testdata/config.toml is embedded so we can test the configFS code-path.
//
//go:embed testdata/config.toml
var testFS embed.FS

// minimalTOML is the smallest valid TOML config accepted by InitEnv.
// It defines the app.env.prefix key that InitEnv reads after loading the file.
const minimalTOML = `
[app.env]
prefix = "TEST_"

[server]
host.name       = "embedded-host"
host.ip.address = "127.0.0.1"
port            = 7070

[feature]
enable_cache = true
`

// writeTempTOML writes content to a temporary file and returns its path.
// The caller is responsible for removing the file.
func writeTempTOML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "env_test_*.toml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		t.Fatalf("WriteString: %v", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatalf("Close: %v", err)
	}
	return f.Name()
}

// resetSingleton resets the singleton state so each test starts from a clean slate.
// It also registers a Cleanup to reset again after the test, ensuring isolation
// even when tests run in parallel or in arbitrary order.
func resetSingleton(t *testing.T) {
	t.Helper()
	instance = nil
	once = sync.Once{}
	t.Cleanup(func() {
		instance = nil
		once = sync.Once{}
	})
}

// TestEnvIsNilBeforeInit verifies that Env() returns nil before InitEnv is called.
func TestEnvIsNilBeforeInit(t *testing.T) {
	resetSingleton(t)
	if got := Env(); got != nil {
		t.Errorf("Env() before InitEnv: expected nil, got %v", got)
	}
}

// TestInitEnvSuccess verifies that InitEnv loads a TOML file without error
// and that Env() returns a non-nil instance afterwards.
func TestInitEnvSuccess(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv returned unexpected error: %v", err)
	}

	e := Env()
	if e == nil {
		t.Fatal("Env() returned nil after successful InitEnv")
	}
}

// TestInitEnvLoadsStringValue verifies that a string value written into the TOML
// file can be read back via koanf's String method.
func TestInitEnvLoadsStringValue(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	got := Env().String("server.host.name")
	if got != "embedded-host" {
		t.Errorf("server.host.name: expected %q, got %q", "embedded-host", got)
	}
}

// TestInitEnvLoadsBoolValue verifies that a boolean value written into the TOML
// file can be read back via koanf's Bool method.
func TestInitEnvLoadsBoolValue(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	got := Env().Bool("feature.enable_cache")
	if !got {
		t.Errorf("feature.enable_cache: expected true, got false")
	}
}

// TestInitEnvLoadsIntValue verifies that an integer value written into the TOML
// file can be read back via koanf's Int method.
func TestInitEnvLoadsIntValue(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	got := Env().Int("server.port")
	if got != 7070 {
		t.Errorf("server.port: expected 7070, got %d", got)
	}
}

// TestInitEnvMissingFileReturnsError verifies that InitEnv returns a non-nil
// error when the specified config file does not exist.
func TestInitEnvMissingFileReturnsError(t *testing.T) {
	resetSingleton(t)
	err := InitEnv("/nonexistent/path/config_that_does_not_exist.toml")
	if err == nil {
		t.Error("InitEnv with missing file: expected error, got nil")
	}
}

// TestInitEnvFailedInitKeepsNilInstance verifies that after a failed InitEnv
// call Env() still returns nil.
func TestInitEnvFailedInitKeepsNilInstance(t *testing.T) {
	resetSingleton(t)

	err := InitEnv("/nonexistent/path/config_that_does_not_exist.toml")
	if err == nil {
		t.Fatal("expected error from first InitEnv, got nil")
	}
	if Env() != nil {
		t.Error("Env() should be nil after failed InitEnv")
	}
}

// TestInitEnvEnvPrefix verifies that the app.env.prefix key is loaded correctly,
// confirming that nested TOML sections are parsed.
func TestInitEnvEnvPrefix(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	prefix := Env().String("app.env.prefix")
	if prefix != "TEST_" {
		t.Errorf("app.env.prefix: expected %q, got %q", "TEST_", prefix)
	}
}

// TestInitEnvMissingKeyReturnsDefault verifies that reading a non-existent key
// returns the zero-value for the requested type (koanf's documented behaviour).
func TestInitEnvMissingKeyReturnsDefault(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	if s := Env().String("no.such.key"); s != "" {
		t.Errorf("missing string key: expected empty string, got %q", s)
	}
	if b := Env().Bool("no.such.key"); b != false {
		t.Errorf("missing bool key: expected false, got true")
	}
	if i := Env().Int("no.such.key"); i != 0 {
		t.Errorf("missing int key: expected 0, got %d", i)
	}
}

// TestInitEnvWithEmbeddedFS exercises the configFS code-path in InitEnv, using
// an embedded filesystem instead of a real file on disk.
func TestInitEnvWithEmbeddedFS(t *testing.T) {
	resetSingleton(t)
	if err := InitEnv("testdata/config.toml", &testFS); err != nil {
		t.Fatalf("InitEnv with embedded FS: %v", err)
	}

	e := Env()
	if e == nil {
		t.Fatal("Env() returned nil after InitEnv with embedded FS")
	}
	if got := e.String("server.host.name"); got != "embedded-host" {
		t.Errorf("embedded server.host.name: expected %q, got %q", "embedded-host", got)
	}
	if got := e.String("server.host.ip.address"); got != "127.0.0.1" {
		t.Errorf("embedded server.host.ip.address: expected %q, got %q", "127.0.0.1", got)
	}
	if got := e.Int("server.port"); got != 7070 {
		t.Errorf("embedded server.port: expected 7070, got %d", got)
	}
	if got := e.String("app.env.prefix"); got != "EMBED_" {
		t.Errorf("embedded app.env.prefix: expected %q, got %q", "EMBED_", got)
	}
}

// TestHostAndHostName_TOMLRejectsConflict verifies that a TOML file defining
// the same key as both a scalar (host = "…") and a dotted sub-key (host.name = "…")
// within the same table is correctly rejected by the parser.
//
// The TOML v1.0 specification disallows this: once a key is bound to a scalar
// value it cannot be reused as a table path. koanf's TOML parser enforces this
// rule, so InitEnv returns a non-nil error and Env() stays nil.
//
// Background: in older koanf versions a related issue existed where koanf's own
// dot-delimiter caused silent key collisions when loading formats that do not
// enforce such rules (e.g. YAML or env vars). In the TOML code path the parser
// itself catches the conflict before koanf's key merging is involved.
func TestHostAndHostName_TOMLRejectsConflict(t *testing.T) {
	resetSingleton(t)

	const conflictingTOML = `
[server]
host     = "embedded-host"
host.name = "embedded-host-name"
`
	path := writeTempTOML(t, conflictingTOML)
	defer os.Remove(path)

	err := InitEnv(path)
	if err == nil {
		t.Fatal("expected parse error for conflicting TOML keys (host as scalar and host.name as dotted key), got nil")
	}
	if Env() != nil {
		t.Error("Env() should remain nil after failed InitEnv")
	}
}

// TestInitEnvSecondCallIsIgnored verifies the singleton guarantee: a second call to
// InitEnv with a different config file has no effect; the first configuration is retained.
func TestInitEnvSecondCallIsIgnored(t *testing.T) {
	resetSingleton(t)
	first := writeTempTOML(t, minimalTOML)
	defer os.Remove(first)

	const secondTOML = `
[app.env]
prefix = "SECOND_"

[server]
host.name       = "remotehost"
host.ip.address = "10.0.0.1"
port            = 9090
`
	second := writeTempTOML(t, secondTOML)
	defer os.Remove(second)

	if err := InitEnv(first); err != nil {
		t.Fatalf("first InitEnv: %v", err)
	}
	// Second call must be a no-op; the singleton is already initialised.
	if err := InitEnv(second); err != nil {
		t.Fatalf("second InitEnv returned unexpected error: %v", err)
	}

	// Values must still reflect the first config file.
	if got := Env().String("server.host.name"); got != "embedded-host" {
		t.Errorf("after second InitEnv server.host.name: expected %q (first config), got %q", "embedded-host", got)
	}
	if got := Env().Int("server.port"); got != 7070 {
		t.Errorf("after second InitEnv server.port: expected 7070 (first config), got %d", got)
	}
}
