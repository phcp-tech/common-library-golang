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

//go:build darwin
// +build darwin

package cgroup

// CPURequestMilli gets CPU request - macOS fallback implementation
func CPURequestMilli() (int, error) {
	// On macOS, return 0
	return 0, nil
}

// CPULimitMilli gets CPU limits - macOS fallback implementation
func CPULimitMilli() (int, error) {
	// On macOS, return 0
	return 0, nil
}

// MemoryLimitBytes gets memory limit - macOS fallback implementation
func MemoryLimitBytes() (int64, error) {
	return 0, nil
}

// MemoryRequestBytes gets memory request - macOS fallback implementation
func MemoryRequestBytes() (int64, error) {
	return 0, nil
}

// CPUThrottleStat gets CPU throttle counters - macOS fallback implementation
func CPUThrottleStat() (CPUThrottle, error) {
	return CPUThrottle{}, nil
}
