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

// Package version provides application version and build metadata for API endpoints.
//
// Basic usage:
//
//	router.GET("/version", func(c *gin.Context) {
//	    c.JSON(http.StatusOK, version.Get())
//	})
package version

import (
	"runtime/debug"

	"github.com/phcp-tech/common-library-golang/env"
)

// Version holds application version and build metadata typically returned by
// a version API endpoint.
type Version struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"` // runtime environment (e.g. production, staging)
	GoVersion   string `json:"goVersion,omitempty"`   // Go toolchain version used to build the binary
	BuildInfo   string `json:"buildInfo,omitempty"`   // module version from the embedded build info
}

// Get returns a Version populated from the application environment configuration
// and the embedded Go build info. It is intended for API version endpoints; CLI
// processes may not have the required environment variables set.
func Get() Version {
	v := Version{
		Name:        env.Env().String("app.name"),
		Version:     env.Env().String("app.version"),
		Environment: env.Env().String("app.env.value"),
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		v.GoVersion = info.GoVersion
		v.BuildInfo = info.Main.Version
	}
	return v
}
