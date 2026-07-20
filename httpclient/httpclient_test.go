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

package httpclient_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phcp-tech/common-library-golang/httpclient"
)

// -----------------------------------------------------------------------
// HttpClient — construction
// -----------------------------------------------------------------------

func TestNewHttpClient_ReturnsNonNil(t *testing.T) {
	if httpclient.NewHttpClient() == nil {
		t.Fatal("NewHttpClient() returned nil")
	}
}

func TestNewHttpClient_ClientNonNil(t *testing.T) {
	if httpclient.NewHttpClient().Client() == nil {
		t.Error("Client() returned nil underlying resty.Client")
	}
}

// -----------------------------------------------------------------------
// HttpClient — HTTP methods
// -----------------------------------------------------------------------

func TestHttpClient_Get_CorrectMethodAndToken(t *testing.T) {
	var gotMethod, gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{RetryMax: 1})
	resp, err := cli.Get(ts.URL, "mytoken", nil)
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode(), http.StatusOK)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if !strings.Contains(gotAuth, "mytoken") {
		t.Errorf("Authorization header %q should contain token", gotAuth)
	}
}

func TestHttpClient_Post_CorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{RetryMax: 1})
	resp, err := cli.Post(ts.URL, "", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("Post error = %v", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode(), http.StatusCreated)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
}

func TestHttpClient_Put_CorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{RetryMax: 1})
	resp, err := cli.Put(ts.URL, "", nil)
	if err != nil {
		t.Fatalf("Put error = %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode(), http.StatusOK)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
}

func TestHttpClient_Delete_CorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{RetryMax: 1})
	resp, err := cli.Delete(ts.URL, "", nil)
	if err != nil {
		t.Fatalf("Delete error = %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode(), http.StatusNoContent)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
}

// -----------------------------------------------------------------------
// HttpClient — RetryOnServerErrors
// -----------------------------------------------------------------------

// TestNewHttpClient_RetryOnServerErrors_Disabled_DoesNotRetry5xx documents
// resty's default retry condition: a well-formed 503 response is not a Go
// error, so RetryMax never fires for it unless RetryOnServerErrors is set.
func TestNewHttpClient_RetryOnServerErrors_Disabled_DoesNotRetry5xx(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{
		RetryMax:      3,
		RetryWaitTime: 10 * time.Millisecond,
	})
	resp, err := cli.Get(ts.URL, "", nil)
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if resp.StatusCode() != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusServiceUnavailable)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("server saw %d requests, want 1 (no retry on 5xx by default)", got)
	}
}

// TestNewHttpClient_RetryOnServerErrors_Enabled_Retries5xxThenSucceeds verifies
// that RetryOnServerErrors retries a 503 until the server recovers.
func TestNewHttpClient_RetryOnServerErrors_Enabled_Retries5xxThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{
		RetryMax:            3,
		RetryWaitTime:       10 * time.Millisecond,
		RetryOnServerErrors: true,
	})
	resp, err := cli.Get(ts.URL, "", nil)
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusOK)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("server saw %d requests, want 3 (2 failed attempts + 1 success)", got)
	}
}

// TestNewHttpClient_RetryOnServerErrors_Enabled_RetriesOn429 verifies 429 is
// treated the same as a 5xx.
func TestNewHttpClient_RetryOnServerErrors_Enabled_RetriesOn429(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cli := httpclient.NewHttpClient(httpclient.Config{
		RetryMax:            2,
		RetryWaitTime:       10 * time.Millisecond,
		RetryOnServerErrors: true,
	})
	resp, err := cli.Get(ts.URL, "", nil)
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusOK)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("server saw %d requests, want 2", got)
	}
}

// flakyTransport simulates a transport-level failure (connection reset,
// timeout, etc.) for its first N round trips, then succeeds.
type flakyTransport struct {
	failures int
	calls    atomic.Int32
}

func (f *flakyTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if int(f.calls.Add(1)) <= f.failures {
		return nil, errors.New("simulated transport failure")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

// TestNewHttpClient_RetryOnServerErrors_StillRetriesTransportErrors verifies
// that enabling RetryOnServerErrors doesn't drop resty's original
// transport-error retry behavior — the custom condition explicitly ORs in
// err != nil precisely to preserve it (registering any custom retry
// condition replaces, rather than augments, resty's default one).
func TestNewHttpClient_RetryOnServerErrors_StillRetriesTransportErrors(t *testing.T) {
	ft := &flakyTransport{failures: 2}
	cli := httpclient.NewHttpClient(httpclient.Config{
		RetryMax:            3,
		RetryWaitTime:       10 * time.Millisecond,
		RetryOnServerErrors: true,
	})
	cli.Client().SetTransport(ft)

	resp, err := cli.Get("http://example.invalid", "", nil)
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusOK)
	}
	if got := ft.calls.Load(); got != 3 {
		t.Errorf("transport saw %d round trips, want 3 (2 failures + 1 success)", got)
	}
}
