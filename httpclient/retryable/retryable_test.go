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

package retryable_test

import (
	"testing"
	"time"

	"github.com/phcp-tech/common-library-golang/httpclient/retryable"
)

// -----------------------------------------------------------------------
// NewHttpClient — construction
// -----------------------------------------------------------------------

func TestNewHttpClient_DefaultConfig_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient()
	if c == nil {
		t.Fatal("NewHttpClient() returned nil")
	}
}

func TestNewHttpClient_ExplicitRetry_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient(retryable.Config{RetryMax: 2})
	if c == nil {
		t.Fatal("NewHttpClient(RetryMax=2) returned nil")
	}
}

func TestNewHttpClient_CustomTimeout_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient(retryable.Config{
		Timeout:  5 * time.Second,
		RetryMax: 1,
	})
	if c == nil {
		t.Fatal("NewHttpClient with custom timeout returned nil")
	}
}

func TestNewHttpClient_InsecureSkipVerify_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient(retryable.Config{InsecureSkipVerify: true})
	if c == nil {
		t.Fatal("NewHttpClient with InsecureSkipVerify returned nil")
	}
}

// -----------------------------------------------------------------------
// Client — underlying client
// -----------------------------------------------------------------------

func TestClient_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient()
	if c.Client() == nil {
		t.Error("Client() returned nil underlying retryablehttp.Client")
	}
}

func TestClient_StandardClient_ReturnsNonNil(t *testing.T) {
	c := retryable.NewHttpClient()
	if c.Client().StandardClient() == nil {
		t.Error("StandardClient() returned nil *http.Client")
	}
}
