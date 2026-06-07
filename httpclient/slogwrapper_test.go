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

// Internal test file — package httpclient (not httpclient_test) so that the
// unexported slogWrapper type and struct fields are accessible.
package httpclient

import (
	"log/slog"
	"testing"
	"time"
)

// -----------------------------------------------------------------------
// slogWrapper — resty.Logger adapter (unexported)
// Resty calls Errorf/Warnf/Debugf with printf-style format strings;
// slogWrapper formats them via fmt.Sprintf before passing to slog.
// -----------------------------------------------------------------------

func TestSlogWrapper_Errorf_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{Logger: slog.Default()}
	w.Errorf("error %s %d", "msg", 42)
}

func TestSlogWrapper_Warnf_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{Logger: slog.Default()}
	w.Warnf("warn %s", "msg")
}

func TestSlogWrapper_Debugf_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{Logger: slog.Default()}
	w.Debugf("debug %d", 1)
}

// -----------------------------------------------------------------------
// Config.resolve — all fields (internal access via unexported struct fields)
// -----------------------------------------------------------------------

func TestConfig_Resolve_DefaultRetryMax(t *testing.T) {
	cli := NewHttpClient()
	if got := cli.httpClient.RetryCount; got != defaultRetryMax {
		t.Errorf("default RetryMax: want %d, got %d", defaultRetryMax, got)
	}
}

func TestConfig_Resolve_CustomRetryMax(t *testing.T) {
	cli := NewHttpClient(Config{RetryMax: 5})
	if got := cli.httpClient.RetryCount; got != 5 {
		t.Errorf("custom RetryMax: want 5, got %d", got)
	}
}

func TestConfig_Resolve_DefaultTimeout(t *testing.T) {
	cli := NewHttpClient()
	if got := cli.httpClient.GetClient().Timeout; got != defaultTimeout {
		t.Errorf("default Timeout: want %v, got %v", defaultTimeout, got)
	}
}

func TestConfig_Resolve_CustomTimeout(t *testing.T) {
	want := 5 * time.Second
	cli := NewHttpClient(Config{Timeout: want})
	if got := cli.httpClient.GetClient().Timeout; got != want {
		t.Errorf("custom Timeout: want %v, got %v", want, got)
	}
}

func TestConfig_Resolve_DefaultRetryWaitTime(t *testing.T) {
	cli := NewHttpClient()
	if got := cli.httpClient.RetryWaitTime; got != defaultRetryWaitTime {
		t.Errorf("default RetryWaitTime: want %v, got %v", defaultRetryWaitTime, got)
	}
}

func TestConfig_Resolve_CustomRetryWaitTime(t *testing.T) {
	want := 3 * time.Second
	cli := NewHttpClient(Config{RetryWaitTime: want})
	if got := cli.httpClient.RetryWaitTime; got != want {
		t.Errorf("custom RetryWaitTime: want %v, got %v", want, got)
	}
}

func TestConfig_Resolve_DefaultRetryMaxWaitTime(t *testing.T) {
	cli := NewHttpClient()
	if got := cli.httpClient.RetryMaxWaitTime; got != defaultRetryMaxWaitTime {
		t.Errorf("default RetryMaxWaitTime: want %v, got %v", defaultRetryMaxWaitTime, got)
	}
}

func TestConfig_Resolve_CustomRetryMaxWaitTime(t *testing.T) {
	want := 45 * time.Second
	cli := NewHttpClient(Config{RetryMaxWaitTime: want})
	if got := cli.httpClient.RetryMaxWaitTime; got != want {
		t.Errorf("custom RetryMaxWaitTime: want %v, got %v", want, got)
	}
}
