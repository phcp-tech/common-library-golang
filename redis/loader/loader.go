package loader

import (
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/redis"
)

func LoadDefault() error {
	redisCfg := &redis.Config{
		Clusters: strings.Split(env.Env().String("redis.clusters"), ","),
		DB:       env.Env().Int("redis.database"),
		Password: env.Env().String("redis.password"),
	}

	if err := redis.InitDefault(redisCfg); err != nil {
		slog.Info("Connect Redis failed: " + err.Error())
		return err
	}
	slog.Info("Connect Redis successfully.")
	return nil
}
