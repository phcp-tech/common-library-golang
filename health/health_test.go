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

package health

import (
	"context"
	"testing"
)

func TestStatusConstants(t *testing.T) {
	if StatusHealthy == StatusUnhealthy {
		t.Error("StatusHealthy and StatusUnhealthy must be distinct values")
	}
	if StatusUnhealthy != 0 {
		t.Errorf("StatusUnhealthy = %d, want 0", StatusUnhealthy)
	}
	if StatusHealthy != 1 {
		t.Errorf("StatusHealthy = %d, want 1", StatusHealthy)
	}
}

func TestCheck_Empty(t *testing.T) {
	results := Check(context.Background())
	if results == nil {
		t.Error("Check() with no checkers must return non-nil slice")
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestCheck_SingleChecker(t *testing.T) {
	checker := func(ctx context.Context) Result {
		return Result{Name: "db", Status: StatusHealthy}
	}
	results := Check(context.Background(), checker)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Name != "db" {
		t.Errorf("results[0].Name = %q, want %q", results[0].Name, "db")
	}
	if results[0].Status != StatusHealthy {
		t.Errorf("results[0].Status = %d, want StatusHealthy (%d)", results[0].Status, StatusHealthy)
	}
}

func TestCheck_MultipleCheckers_OrderPreserved(t *testing.T) {
	names := []string{"postgres", "redis", "s3"}
	checkers := make([]Checker, len(names))
	for i, name := range names {
		n := name // capture
		checkers[i] = func(ctx context.Context) Result {
			return Result{Name: n, Status: StatusHealthy}
		}
	}

	results := Check(context.Background(), checkers...)
	if len(results) != len(names) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(names))
	}
	for i, want := range names {
		if results[i].Name != want {
			t.Errorf("results[%d].Name = %q, want %q", i, results[i].Name, want)
		}
	}
}

func TestCheck_UnhealthyChecker(t *testing.T) {
	checker := func(ctx context.Context) Result {
		return Result{Name: "db", Status: StatusUnhealthy}
	}
	results := Check(context.Background(), checker)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Status != StatusUnhealthy {
		t.Errorf("results[0].Status = %d, want StatusUnhealthy (%d)", results[0].Status, StatusUnhealthy)
	}
}

func TestCheck_ContextPropagated(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "sentinel")
	var receivedCtx context.Context
	checker := func(c context.Context) Result {
		receivedCtx = c
		return Result{Name: "x", Status: StatusHealthy}
	}
	Check(ctx, checker)
	if receivedCtx.Value(key{}) != "sentinel" {
		t.Error("context was not propagated to checker")
	}
}
