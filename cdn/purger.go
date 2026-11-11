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
	"time"
)

// Purger removes cached copies of URLs from the CDN.
type Purger interface {
	PurgeURLs(ctx context.Context, urls []string) error
}

// noop purges nothing: cached entries expire by their TTL.
type noop struct{}

func (noop) PurgeURLs(context.Context, []string) error { return nil }

// maxURLsPerRequest is Cloudflare's limit of URLs in one purge request.
const maxURLsPerRequest = 30

type cloudflare struct {
	apiBase string
	zoneID  string
	token   string
	client  *http.Client
}

// New returns the Purger of provider: "cloudflare" or "none".
func New(provider, zoneID, token string) (Purger, error) {
	switch provider {
	case "none":
		return noop{}, nil
	case "cloudflare":
		if zoneID == "" || token == "" {
			return nil, errors.New("cdn: cloudflare needs a zone id and a purge token")
		}
		return &cloudflare{
			apiBase: "https://api.cloudflare.com/client/v4",
			zoneID:  zoneID,
			token:   token,
			client:  &http.Client{Timeout: 10 * time.Second},
		}, nil
	default:
		return nil, fmt.Errorf("cdn: unknown provider %q (cloudflare|none)", provider)
	}
}

// PurgeURLs calls POST /zones/{zone_id}/purge_cache in batches. It tries every
// batch and returns all failures.
func (c *cloudflare) PurgeURLs(ctx context.Context, urls []string) error {
	var errs []error
	for start := 0; start < len(urls); start += maxURLsPerRequest {
		batch := urls[start:min(start+maxURLsPerRequest, len(urls))]
		if err := c.purge(ctx, batch); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *cloudflare) purge(ctx context.Context, urls []string) error {
	body, _ := json.Marshal(map[string][]string{"files": urls})
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
