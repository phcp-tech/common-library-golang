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
	"testing"
)

// testdata/config.toml is embedded so we can test the configFS code-path.
//
//go:embed testdata/config.toml
var testFS embed.FS

// minimalTOML is the smallest valid TOML config accepted by InitEnv.
// It defines the app.env.prefix key that InitEnv reads after loading the file.
const minimalTOML = `
[app]
  [app.env]
    prefix = "TEST_"

[server]
  host = "localhost"
  port = 8080

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

// TestEnvIsNilBeforeInit verifies that Env() returns nil before InitEnv is called.
// This relies on the package-level _ENV starting as nil (or being reset to nil
// via the global state — tested indirectly by the ordering of sub-tests).
func TestEnvIsNilBeforeInit(t *testing.T) {
	// Reset global state
	_ENV = nil
	if got := Env(); got != nil {
		t.Errorf("Env() before InitEnv: expected nil, got %v", got)
	}
}

// TestInitEnvSuccess verifies that InitEnv loads a TOML file without error
// and that Env() returns a non-nil instance afterwards.
func TestInitEnvSuccess(t *testing.T) {
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
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	got := Env().String("server.host")
	if got != "localhost" {
		t.Errorf("server.host: expected %q, got %q", "localhost", got)
	}
}

// TestInitEnvLoadsBoolValue verifies that a boolean value written into the TOML
// file can be read back via koanf's Bool method.
func TestInitEnvLoadsBoolValue(t *testing.T) {
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
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	got := Env().Int("server.port")
	if got != 8080 {
		t.Errorf("server.port: expected 8080, got %d", got)
	}
}

// TestInitEnvMissingFileReturnsError verifies that InitEnv returns a non-nil
// error when the specified config file does not exist.
func TestInitEnvMissingFileReturnsError(t *testing.T) {
	err := InitEnv("/nonexistent/path/config_that_does_not_exist.toml")
	if err == nil {
		t.Error("InitEnv with missing file: expected error, got nil")
	}
}

// TestInitEnvEnvPrefix verifies that the app.env.prefix key is loaded correctly,
// confirming that nested TOML sections are parsed.
func TestInitEnvEnvPrefix(t *testing.T) {
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
	if err := InitEnv("testdata/config.toml", &testFS); err != nil {
		t.Fatalf("InitEnv with embedded FS: %v", err)
	}

	e := Env()
	if e == nil {
		t.Fatal("Env() returned nil after InitEnv with embedded FS")
	}
	if got := e.String("server.host"); got != "embedded-host" {
		t.Errorf("embedded server.host: expected %q, got %q", "embedded-host", got)
	}
	if got := e.Int("server.port"); got != 7070 {
		t.Errorf("embedded server.port: expected 7070, got %d", got)
	}
	if got := e.String("app.env.prefix"); got != "EMBED_" {
		t.Errorf("embedded app.env.prefix: expected %q, got %q", "EMBED_", got)
	}
}

// TestInitEnvReinitialization verifies that calling InitEnv a second time with
// a different config replaces the previous configuration.
func TestInitEnvReinitialization(t *testing.T) {
	first := writeTempTOML(t, minimalTOML)
	defer os.Remove(first)

	const secondTOML = `
[app]
  [app.env]
    prefix = "SECOND_"

[server]
  host = "remotehost"
  port = 9090
`
	second := writeTempTOML(t, secondTOML)
	defer os.Remove(second)

	if err := InitEnv(first); err != nil {
		t.Fatalf("first InitEnv: %v", err)
	}
	if err := InitEnv(second); err != nil {
		t.Fatalf("second InitEnv: %v", err)
	}

	if got := Env().String("server.host"); got != "remotehost" {
		t.Errorf("after reinit server.host: expected %q, got %q", "remotehost", got)
	}
	if got := Env().Int("server.port"); got != 9090 {
		t.Errorf("after reinit server.port: expected 9090, got %d", got)
	}
}
