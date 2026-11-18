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
	"os"
	"strings"
	"testing"
)

const placeholderTOML = `
[app.env]
prefix = "PHTEST_"

[db]
host = "localhost"
password = "TOBE_REPLACED"

[email]
host = "TOBE_REPLACED"
from.name = ""

[cors]
origins = ["https://a.test", "TOBE_REPLACED"]
`

func TestInitEnv_PlaceholderFails(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, placeholderTOML)
	defer os.Remove(path)

	err := InitEnv(path)
	if err == nil {
		t.Fatal("InitEnv: want an error for placeholder values")
	}
	// All keys, sorted, each with its environment variable.
	want := "cors.origins (PHTEST_CORS_ORIGINS), db.password (PHTEST_DB_PASSWORD), email.host (PHTEST_EMAIL_HOST)"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("err = %v\nwant it to contain %q", err, want)
	}
	if strings.Contains(err.Error(), "db.host") || strings.Contains(err.Error(), "from.name") {
		t.Errorf("err names keys that are set or empty: %v", err)
	}
	if Env() != nil {
		t.Error("Env() must stay nil when InitEnv fails")
	}
}

func TestInitEnv_EnvironmentReplacesPlaceholder(t *testing.T) {
	resetSingleton(t)
	path := writeTempTOML(t, placeholderTOML)
	defer os.Remove(path)
	t.Setenv("PHTEST_DB_PASSWORD", "secret")
	t.Setenv("PHTEST_EMAIL_HOST", "smtp.example.test")
	t.Setenv("PHTEST_CORS_ORIGINS", "https://b.test")

	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}
	if got := Env().String("email.host"); got != "smtp.example.test" {
		t.Errorf("email.host = %q", got)
	}
}

func TestRequireStrings(t *testing.T) {
	resetSingleton(t)
	if err := RequireStrings("a"); err == nil {
		t.Error("before InitEnv: want an error")
	}

	path := writeTempTOML(t, minimalTOML+`
[email]
host = "smtp.example.test"
port = "587"
username = "  "
from.name = ""
`)
	defer os.Remove(path)
	if err := InitEnv(path); err != nil {
		t.Fatalf("InitEnv: %v", err)
	}

	if err := RequireStrings("email.host", "email.port"); err != nil {
		t.Errorf("set keys: err = %v", err)
	}
	if err := RequireStrings(); err != nil {
		t.Errorf("no keys: err = %v", err)
	}

	// A placeholder set after InitEnv (e.g. by a test) is also rejected.
	if err := Env().Set("email.password", Placeholder); err != nil {
		t.Fatal(err)
	}
	err := RequireStrings("email.host", "email.username", "email.from.name", "email.password", "email.missing")
	if err == nil {
		t.Fatal("want an error")
	}
	want := "email.username (TEST_EMAIL_USERNAME), email.from.name (TEST_EMAIL_FROM_NAME), email.password (TEST_EMAIL_PASSWORD), email.missing (TEST_EMAIL_MISSING)"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("err = %v\nwant it to contain %q", err, want)
	}
	if strings.Contains(err.Error(), "email.host") {
		t.Errorf("err names a key that is set: %v", err)
	}
}

func TestVarName(t *testing.T) {
	resetSingleton(t)
	if got := VarName("email.from.address"); got != "EMAIL_FROM_ADDRESS" {
		t.Errorf("before InitEnv: VarName = %q", got)
	}
	path := writeTempTOML(t, minimalTOML)
	defer os.Remove(path)
	if err := InitEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := VarName("email.from.address"); got != "TEST_EMAIL_FROM_ADDRESS" {
		t.Errorf("VarName = %q", got)
	}
}
