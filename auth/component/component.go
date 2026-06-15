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

// Package component provides Casbin lifecycle integration for bootstrap.
package component

import (
	"github.com/phcp-tech/common-library-golang/auth"
	"github.com/phcp-tech/common-library-golang/bootstrap"
)

// Component initialises the Casbin authorisation engine.
// When fromString is true, model and policy are loaded from the provided
// strings (suitable for embedded configs); when false they are treated as
// file paths.
//
// Close is a no-op: Casbin holds no resources that require explicit release.
func Component(fromString bool, model, policy string) bootstrap.IComponent {
	return bootstrap.Func("casbin",
		func() error { return auth.InitCasbin(fromString, model, policy) },
		nil, // no-op close
	)
}
