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

// Package component provides env lifecycle integration for bootstrap.
package component

import (
	"embed"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
)

// Component loads the TOML config file into the koanf singleton.
// Pass an embedded FS for single-binary deployments; omit it to load from disk.
//
// Close is a no-op: the koanf singleton is a pure in-memory object with no
// resources to release.
//
// This component must be passed as the first argument to bootstrap.New() so
// that the framework guarantees it is initialised before all other components.
func Component(file string, fs ...*embed.FS) bootstrap.IComponent {
	var embedFS *embed.FS
	if len(fs) > 0 {
		embedFS = fs[0]
	}
	return bootstrap.Func("env",
		func() error { return env.InitEnv(file, embedFS) },
		nil, // no-op close: koanf holds no releasable resources
	)
}
