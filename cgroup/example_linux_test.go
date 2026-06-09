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

//go:build linux
// +build linux

package cgroup_test

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/phcp-tech/common-library-golang/cgroup"
)

// ExampleCPULimitMilli shows how to read the CPU limit from cgroup v2
// and use it to cap GOMAXPROCS so the Go scheduler does not create more
// OS threads than the container is allowed to run.
func ExampleCPULimitMilli() {
	milli, err := cgroup.CPULimitMilli()
	if err != nil {
		slog.Error("read cpu limit", "error", err)
		return
	}
	if milli == 0 {
		fmt.Println("no CPU limit configured (unlimited)")
		return
	}
	// Convert millicores to whole CPUs, minimum 1.
	cpus := milli / 1000
	if cpus < 1 {
		cpus = 1
	}
	runtime.GOMAXPROCS(cpus)
	fmt.Printf("CPU limit: %d mCPU → GOMAXPROCS set to %d\n", milli, cpus)
}

// ExampleCPURequestMilli shows how to read the CPU request (guaranteed share)
// from cgroup v2. This is useful for logging resource allocation at startup.
func ExampleCPURequestMilli() {
	milli, err := cgroup.CPURequestMilli()
	if err != nil {
		slog.Error("read cpu request", "error", err)
		return
	}
	if milli == 0 {
		fmt.Println("no CPU request configured (BestEffort)")
		return
	}
	fmt.Printf("CPU request: %d mCPU\n", milli)
}

// ExampleMemoryLimitBytes shows how to read the memory limit from cgroup v2
// and use it to size an in-process cache as a fraction of the container limit.
func ExampleMemoryLimitBytes() {
	limitBytes, err := cgroup.MemoryLimitBytes()
	if err != nil {
		slog.Error("read memory limit", "error", err)
		return
	}
	if limitBytes == 0 {
		fmt.Println("no memory limit configured (unlimited)")
		return
	}
	// Use at most 25 % of the container memory limit for an in-process cache.
	cacheBytes := limitBytes / 4
	fmt.Printf("memory limit: %d MiB → cache budget: %d MiB\n",
		limitBytes>>20, cacheBytes>>20)
}

// ExampleMemoryRequestBytes shows how to read the memory soft limit (request)
// from cgroup v2. A value of 0 means no soft limit is set.
func ExampleMemoryRequestBytes() {
	requestBytes, err := cgroup.MemoryRequestBytes()
	if err != nil {
		slog.Error("read memory request", "error", err)
		return
	}
	if requestBytes == 0 {
		fmt.Println("no memory request configured")
		return
	}
	fmt.Printf("memory request (soft limit): %d MiB\n", requestBytes>>20)
}
