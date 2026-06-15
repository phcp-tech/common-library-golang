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

package component_test

import (
	"testing"

	authComp "github.com/phcp-tech/common-library-golang/auth/component"
)

const testModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

const testPolicy = `p, alice, data1, read`

func TestComponent_ReturnsNonNil(t *testing.T) {
	c := authComp.Component(true, testModel, testPolicy)
	if c == nil {
		t.Error("Component() returned nil")
	}
}

func TestComponent_Name(t *testing.T) {
	c := authComp.Component(true, testModel, testPolicy)
	if c.Name() != "casbin" {
		t.Errorf("Name() = %q, want %q", c.Name(), "casbin")
	}
}

// TestComponent_Init_SuccessWithValidStrings verifies that Init() succeeds when
// valid Casbin model and policy strings are provided.
func TestComponent_Init_SuccessWithValidStrings(t *testing.T) {
	c := authComp.Component(true, testModel, testPolicy)
	if err := c.Init(); err != nil {
		t.Errorf("Init() = %v, want nil", err)
	}
}
