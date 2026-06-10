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

package app

import (
	"runtime/debug"

	"github.com/phcp-tech/common-library-golang/env"
)

// Version holds application version and build metadata typically returned by
// a version API endpoint.
type Version struct {
	Name        string `json:"name,omitempty"`        // Application name from the app.name environment key.
	Version     string `json:"version,omitempty"`     // Application version from the app.version environment key.
	Environment string `json:"environment,omitempty"` // Runtime environment (e.g. production, staging) from app.env.value.
	GoVersion   string `json:"goVersion,omitempty"`   // Go toolchain version used to build the binary.
	BuildInfo   string `json:"buildInfo,omitempty"`   // Module version from the embedded build info.
}

// GetVersion returns a Version struct populated from the application environment
// configuration and the embedded Go build info. It is intended for use by API
// version endpoints; CLI processes may not have the required environment variables set.
func GetVersion() Version {
	info, _ := debug.ReadBuildInfo()
	return Version{
		Name:        env.Env().String("app.name"),
		Version:     env.Env().String("app.version"),
		Environment: env.Env().String("app.env.value"),
		GoVersion:   info.GoVersion,
		BuildInfo:   info.Main.Version,
	}
}
