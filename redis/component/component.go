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

// Package component provides Redis lifecycle integration for bootstrap.
package component

import (
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/redis"
)

// loadFromEnv reads Redis connection parameters from the koanf env singleton
// and initialises the package-level default Redis client.
// Configuration keys:
//
//	redis.clusters  — comma-separated node addresses
//	redis.database  — database index (ignored in cluster mode)
//	redis.password  — authentication password
func loadFromEnv() error {
	redisCfg := &redis.Config{
		Clusters: strings.Split(env.Env().String("redis.clusters"), ","),
		DB:       env.Env().Int("redis.database"),
		Password: env.Env().String("redis.password"),
	}

	if err := redis.InitDefault(redisCfg); err != nil {
		slog.Error("Connect Redis failed", "error", err)
		return err
	}
	slog.Info("Connect Redis successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default Redis client.
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"redis",
		loadFromEnv,
		func() {
			if cli := redis.Default(); cli != nil {
				cli.Close()
				slog.Info("Redis has been disconnected")
			}
		},
	)
}
