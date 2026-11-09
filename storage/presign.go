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
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DefaultPresignExpiry is the lifetime of a presigned URL when PutOptions
// sets none.
const DefaultPresignExpiry = 5 * time.Minute

// PutOptions are the object headers a presigned upload signs. The client must
// send every one of them, as returned in PresignedPut.Headers.
type PutOptions struct {
	ContentType        string            // required
	Size               int64             // required: the exact body size in bytes
	CacheControl       string            // optional
	ContentDisposition string            // optional; empty: no header, browsers show the file inline
	Metadata           map[string]string // optional x-amz-meta-* values, e.g. {"uploader-id": "7"}
	Expires            time.Duration     // optional; 0: DefaultPresignExpiry
}

// PresignedPut is an upload URL and the headers the PUT request must send.
type PresignedPut struct {
	URL     string
	Headers map[string]string
}

// PresignPut returns a URL that uploads one object to key in the main bucket
// (Config.Bucket) with a plain HTTP PUT, e.g. straight from a browser. It does
// not contact the store.
func (s *Store) PresignPut(ctx context.Context, key string, o PutOptions) (PresignedPut, error) {
	if o.ContentType == "" || o.Size <= 0 {
		return PresignedPut{}, fmt.Errorf("storage: presign %s: ContentType and a positive Size are required", key)
	}
	in := &s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.Bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(o.ContentType),
		ContentLength: aws.Int64(o.Size),
		Metadata:      o.Metadata,
	}
	if o.CacheControl != "" {
		in.CacheControl = aws.String(o.CacheControl)
	}
	if o.ContentDisposition != "" {
		in.ContentDisposition = aws.String(o.ContentDisposition)
	}
	expires := o.Expires
	if expires <= 0 {
		expires = DefaultPresignExpiry
	}
	req, err := s3.NewPresignClient(s.client).PresignPutObject(ctx, in, s3.WithPresignExpires(expires))
	if err != nil {
		return PresignedPut{}, fmt.Errorf("storage: presign %s: %w", key, err)
	}
	// Browsers set Host and Content-Length themselves and refuse them by hand.
	headers := map[string]string{}
	for name, values := range req.SignedHeader {
		name = http.CanonicalHeaderKey(name)
		if name == "Host" || name == "Content-Length" || len(values) == 0 {
			continue
		}
		headers[name] = values[0]
	}
	return PresignedPut{URL: req.URL, Headers: headers}, nil
}
