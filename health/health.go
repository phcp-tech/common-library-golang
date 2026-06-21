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

// Package health provides a composable health-check mechanism.
//
// Each infrastructure component (postgres, redis, …) supplies a [Checker] that
// reports its own name and reachability. [Check] runs all checkers and returns
// their combined results, suitable for a /health HTTP endpoint.
//
// Basic usage:
//
//	router.GET("/health", func(c *gin.Context) {
//	    c.JSON(http.StatusOK, health.Check(
//	        c.Request.Context(),
//	        postgres.HealthChecker(),
//	    ))
//	})
package health

import "context"

// StatusHealthy is the status value returned when a component is reachable.
const StatusHealthy = 1

// StatusUnhealthy is the status value returned when a component is unreachable.
const StatusUnhealthy = 0

// Result holds the health status of a single named component.
type Result struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// Checker is a function that returns the health status of one component.
// The Checker owns both its name and status; callers compose multiple checkers
// via [Check] without knowledge of individual component names.
type Checker func(ctx context.Context) Result

// Check runs each checker in registration order and returns their combined results.
// An empty checkers list returns an empty (non-nil) slice.
func Check(ctx context.Context, checkers ...Checker) []Result {
	results := make([]Result, 0, len(checkers))
	for _, c := range checkers {
		results = append(results, c(ctx))
	}
	return results
}
