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

// Package component provides object storage lifecycle integration for bootstrap.
package component

import (
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/storage"
)

// loadFromEnv reads the storage settings from the koanf env singleton and
// initialises the package-level default Store. Configuration keys:
//
//	storage.endpoint          — S3 API endpoint (required)
//	storage.region            — region, "auto" for R2 (required)
//	storage.path.style        — true for MinIO
//	storage.access.key.id     — access key id (required)
//	storage.secret.access.key — secret access key (required)
//	storage.bucket            — main bucket (required)
//	storage.archive.bucket    — optional second bucket
//	storage.public.base.url   — optional public URL prefix of storage.bucket
//
// A required key that is empty fails Init with the name of the environment
// variable that has to be set (env.RequireStrings). env.InitEnv already rejects
// keys left at "TOBE_REPLACED", so credentials are never read from the file.
func loadFromEnv() error {
	cfg, err := configFromEnv()
	if err != nil {
		slog.Error("Load storage config failed", "error", err)
		return err
	}
	if err := storage.InitDefault(cfg); err != nil {
		slog.Error("Create storage client failed", "error", err)
		return err
	}
	slog.Info("Storage client created", "bucket", cfg.Bucket, "archiveBucket", cfg.ArchiveBucket)
	return nil
}

func configFromEnv() (storage.Config, error) {
	if err := env.RequireStrings("storage.endpoint", "storage.region", "storage.access.key.id",
		"storage.secret.access.key", "storage.bucket"); err != nil {
		return storage.Config{}, err
	}
	k := env.Env()
	cfg := storage.Config{UsePathStyle: k.Bool("storage.path.style")}
	for _, f := range []struct {
		key string
		dst *string
	}{
		{"storage.endpoint", &cfg.Endpoint},
		{"storage.region", &cfg.Region},
		{"storage.access.key.id", &cfg.AccessKeyID},
		{"storage.secret.access.key", &cfg.SecretAccessKey},
		{"storage.bucket", &cfg.Bucket},
		{"storage.archive.bucket", &cfg.ArchiveBucket},
		{"storage.public.base.url", &cfg.PublicBaseURL},
	} {
		*f.dst = k.String(f.key)
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")
	return cfg, nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent. Close is a no-op:
// the S3 client holds no resources that need releasing.
func Component() bootstrap.IComponent {
	return bootstrap.Func("storage", loadFromEnv, nil)
}
