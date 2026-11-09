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
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestStore_PresignPut(t *testing.T) {
	s, err := New(validConfig())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	p, err := s.PresignPut(ctx, "images/2026/11/u.png", PutOptions{
		ContentType:  "image/png",
		Size:         12,
		CacheControl: "public, max-age=31536000, immutable",
		Metadata:     map[string]string{"uploader-id": "7"},
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(p.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(u.Path, "/images/2026/11/u.png") || !strings.Contains(u.Host+u.Path, "content") {
		t.Errorf("URL %s does not target key in the main bucket", p.URL)
	}
	if got := u.Query().Get("X-Amz-Expires"); got != "300" {
		t.Errorf("X-Amz-Expires = %q, want the 5-minute default", got)
	}
	want := map[string]string{
		"Content-Type":           "image/png",
		"Cache-Control":          "public, max-age=31536000, immutable",
		"X-Amz-Meta-Uploader-Id": "7",
	}
	for k, v := range want {
		if p.Headers[k] != v {
			t.Errorf("header %s = %q, want %q (all: %v)", k, p.Headers[k], v, p.Headers)
		}
	}
	for _, k := range []string{"Host", "Content-Length", "Content-Disposition"} {
		if _, ok := p.Headers[k]; ok {
			t.Errorf("headers include %s: %v", k, p.Headers)
		}
	}

	d, err := s.PresignPut(ctx, "documents/2026/11/u.pdf", PutOptions{
		ContentType: "application/pdf", Size: 1, ContentDisposition: "attachment", Expires: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Headers["Content-Disposition"] != "attachment" || !strings.Contains(d.URL, "X-Amz-Expires=60") {
		t.Errorf("attachment: url %s, headers %v", d.URL, d.Headers)
	}
	if _, ok := d.Headers["Cache-Control"]; ok {
		t.Errorf("empty CacheControl is signed: %v", d.Headers)
	}

	for _, o := range []PutOptions{{Size: 1}, {ContentType: "image/png"}} {
		if _, err := s.PresignPut(ctx, "k", o); err == nil {
			t.Errorf("%+v: want an error", o)
		}
	}
}
