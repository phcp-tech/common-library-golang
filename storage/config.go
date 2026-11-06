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

package storage

import (
	"errors"
	"fmt"
)

// Config holds the settings of an S3-compatible object store.
// The caller is responsible for reading values from env (or any other source)
// at the composition root so this package has no dependency on env;
// storage/component does that for the storage.* keys.
//
// The fields carry no platform name: Cloudflare R2, AWS S3, MinIO and Aliyun
// OSS differ only in their values.
type Config struct {
	Endpoint        string // S3 API endpoint, e.g. https://{accountId}.r2.cloudflarestorage.com
	Region          string // "auto" for R2
	UsePathStyle    bool   // true for MinIO
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string // main bucket
	ArchiveBucket   string // optional second bucket, e.g. for private archived objects
	PublicBaseURL   string // optional public URL prefix of Bucket, without a trailing slash
}

// Validate reports the first required field that is empty, or a conflict
// between Bucket and ArchiveBucket.
func (c Config) Validate() error {
	required := []struct{ name, value string }{
		{"Endpoint", c.Endpoint},
		{"Region", c.Region},
		{"AccessKeyID", c.AccessKeyID},
		{"SecretAccessKey", c.SecretAccessKey},
		{"Bucket", c.Bucket},
	}
	for _, r := range required {
		if r.value == "" {
			return fmt.Errorf("storage: %s is required", r.name)
		}
	}
	if c.ArchiveBucket != "" && c.ArchiveBucket == c.Bucket {
		return errors.New("storage: ArchiveBucket must differ from Bucket")
	}
	return nil
}

// buckets returns the configured bucket names.
func (c Config) buckets() []string {
	if c.ArchiveBucket == "" {
		return []string{c.Bucket}
	}
	return []string{c.Bucket, c.ArchiveBucket}
}
