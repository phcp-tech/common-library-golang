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

// Package storage accesses S3-compatible object storage (Cloudflare R2, AWS S3,
// MinIO, Aliyun OSS) through one generic S3 client, so switching platforms only
// changes configuration.
package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Store is an S3 client bound to the buckets of one Config.
type Store struct {
	client *s3.Client
	cfg    Config
}

// New validates cfg and creates the S3 client. It does not contact the store;
// call Check for that.
func New(cfg Config) (*Store, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: cfg.UsePathStyle,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		// S3-compatible stores do not all accept the SDK's default request
		// checksums; send and validate them only when an operation requires it.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &Store{client: client, cfg: cfg}, nil
}

// Config returns the settings the store was created with.
func (s *Store) Config() Config { return s.cfg }

// Client returns the underlying S3 client for operations this package does
// not wrap.
func (s *Store) Client() *s3.Client { return s.client }

// Check verifies that every configured bucket is reachable with the
// credentials. It lists at most one object per bucket, which a token limited
// to object read/write is allowed to do.
func (s *Store) Check(ctx context.Context) error {
	for _, bucket := range s.cfg.buckets() {
		_, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			MaxKeys: aws.Int32(1),
		})
		if err != nil {
			return fmt.Errorf("storage: list bucket %s: %w", bucket, err)
		}
	}
	return nil
}
