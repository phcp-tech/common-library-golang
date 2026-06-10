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

//go:build windows
// +build windows

package metrics

import (
	"testing"
)

// cgroup is unavailable on Windows; all four cgroup metrics must be "0"

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
