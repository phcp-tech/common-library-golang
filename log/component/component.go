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

// Package component provides log lifecycle integration for bootstrap.
package component

import (
	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/log"
)

type logComponent struct{}

func (l *logComponent) Name() string { return "log" }

// Init reads log configuration from env and initialises the logging system.
// Configuration keys:
//
//	log.level           — log level (default "info")
//	log.file.path       — file path; empty means stdout only
//	log.file.max.size   — max file size in MB
//	log.file.max.backups — number of backup files to retain
//	log.file.max.age    — max age in days
//	log.file.compress   — whether to compress rotated files
func (l *logComponent) Init() error {
	cfg := &log.Config{
		Level: env.Env().String("log.level"),
	}
	if env.Env().String("log.file.path") != "" {
		cfg.FilePath = env.Env().String("log.file.path")
		cfg.MaxSizeMB = env.Env().Int("log.file.max.size")
		cfg.MaxBackups = env.Env().Int("log.file.max.backups")
		cfg.MaxAgeDays = env.Env().Int("log.file.max.age")
		cfg.Compress = env.Env().Bool("log.file.compress")
	}
	log.InitLog(cfg)
	log.Info("Initial environment config and log successfully")
	return nil
}

// Close flushes the RingMPSC async write buffer and closes any open log file.
// This component must be passed as the second argument to bootstrap.New() so
// that the framework guarantees it is closed absolutely last — after all custom
// components — ensuring cleanup log messages are captured before process exit.
func (l *logComponent) Close() {
	log.Info("Log file has been closed, application exit")
	log.Close()
}

// Component initialises the logging system from env configuration.
// It reads log.level, log.file.path and related keys during Init().
//
// No configuration is accepted at construction time to prevent accidental
// calls to env.Env() before the env component has been initialised.
func Component() bootstrap.IComponent {
	return &logComponent{}
}
