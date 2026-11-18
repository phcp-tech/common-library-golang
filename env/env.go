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
	"errors"
	"fmt"
	"slices"
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
		fmt.Println("Application read environment variables.")
		//k.Print()

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

		// A setting still at the placeholder was never given its real value.
		if keys := placeholderKeys(k); len(keys) > 0 {
			err = fmt.Errorf("config keys still set to %s, set them with environment variables: %s",
				Placeholder, describeKeys(prefix, keys))
			return
		}

		instance = k
	})
	return err
}

// Placeholder is the value config/app.toml files use for settings whose real
// value is set at deployment, usually with an environment variable. InitEnv
// fails while any setting still has this value. For an optional setting, or one
// used only by some configurations, use "" in app.toml instead and check it
// where it is used, for example with RequireStrings.
const Placeholder = "TOBE_REPLACED"

// placeholderKeys returns the sorted keys whose value, or one of whose list
// elements, is Placeholder.
func placeholderKeys(k *koanf.Koanf) []string {
	var keys []string
	for key, v := range k.All() {
		if isPlaceholder(v) {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}

func isPlaceholder(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == Placeholder
	case []any:
		return slices.ContainsFunc(t, isPlaceholder)
	case []string:
		return slices.ContainsFunc(t, func(s string) bool { return strings.TrimSpace(s) == Placeholder })
	}
	return false
}

// VarName returns the environment variable that sets key: app.env.prefix plus
// the key in upper case with "." replaced by "_", e.g. email.host -> TM_EMAIL_HOST.
func VarName(key string) string {
	prefix := ""
	if instance != nil {
		prefix = instance.String("app.env.prefix")
	}
	return varName(prefix, key)
}

func varName(prefix, key string) string {
	return prefix + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

func describeKeys(prefix string, keys []string) string {
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = fmt.Sprintf("%s (%s)", key, varName(prefix, key))
	}
	return strings.Join(parts, ", ")
}

// RequireStrings checks that every key has a non-empty value that is not
// Placeholder. It reports all missing keys at once, each with the environment
// variable that sets it. Use it in a component's Init for the settings that
// component needs; InitEnv already rejects placeholders, but not empty values.
func RequireStrings(keys ...string) error {
	if instance == nil {
		return errors.New("env: config is not loaded, call InitEnv first")
	}
	var missing []string
	for _, key := range keys {
		v := strings.TrimSpace(instance.String(key))
		if v == "" || v == Placeholder {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("config keys must be set: %s", describeKeys(instance.String("app.env.prefix"), missing))
	}
	return nil
}
