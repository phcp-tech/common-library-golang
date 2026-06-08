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

// Internal test file — package retryable (not retryable_test) so that the
// unexported slogWrapper type and struct fields are accessible.
package retryable

import (
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// -----------------------------------------------------------------------
// slogWrapper — retryablehttp.LeveledLogger adapter (unexported)
// retryablehttp passes key-value pairs compatible with slog directly.
// -----------------------------------------------------------------------

func TestSlogWrapper_Error_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{logger: slog.Default()}
	w.Error("error msg", "key", "val")
}

func TestSlogWrapper_Info_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{logger: slog.Default()}
	w.Info("info msg", "key", "val")
}

func TestSlogWrapper_Debug_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{logger: slog.Default()}
	w.Debug("debug msg", "key", "val")
}

func TestSlogWrapper_Warn_DoesNotPanic(t *testing.T) {
	w := &slogWrapper{logger: slog.Default()}
	w.Warn("warn msg", "key", "val")
}

// -----------------------------------------------------------------------
// Config.resolve — all fields (internal access via unexported struct fields)
// -----------------------------------------------------------------------

func TestConfig_Resolve_DefaultRetryMax(t *testing.T) {
	c := NewHttpClient()
	if got := c.client.RetryMax; got != defaultRetryMax {
		t.Errorf("default RetryMax: want %d, got %d", defaultRetryMax, got)
	}
}

func TestConfig_Resolve_CustomRetryMax(t *testing.T) {
	c := NewHttpClient(Config{RetryMax: 5})
	if got := c.client.RetryMax; got != 5 {
		t.Errorf("custom RetryMax: want 5, got %d", got)
	}
}

func TestConfig_Resolve_DefaultTimeout(t *testing.T) {
	c := NewHttpClient()
	if got := c.client.HTTPClient.Timeout; got != defaultTimeout {
		t.Errorf("default Timeout: want %v, got %v", defaultTimeout, got)
	}
}

func TestConfig_Resolve_CustomTimeout(t *testing.T) {
	want := 5 * time.Second
	c := NewHttpClient(Config{Timeout: want})
	if got := c.client.HTTPClient.Timeout; got != want {
		t.Errorf("custom Timeout: want %v, got %v", want, got)
	}
}

// -----------------------------------------------------------------------
// Config.resolve — InsecureSkipVerify applied to TLS transport
// -----------------------------------------------------------------------

func TestConfig_Resolve_InsecureSkipVerify_True(t *testing.T) {
	c := NewHttpClient(Config{InsecureSkipVerify: true})
	transport, ok := c.client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("expected non-nil TLSClientConfig")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify=true")
	}
}

func TestConfig_Resolve_InsecureSkipVerify_False(t *testing.T) {
	c := NewHttpClient()
	transport, ok := c.client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("expected non-nil TLSClientConfig")
	}
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify=false by default")
	}
}
