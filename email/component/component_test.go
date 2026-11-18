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
	"strings"
	"testing"

	"github.com/phcp-tech/common-library-golang/email"
	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the component test suite.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("email/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// setKey overrides one config key for the duration of the test.
func setKey(t *testing.T, key, value string) {
	t.Helper()
	k := env.Env()
	old := k.String(key)
	if err := k.Set(key, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = k.Set(key, old) })
}

func TestComponent_Name(t *testing.T) {
	if got := Component().Name(); got != "email" {
		t.Errorf("Component().Name() = %q, want %q", got, "email")
	}
}

// TestComponent_Init creates the default sender without contacting the server.
func TestComponent_Init(t *testing.T) {
	if err := Component().Init(); err != nil {
		t.Fatalf("Init() = %v", err)
	}
	if email.Default() == nil {
		t.Fatal("email.Default() = nil after Init")
	}
	cfg, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	want := email.Config{Host: "smtp.example.test", Port: 587, Username: "user", Password: "secret",
		FromAddress: "noreply@example.test", FromName: "Toolmaker"}
	if cfg != want {
		t.Errorf("config = %+v", cfg)
	}
}

func TestComponent_InitFailsOnBadConfig(t *testing.T) {
	for _, c := range []struct{ key, value, want string }{
		{"email.host", "", "TEST_EMAIL_HOST"},
		{"email.password", " ", "TEST_EMAIL_PASSWORD"},
		{"email.from.address", "", "TEST_EMAIL_FROM_ADDRESS"},
		{"email.port", "abc", "not a number"},
		{"email.port", "0", "not a valid port number"},
		{"email.from.address", "noreply", "not a valid email address"},
	} {
		t.Run(c.key+"="+c.value, func(t *testing.T) {
			setKey(t, c.key, c.value)
			err := Component().Init()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Init() = %v, want an error containing %q", err, c.want)
			}
		})
	}
}
