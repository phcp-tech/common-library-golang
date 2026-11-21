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

// Package component provides email sender lifecycle integration for bootstrap.
package component

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/email"
	"github.com/phcp-tech/common-library-golang/env"
)

// loadFromEnv reads the [email] settings from the env singleton and
// initialises the package-level default sender. Configuration keys:
//
//	email.host         — SMTP server host name (required)
//	email.port         — 465 for implicit TLS, else a STARTTLS port such as 587 (required)
//	email.username     — SMTP login (required)
//	email.password     — SMTP password (required)
//	email.from.address — default sender address (required)
//	email.from.name    — display name of the default sender (optional)
//
// A missing or invalid setting fails Init, so the service does not start.
func loadFromEnv() error {
	cfg, err := configFromEnv()
	if err != nil {
		slog.Error("Load email config failed", "error", err)
		return err
	}
	if err := email.InitDefault(cfg); err != nil {
		slog.Error("Create email sender failed", "error", err)
		return err
	}
	slog.Info("Email sender created", "from", cfg.FromAddress, "fromName", cfg.FromName)
	return nil
}

func configFromEnv() (email.Config, error) {
	if err := env.RequireStrings("email.host", "email.port", "email.username", "email.password", "email.from.address"); err != nil {
		return email.Config{}, err
	}
	k := env.Env()
	get := func(key string) string { return strings.TrimSpace(k.String(key)) }
	port, err := strconv.Atoi(get("email.port"))
	if err != nil {
		return email.Config{}, fmt.Errorf("config email.port %q is not a number (%s)", get("email.port"), env.VarName("email.port"))
	}
	return email.Config{
		Host:        get("email.host"),
		Port:        port,
		Username:    get("email.username"),
		Password:    get("email.password"),
		FromAddress: get("email.from.address"),
		FromName:    get("email.from.name"),
	}, nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent. Close is a no-op:
// the sender opens a connection per message and holds none between sends.
func Component() bootstrap.IComponent {
	return bootstrap.Func("email", loadFromEnv, nil)
}
