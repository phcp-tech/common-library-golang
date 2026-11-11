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

package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	if p, err := New("none", "", ""); err != nil || p == nil {
		t.Fatalf("none: %v", err)
	}
	if _, err := New("cloudflare", "zone", ""); err == nil {
		t.Error("cloudflare without a token: want an error")
	}
	if _, err := New("cloudfront", "zone", "token"); err == nil {
		t.Error("unknown provider: want an error")
	}
}

func TestCloudflare_PurgeInBatches(t *testing.T) {
	var mu sync.Mutex
	var batches [][]string
	fail := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/zones/z1/purge_cache" || r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("request %s %s, auth %q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
		}
		var body struct{ Files []string }
		json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		batches = append(batches, body.Files)
		mu.Unlock()
		if fail {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"success":false}`)
			return
		}
		fmt.Fprint(w, `{"success":true}`)
	}))
	defer srv.Close()

	p, err := New("cloudflare", "z1", "tok")
	if err != nil {
		t.Fatal(err)
	}
	p.(*cloudflare).apiBase = srv.URL

	urls := make([]string, 65)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://content.test/articles/2026/11/S/r%d.md", i+1)
	}
	if err := p.PurgeURLs(context.Background(), urls); err != nil {
		t.Fatal(err)
	}
	if len(batches) != 3 || len(batches[0]) != 30 || len(batches[1]) != 30 || len(batches[2]) != 5 || batches[2][4] != urls[64] {
		t.Fatalf("batch sizes = %d %v", len(batches), batches)
	}

	fail = true
	if err := p.PurgeURLs(context.Background(), urls[:1]); err == nil {
		t.Fatal("rejected purge: want an error")
	}
	if err := p.PurgeURLs(context.Background(), nil); err != nil {
		t.Fatalf("no URLs: %v", err)
	}
}
