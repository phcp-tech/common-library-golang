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

// package storage_test demonstrates the public API from a caller's perspective.
package storage_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/phcp-tech/common-library-golang/storage"
)

// ExampleNew shows how to create a Store for Cloudflare R2.
// The caller reads configuration from env (or any other source) at the
// composition root — this package has no dependency on env. New does not
// contact the store.
func ExampleNew() {
	s, err := storage.New(storage.Config{
		Endpoint:        "https://account-id.r2.cloudflarestorage.com",
		Region:          "auto",
		AccessKeyID:     "access-key-id",
		SecretAccessKey: "secret-access-key",
		Bucket:          "zhishi-content-dev",
		ArchiveBucket:   "zhishi-archive-dev", // optional
		PublicBaseURL:   "https://dev.content.zhishi.io",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s.Config().Bucket)
	// Output:
	// zhishi-content-dev
}

// ExampleNew_minio shows the only difference for MinIO: path-style URLs.
func ExampleNew_minio() {
	s, err := storage.New(storage.Config{
		Endpoint:        "http://127.0.0.1:9000",
		Region:          "us-east-1",
		UsePathStyle:    true,
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		Bucket:          "content",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s.Config().UsePathStyle)
	// Output:
	// true
}

// ExampleConfig_Validate shows the error for a missing required field.
func ExampleConfig_Validate() {
	err := storage.Config{Endpoint: "https://s3.example.com", Region: "auto"}.Validate()
	fmt.Println(err)
	// Output:
	// storage: AccessKeyID is required
}

// ExampleStore_Check verifies that every configured bucket is reachable.
// A local stand-in answers ListObjectsV2 here; in production the endpoint is
// R2, S3, MinIO or OSS.
func ExampleStore_Check() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket := strings.Trim(r.URL.Path, "/")
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintf(w, `<ListBucketResult><Name>%s</Name><KeyCount>0</KeyCount></ListBucketResult>`, bucket)
	}))
	defer srv.Close()

	s, err := storage.New(storage.Config{
		Endpoint:        srv.URL,
		Region:          "us-east-1",
		UsePathStyle:    true,
		AccessKeyID:     "id",
		SecretAccessKey: "secret",
		Bucket:          "content",
		ArchiveBucket:   "archive",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s.Check(context.Background()))
	// Output:
	// <nil>
}

// ExampleStore_Client shows how to call an S3 operation that Store does not
// wrap, for example writing an immutable public object.
func ExampleStore_Client() {
	s := storage.Default() // set by storage/component, or call storage.New
	if s == nil {
		return
	}
	_, err := s.Client().PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:       aws.String(s.Config().Bucket),
		Key:          aws.String("articles/2026/11/GeryF6t32aARmCEW/r3.md"),
		Body:         strings.NewReader("# Title\n"),
		ContentType:  aws.String("text/markdown; charset=utf-8"),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		fmt.Println(err)
	}
}
