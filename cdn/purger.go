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

// Package cdn purges CDN cache entries by URL. Callers use only the Purger
// interface; New picks the provider: "cloudflare" or "none" (no purge, the
// cache expires by its TTL). It purges by URL only, not by tag or host name:
// those are plan features of some CDNs and have no common equivalent.
//
// It depends on the standard library only.
package cdn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Purger removes cached copies of URLs from the CDN.
type Purger interface {
	PurgeURLs(ctx context.Context, urls []string) error
}

// noop purges nothing: cached entries expire by their TTL.
type noop struct{}

func (noop) PurgeURLs(context.Context, []string) error { return nil }

// maxURLsPerRequest is Cloudflare's limit of files in one purge request; a
// URL purged for one Origin counts as one file.
const maxURLsPerRequest = 30

type cloudflare struct {
	apiBase string
	zoneID  string
	token   string
	origins []string
	client  *http.Client
}

// Option configures New.
type Option func(*options)

type options struct {
	origins []string
}

// WithOrigins also purges, for every URL, the copies cached for each of these
// request origins, e.g. "https://example.com". A CDN keys a cached object by
// its Origin header too, and a purge by URL alone clears only the copy cached
// without an Origin. Pass the origins that read the URLs cross-origin with
// fetch: the CORS origins of the bucket. A URL loaded by <img> or by a page
// navigation has no Origin and needs none.
func WithOrigins(origins ...string) Option {
	return func(o *options) { o.origins = append(o.origins, origins...) }
}

// New returns the Purger of provider: "cloudflare" or "none".
func New(provider, zoneID, token string, opts ...Option) (Purger, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	switch provider {
	case "none":
		return noop{}, nil
	case "cloudflare":
		if zoneID == "" || token == "" {
			return nil, errors.New("cdn: cloudflare needs a zone id and a purge token")
		}
		for _, origin := range o.origins {
			if err := checkOrigin(origin); err != nil {
				return nil, err
			}
		}
		return &cloudflare{
			apiBase: "https://api.cloudflare.com/client/v4",
			zoneID:  zoneID,
			token:   token,
			origins: o.origins,
			client:  &http.Client{Timeout: 10 * time.Second},
		}, nil
	default:
		return nil, fmt.Errorf("cdn: unknown provider %q (cloudflare|none)", provider)
	}
}

// checkOrigin accepts a browser Origin value: scheme://host[:port], nothing more.
func checkOrigin(origin string) error {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return fmt.Errorf("cdn: origin %q is not scheme://host[:port]", origin)
	}
	return nil
}

// entry is one file of the purge request: the URL, or the URL and the Origin
// of the cached copy to clear. label names it in errors.
type entry struct {
	file  any
	label string
}

// entries returns every URL once without an Origin, then once per origin.
func (c *cloudflare) entries(urls []string) []entry {
	out := make([]entry, 0, len(urls)*(len(c.origins)+1))
	for _, u := range urls {
		out = append(out, entry{file: u, label: u})
		for _, origin := range c.origins {
			out = append(out, entry{
				file:  map[string]any{"url": u, "headers": map[string]string{"Origin": origin}},
				label: u + " (Origin " + origin + ")",
			})
		}
	}
	return out
}

// PurgeURLs calls POST /zones/{zone_id}/purge_cache in batches. It tries every
// batch and returns all failures.
func (c *cloudflare) PurgeURLs(ctx context.Context, urls []string) error {
	all := c.entries(urls)
	var errs []error
	for start := 0; start < len(all); start += maxURLsPerRequest {
		if err := c.purge(ctx, all[start:min(start+maxURLsPerRequest, len(all))]); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *cloudflare) purge(ctx context.Context, batch []entry) error {
	files := make([]any, len(batch))
	urls := make([]string, len(batch))
	for i, e := range batch {
		files[i], urls[i] = e.file, e.label
	}
	body, _ := json.Marshal(map[string][]any{"files": files})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/zones/"+c.zoneID+"/purge_cache", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("purge %v: %w", urls, err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	var out struct {
		Success bool `json:"success"`
	}
	if res.StatusCode != http.StatusOK || json.Unmarshal(data, &out) != nil || !out.Success {
		return fmt.Errorf("purge %v: HTTP %d: %s", urls, res.StatusCode, data)
	}
	return nil
}
