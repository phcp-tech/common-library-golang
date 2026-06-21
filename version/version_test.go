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

package version

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
)

func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("version tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestGet_FieldsFromEnv(t *testing.T) {
	v := Get()
	if v.Name != "test-app" {
		t.Errorf("Version.Name = %q, want %q", v.Name, "test-app")
	}
	if v.Version != "1.0.0" {
		t.Errorf("Version.Version = %q, want %q", v.Version, "1.0.0")
	}
	if v.Environment != "test" {
		t.Errorf("Version.Environment = %q, want %q", v.Environment, "test")
	}
}

func TestGet_GoVersionPopulated(t *testing.T) {
	v := Get()
	if v.GoVersion == "" {
		t.Error("Version.GoVersion must not be empty; debug.ReadBuildInfo() should succeed in tests")
	}
}

func TestGet_NonZeroStruct(t *testing.T) {
	v := Get()
	if v == (Version{}) {
		t.Error("Get() returned zero-value Version struct")
	}
}
