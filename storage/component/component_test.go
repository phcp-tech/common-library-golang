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

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/storage"
)

// TestMain initialises the env singleton once for the entire component test suite.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("storage/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestComponent_Name(t *testing.T) {
	if got := Component().Name(); got != "storage" {
		t.Errorf("Component().Name() = %q, want %q", got, "storage")
	}
}

// TestComponent_Init creates the default Store without contacting the endpoint.
func TestComponent_Init(t *testing.T) {
	c := Component()
	if err := c.Init(); err != nil {
		t.Fatalf("Init() = %v, want nil", err)
	}
	defer c.Close()

	s := storage.Default()
	if s == nil {
		t.Fatal("storage.Default() is nil after Init")
	}
	cfg := s.Config()
	if !cfg.UsePathStyle || cfg.Bucket != "content" || cfg.ArchiveBucket != "archive" {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if cfg.PublicBaseURL != "https://content.example.com" {
		t.Errorf("PublicBaseURL = %q, want trailing slash removed", cfg.PublicBaseURL)
	}
}

// TestConfigFromEnv_Placeholder names the environment variable to set.
func TestConfigFromEnv_Placeholder(t *testing.T) {
	k := env.Env()
	old := k.String("storage.secret.access.key")
	if err := k.Set("storage.secret.access.key", "TOBE_REPLACED"); err != nil {
		t.Fatal(err)
	}
	defer k.Set("storage.secret.access.key", old) //nolint:errcheck

	_, err := configFromEnv()
	if err == nil || !strings.Contains(err.Error(), "TEST_STORAGE_SECRET_ACCESS_KEY") {
		t.Fatalf("configFromEnv() = %v, want error naming TEST_STORAGE_SECRET_ACCESS_KEY", err)
	}
	if err := Component().Init(); err == nil {
		t.Fatal("Init() with placeholder: want error")
	}
}
