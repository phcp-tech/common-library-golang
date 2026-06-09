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

// CPURequestMilli returns the cgroup CPU request in millicores.
// On macOS cgroup information is not available; it always returns 0, nil.
func CPURequestMilli() (int, error) {
	// On macOS, return 0
	return 0, nil
}

// CPULimitMilli returns the cgroup CPU limit in millicores.
// On macOS cgroup information is not available; it always returns 0, nil.
func CPULimitMilli() (int, error) {
	// On macOS, return 0
	return 0, nil
}

// MemoryLimitBytes returns the cgroup memory limit in bytes.
// On macOS cgroup information is not available; it always returns 0, nil.
func MemoryLimitBytes() (int64, error) {
	return 0, nil
}

// MemoryRequestBytes returns the cgroup memory request (soft limit) in bytes.
// On macOS cgroup information is not available; it always returns 0, nil.
func MemoryRequestBytes() (int64, error) {
	return 0, nil
}
