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
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
)

// Golbal ENV
var _ENV *koanf.Koanf

// Env returns the global koanf configuration instance loaded by InitEnv.
func Env() *koanf.Koanf {
	return _ENV
}

// InitEnv initializes the global environment configuration by loading the specified
// TOML config file and then merging any matching OS environment variables (using the
// prefix defined by app.env.prefix in the config file). If configFS is provided and
// non-nil, the file is read from the embedded filesystem instead of the real filesystem.
func InitEnv(configFile string, configFS ...*embed.FS) error {
	//1. Load config file
	_ENV = koanf.New(".")
	if configFS != nil && configFS[0] != nil {
		//for on execute binary file, use fs.Provider instead of file.Provider
		if err := _ENV.Load(fs.Provider(configFS[0], configFile), toml.Parser()); err != nil {
			return err
		}
	} else {
		if err := _ENV.Load(file.Provider(configFile), toml.Parser()); err != nil {
			return err
		}
	}

	// only print variables read from file, don't print variables read from environment
	fmt.Println("Application read environment variables:")
	_ENV.Print()

	// Read log level from environment, then merge into config file variables.
	prefix := "TM_"
	prefix = _ENV.String("app.env.prefix")
	_ENV.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)

	// this Print only for test
	//fmt.Println("\nMerged environment variables:")
	//_ENV.Print()
	return nil
}
