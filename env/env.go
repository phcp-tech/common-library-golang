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
	"sync"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
)

// Global ENV singleton; set exactly once by InitEnv.
var (
	instance *koanf.Koanf
	once     sync.Once
)

// Env returns the global koanf configuration instance loaded by InitEnv.
// It returns nil if InitEnv has not been called or if initialization failed.
func Env() *koanf.Koanf {
	return instance
}

// InitEnv initializes the global environment configuration by loading the specified
// TOML config file and then merging any matching OS environment variables (using the
// prefix defined by app.env.prefix in the config file). If configFS is provided and
// non-nil, the file is read from the embedded filesystem instead of the real filesystem.
// Only the first call takes effect; subsequent calls return the result of the first
// initialization without reloading.
func InitEnv(configFile string, configFS ...*embed.FS) error {
	var err error
	once.Do(func() {
		//1. Load config file
		k := koanf.New(".")
		if configFS != nil && configFS[0] != nil {
			//for on execute binary file, use fs.Provider instead of file.Provider
			if err = k.Load(fs.Provider(configFS[0], configFile), toml.Parser()); err != nil {
				return
			}
		} else {
			if err = k.Load(file.Provider(configFile), toml.Parser()); err != nil {
				return
			}
		}

		// only print variables read from file, don't print variables read from environment
		fmt.Println("Application read environment variables:")
		k.Print()

		// Read log level from environment, then merge into config file variables.
		prefix := "TM_"
		prefix = k.String("app.env.prefix")
		k.Load(env.Provider(prefix, ".", func(s string) string { //nolint:errcheck
			return strings.ReplaceAll(strings.ToLower(
				strings.TrimPrefix(s, prefix)), "_", ".")
		}), nil)

		// this Print only for test
		//fmt.Println("\nMerged environment variables:")
		//k.Print()

		instance = k
	})
	return err
}
