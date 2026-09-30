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

// White-box tests for withHostPortRemoteAddr - it's unexported, so this
// file lives in package lambda (not lambda_test like lambda_test.go's
// black-box suite) to reach it directly.
package lambda

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// -----------------------------------------------------------------------
// withHostPortRemoteAddr — RemoteAddr normalization
// -----------------------------------------------------------------------

func TestWithHostPortRemoteAddr_BareIPv4_GetsSyntheticPort(t *testing.T) {
	got := remoteAddrSeenBy(t, "203.0.113.5")
	want := "203.0.113.5:0"
	if got != want {
		t.Errorf("RemoteAddr = %q, want %q", got, want)
	}
}

// TestWithHostPortRemoteAddr_BareIPv6_GetsBracketedAndSyntheticPort covers
// the case that motivated using net.JoinHostPort instead of a plain string
// concatenation: an unbracketed IPv6 address already contains colons, so
// net.SplitHostPort fails on it exactly like a bare IPv4 address does (for
// a different reason - "too many colons in address", not "missing port in
// address") - it must also be fixed up, and JoinHostPort adds the brackets
// IPv6 needs in host:port form ("[::1]:0"), which naive concatenation
// would not.
func TestWithHostPortRemoteAddr_BareIPv6_GetsBracketedAndSyntheticPort(t *testing.T) {
	got := remoteAddrSeenBy(t, "2001:db8::1")
	want := "[2001:db8::1]:0"
	if got != want {
		t.Errorf("RemoteAddr = %q, want %q", got, want)
	}
}

func TestWithHostPortRemoteAddr_AlreadyHasPort_IsUnchanged(t *testing.T) {
	got := remoteAddrSeenBy(t, "203.0.113.9:54321")
	want := "203.0.113.9:54321"
	if got != want {
		t.Errorf("RemoteAddr = %q, want %q (should be left untouched)", got, want)
	}
}

func TestWithHostPortRemoteAddr_BracketedIPv6WithPort_IsUnchanged(t *testing.T) {
	got := remoteAddrSeenBy(t, "[::1]:54321")
	want := "[::1]:54321"
	if got != want {
		t.Errorf("RemoteAddr = %q, want %q (should be left untouched)", got, want)
	}
}

func TestWithHostPortRemoteAddr_Empty_IsUnchanged(t *testing.T) {
	got := remoteAddrSeenBy(t, "")
	if got != "" {
		t.Errorf("RemoteAddr = %q, want %q (nothing to fix, must not synthesize one)", got, "")
	}
}

// -----------------------------------------------------------------------
// withHostPortRemoteAddr — pass-through behavior
// -----------------------------------------------------------------------

func TestWithHostPortRemoteAddr_CallsNextExactlyOnce(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	})
	wrapped := withHostPortRemoteAddr(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5"
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if calls != 1 {
		t.Errorf("next was called %d times, want 1", calls)
	}
}

func TestWithHostPortRemoteAddr_PreservesResponseAndOtherRequestFields(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/agtapi/v2/agents/update" {
			t.Errorf("URL.Path = %q, want %q", r.URL.Path, "/agtapi/v2/agents/update")
		}
		w.WriteHeader(http.StatusTeapot)
	})
	wrapped := withHostPortRemoteAddr(next)

	req := httptest.NewRequest(http.MethodPost, "/agtapi/v2/agents/update", nil)
	req.RemoteAddr = "203.0.113.5"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d - wrapper must not alter the handler's response", rec.Code, http.StatusTeapot)
	}
}

// remoteAddrSeenBy runs a request with the given RemoteAddr through
// withHostPortRemoteAddr and returns what the wrapped handler actually
// observed on r.RemoteAddr.
func remoteAddrSeenBy(t *testing.T, remoteAddr string) string {
	t.Helper()
	var got string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.RemoteAddr
	})
	wrapped := withHostPortRemoteAddr(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	wrapped.ServeHTTP(httptest.NewRecorder(), req)
	return got
}
