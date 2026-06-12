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

package loader

import (
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/redis"
)

func LoadFromEnv() error {
	redisCfg := &redis.Config{
		Clusters: strings.Split(env.Env().String("redis.clusters"), ","),
		DB:       env.Env().Int("redis.database"),
		Password: env.Env().String("redis.password"),
	}

	if err := redis.InitDefault(redisCfg); err != nil {
		slog.Info("Connect Redis failed: " + err.Error())
		return err
	}
	slog.Info("Connect Redis successfully")
	return nil
}
