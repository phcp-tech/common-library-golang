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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
