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

// package cdn_test demonstrates the public API from a caller's perspective.
package cdn_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/phcp-tech/common-library-golang/cdn"
)

// ExampleNew picks the purger from configuration. "none" purges nothing, for
// local runs without a token; "cloudflare" needs a zone id and a purge token.
func ExampleNew() {
	p, err := cdn.New("none", "", "")
	fmt.Println(err, p.PurgeURLs(context.Background(), []string{"https://example.com/a"}))

	_, err = cdn.New("cloudflare", "zone-id", "")
	fmt.Println(err)

	_, err = cdn.New("cloudfront", "zone-id", "token")
	fmt.Println(err)
	// Output:
	// <nil> <nil>
	// cdn: cloudflare needs a zone id and a purge token
	// cdn: unknown provider "cloudfront" (cloudflare|none)
}

// recorder is a test Purger: callers depend on the interface only, so a unit
// test needs no CDN.
type recorder struct {
	urls []string
	err  error
}

func (r *recorder) PurgeURLs(_ context.Context, urls []string) error {
	r.urls = append(r.urls, urls...)
	return r.err
}

// publish changes data, then purges its URLs. A purge failure is only
// reported: the cache expires by its TTL, so the change itself succeeded.
func publish(ctx context.Context, p cdn.Purger, slug string) error {
	// ... write the new content to storage ...
	if err := p.PurgeURLs(ctx, []string{"https://example.com/articles/" + slug}); err != nil {
		fmt.Println("purge failed, the cache expires by its TTL:", err)
	}
	return nil
}

// ExamplePurger uses a Purger behind an interface and treats a purge as
// best effort.
func ExamplePurger() {
	ctx := context.Background()

	ok := &recorder{}
	fmt.Println(publish(ctx, ok, "hello"), ok.urls)

	failing := &recorder{err: errors.New("HTTP 403")}
	fmt.Println(publish(ctx, failing, "hello"))
	// Output:
	// <nil> [https://example.com/articles/hello]
	// purge failed, the cache expires by its TTL: HTTP 403
	// <nil>
}
