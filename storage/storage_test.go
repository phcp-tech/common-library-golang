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
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{
		Endpoint:        "https://account.r2.cloudflarestorage.com",
		Region:          "auto",
		AccessKeyID:     "id",
		SecretAccessKey: "secret",
		Bucket:          "content",
		ArchiveBucket:   "archive",
	}
}

func TestConfig_Validate(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("valid config: %v", err)
	}

	noArchive := validConfig()
	noArchive.ArchiveBucket = ""
	if err := noArchive.Validate(); err != nil {
		t.Fatalf("ArchiveBucket is optional: %v", err)
	}

	noSecret := validConfig()
	noSecret.SecretAccessKey = ""
	if err := noSecret.Validate(); err == nil || !strings.Contains(err.Error(), "SecretAccessKey") {
		t.Fatalf("missing SecretAccessKey: got %v", err)
	}

	same := validConfig()
	same.ArchiveBucket = same.Bucket
	if err := same.Validate(); err == nil {
		t.Fatal("same Bucket and ArchiveBucket: want error")
	}
}

func TestNew_RejectsInvalidConfig(t *testing.T) {
	cfg := validConfig()
	cfg.Endpoint = ""
	if _, err := New(cfg); err == nil {
		t.Fatal("want error for missing Endpoint")
	}
}

// fakeS3 answers ListObjectsV2 for the given buckets (path-style) and 404 otherwise.
func fakeS3(t *testing.T, buckets ...string) *httptest.Server {
	t.Helper()
	known := map[string]bool{}
	for _, b := range buckets {
		known[b] = true
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket := strings.Trim(r.URL.Path, "/")
		w.Header().Set("Content-Type", "application/xml")
		if !known[bucket] {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<Error><Code>NoSuchBucket</Code><Message>not found</Message></Error>`))
			return
		}
		_, _ = w.Write([]byte(`<ListBucketResult><Name>` + bucket + `</Name><KeyCount>0</KeyCount><MaxKeys>1</MaxKeys><IsTruncated>false</IsTruncated></ListBucketResult>`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestStore_Check(t *testing.T) {
	srv := fakeS3(t, "content", "archive")
	cfg := validConfig()
	cfg.Endpoint = srv.URL
	cfg.UsePathStyle = true

	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Check(context.Background()); err != nil {
		t.Fatalf("Check() = %v, want nil", err)
	}

	cfg.ArchiveBucket = "missing"
	s, _ = New(cfg)
	if err := s.Check(context.Background()); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("Check() with missing bucket = %v, want error naming the bucket", err)
	}
}

func TestDefault(t *testing.T) {
	if err := InitDefault(Config{}); err == nil {
		t.Fatal("InitDefault with empty config: want error")
	}
	if err := InitDefault(validConfig()); err != nil {
		t.Fatal(err)
	}
	if Default() == nil {
		t.Fatal("Default() is nil after InitDefault")
	}
}
