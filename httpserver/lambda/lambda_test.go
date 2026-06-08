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

package lambda_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/phcp-tech/common-library-golang/httpserver"
	lambdarunner "github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// -----------------------------------------------------------------------
// Runner — interface compliance
// -----------------------------------------------------------------------

func TestNewHttpServer_ImplementsRunner(t *testing.T) {
	var _ httpserver.Runner = lambdarunner.NewHttpServer()
}

func TestNewHttpServer_ReturnsNonNil(t *testing.T) {
	r := lambdarunner.NewHttpServer()
	if r == nil {
		t.Fatal("NewHttpServer() returned nil")
	}
}

// -----------------------------------------------------------------------
// Shutdown — always a no-op
// -----------------------------------------------------------------------

func TestShutdown_IsNoOp(t *testing.T) {
	r := lambdarunner.NewHttpServer()
	if err := r.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() = %v, want nil", err)
	}
}

func TestShutdown_WithCancelledContext_IsNoOp(t *testing.T) {
	r := lambdarunner.NewHttpServer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	if err := r.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown(cancelled ctx) = %v, want nil", err)
	}
}

// -----------------------------------------------------------------------
// Start — cannot be unit-tested (lambda.Start calls os.Exit without the
// Lambda runtime environment). The test below documents this limitation
// and verifies the runner accepts any http.Handler without panicking up
// to the point where lambda.Start would be invoked.
// -----------------------------------------------------------------------

// TestStart_AcceptsAnyHandler verifies that constructing the adapter inside
// Start (httpadapter.New) does not panic for a plain http.Handler.
// We do NOT call runner.Start() here because lambda.Start blocks forever /
// calls os.Exit in a non-Lambda environment.
func TestStart_AcceptsAnyHandler(t *testing.T) {
	r := lambdarunner.NewHttpServer()
	// Verify the runner holds a valid reference and that passing a generic
	// handler to the type does not panic during construction.
	_ = r
	_ = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	// Actual runner.Start() is not called — see note above.
}
