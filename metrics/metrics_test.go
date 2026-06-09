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

package metrics

import (
	"runtime"
	"strconv"
	"testing"
)

// findMetric is a test helper that locates a NameValue entry by name inside
// the slice returned by GetMetrics and returns (value, true) when found.
func findMetric(metrics []NameValue, name string) (string, bool) {
	for _, m := range metrics {
		if m.Name == name {
			return m.Value, true
		}
	}
	return "", false
}

// TestGetMetricsReturnsNonNil verifies that GetMetrics() never returns nil.
func TestGetMetricsReturnsNonNil(t *testing.T) {
	result := GetMetrics()
	if result == nil {
		t.Fatal("GetMetrics() returned nil")
	}
}

// TestGetMetricsIsNotEmpty verifies that the returned slice contains at least
// one element.
func TestGetMetricsIsNotEmpty(t *testing.T) {
	result := GetMetrics()
	if len(result) == 0 {
		t.Fatal("GetMetrics() returned an empty slice")
	}
}

// TestGetMetricsContainsExpectedKeys verifies that all documented metric names
// are present in the returned slice.
func TestGetMetricsContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{
		"cpuPercent",
		"memorySize",
		"threads",
		"goroutines",
		"gomaxprocs",
		"numCPU",
		"cpuRequest",
		"cpuLimit",
		"memoryRequest",
		"memoryLimit",
		"age",
	}

	result := GetMetrics()
	for _, key := range expectedKeys {
		if _, found := findMetric(result, key); !found {
			t.Errorf("GetMetrics() missing expected key %q", key)
		}
	}
}

// TestGetMetricsCpuPercentIsNonNegative verifies that the cpuPercent value
// parses as a non-negative integer.
func TestGetMetricsCpuPercentIsNonNegative(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "cpuPercent")
	if !found {
		t.Fatal("GetMetrics() missing cpuPercent")
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("cpuPercent value %q is not a valid integer: %v", val, err)
	}
	if n < 0 {
		t.Errorf("cpuPercent must be >= 0, got %d", n)
	}
}

// TestGetMetricsGoroutinesIsPositive verifies that the goroutines metric
// reports a value greater than zero (at least the test goroutine is running).
func TestGetMetricsGoroutinesIsPositive(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "goroutines")
	if !found {
		t.Fatal("GetMetrics() missing goroutines")
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("goroutines value %q is not a valid integer: %v", val, err)
	}
	if n <= 0 {
		t.Errorf("goroutines must be > 0, got %d", n)
	}
}

// TestGetMetricsGomaxprocsMatchesRuntime verifies that the gomaxprocs metric
// matches the value returned by runtime.GOMAXPROCS(0).
func TestGetMetricsGomaxprocsMatchesRuntime(t *testing.T) {
	expected := runtime.GOMAXPROCS(0)
	result := GetMetrics()
	val, found := findMetric(result, "gomaxprocs")
	if !found {
		t.Fatal("GetMetrics() missing gomaxprocs")
	}
	got, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("gomaxprocs value %q is not a valid integer: %v", val, err)
	}
	if got != expected {
		t.Errorf("gomaxprocs: expected %d (runtime.GOMAXPROCS(0)), got %d", expected, got)
	}
}

// TestGetMetricsNumCPUMatchesRuntime verifies that the numCPU metric matches
// the value returned by runtime.NumCPU().
func TestGetMetricsNumCPUMatchesRuntime(t *testing.T) {
	expected := runtime.NumCPU()
	result := GetMetrics()
	val, found := findMetric(result, "numCPU")
	if !found {
		t.Fatal("GetMetrics() missing numCPU")
	}
	got, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("numCPU value %q is not a valid integer: %v", val, err)
	}
	if got != expected {
		t.Errorf("numCPU: expected %d (runtime.NumCPU()), got %d", expected, got)
	}
}

// TestGetMetricsCpuRequestIsZeroOnWindows verifies that on Windows the cpuRequest
// cgroup metric is always "0" (cgroup not available).
func TestGetMetricsCpuRequestIsZeroOnWindows(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "cpuRequest")
	if !found {
		t.Fatal("GetMetrics() missing cpuRequest")
	}
	if val != "0" {
		t.Errorf("cpuRequest on Windows: expected \"0\", got %q", val)
	}
}

// TestGetMetricsCpuLimitIsZeroOnWindows verifies that on Windows the cpuLimit
// cgroup metric is always "0".
func TestGetMetricsCpuLimitIsZeroOnWindows(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "cpuLimit")
	if !found {
		t.Fatal("GetMetrics() missing cpuLimit")
	}
	if val != "0" {
		t.Errorf("cpuLimit on Windows: expected \"0\", got %q", val)
	}
}

// TestGetMetricsMemoryRequestIsZeroOnWindows verifies that on Windows the
// memoryRequest cgroup metric is always "0".
func TestGetMetricsMemoryRequestIsZeroOnWindows(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "memoryRequest")
	if !found {
		t.Fatal("GetMetrics() missing memoryRequest")
	}
	if val != "0" {
		t.Errorf("memoryRequest on Windows: expected \"0\", got %q", val)
	}
}

// TestGetMetricsMemoryLimitIsZeroOnWindows verifies that on Windows the
// memoryLimit cgroup metric is always "0".
func TestGetMetricsMemoryLimitIsZeroOnWindows(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "memoryLimit")
	if !found {
		t.Fatal("GetMetrics() missing memoryLimit")
	}
	if val != "0" {
		t.Errorf("memoryLimit on Windows: expected \"0\", got %q", val)
	}
}

// TestGetMetricsAgeIsNonEmpty verifies that the age metric is a non-empty string
// with the expected "Xd Xh Xm Xs" format.
func TestGetMetricsAgeIsNonEmpty(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "age")
	if !found {
		t.Fatal("GetMetrics() missing age")
	}
	if val == "" {
		t.Error("age metric must not be empty")
	}
	// The age format produced by libTime.GetAge is "Xd Xh Xm Xs"
	// Verify it contains 'd', 'h', 'm', 's' as format markers.
	for _, marker := range []byte{'d', 'h', 'm', 's'} {
		found := false
		for _, ch := range []byte(val) {
			if ch == marker {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("age value %q missing expected marker '%c'", val, marker)
		}
	}
}

// TestGetMetricsMemorySizeIsNonNegative verifies that the memorySize metric
// parses as a non-negative integer.
func TestGetMetricsMemorySizeIsNonNegative(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "memorySize")
	if !found {
		t.Fatal("GetMetrics() missing memorySize")
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("memorySize value %q is not a valid integer: %v", val, err)
	}
	if n < 0 {
		t.Errorf("memorySize must be >= 0, got %d", n)
	}
}

// TestGetMetricsThreadsIsNonNegative verifies that the threads metric
// parses as a non-negative integer.
func TestGetMetricsThreadsIsNonNegative(t *testing.T) {
	result := GetMetrics()
	val, found := findMetric(result, "threads")
	if !found {
		t.Fatal("GetMetrics() missing threads")
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("threads value %q is not a valid integer: %v", val, err)
	}
	if n < 0 {
		t.Errorf("threads must be >= 0, got %d", n)
	}
}

// TestGetMetricsAllValuesAreNonEmpty verifies that every NameValue entry
// returned by GetMetrics has a non-empty Name and a parseable Value.
func TestGetMetricsAllValuesAreNonEmpty(t *testing.T) {
	result := GetMetrics()
	for _, m := range result {
		if m.Name == "" {
			t.Error("GetMetrics() returned an entry with an empty Name")
		}
		// Value must be present (may be "0" but never completely empty for numeric fields)
		// The "age" field is the only string value; others are all numeric.
		if m.Name != "age" && m.Value == "" {
			t.Errorf("metric %q has an empty Value", m.Name)
		}
	}
}

// TestGetMetricsReturnedSliceHasExactLength verifies that the slice contains
// exactly the 11 documented metrics.
func TestGetMetricsReturnedSliceHasExactLength(t *testing.T) {
	const expectedLen = 11
	result := GetMetrics()
	if len(result) != expectedLen {
		t.Errorf("GetMetrics() slice length: expected %d, got %d", expectedLen, len(result))
	}
}
