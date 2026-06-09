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

// Package cgroup reads CPU and memory resource limits from the Linux cgroup v2
// unified hierarchy (/sys/fs/cgroup). cgroup v1 is not supported.
// On non-Linux platforms all functions return (0, nil).
//
// All public functions return 0 when no limit is configured (unlimited).
package cgroup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// cgroupRoot is the cgroup v2 unified hierarchy mount point.
// Overridden in tests to point to a temporary directory.
var cgroupRoot = "/sys/fs/cgroup"

// procSelfCgroup is the path to the current process's cgroup membership file.
// Overridden in tests to point to a temporary file.
var procSelfCgroup = "/proc/self/cgroup"

// cgroupV2Dir returns the cgroup v2 directory for the current process.
// It checks that cgroupRoot is a v2 hierarchy (presence of cgroup.controllers),
// then reads procSelfCgroup to locate the "0::<path>" entry.
func cgroupV2Dir() (string, error) {
	if _, err := os.Stat(filepath.Join(cgroupRoot, "cgroup.controllers")); err != nil {
		return "", fmt.Errorf("cgroup v2 not available at %s: %w", cgroupRoot, err)
	}
	data, err := os.ReadFile(procSelfCgroup)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", procSelfCgroup, err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, ":", 3)
		// v2 entry is always "0::<path>"
		if len(parts) == 3 && parts[0] == "0" && parts[1] == "" {
			return filepath.Join(cgroupRoot, parts[2]), nil
		}
	}
	return "", fmt.Errorf("no cgroup v2 entry (0::<path>) in %s", procSelfCgroup)
}

// CPULimitMilli returns the CPU limit of the current process's cgroup in
// millicores (mCPU). Reads cpu.max; "max <period>" means unlimited → returns 0.
func CPULimitMilli() (int, error) {
	dir, err := cgroupV2Dir()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "cpu.max"))
	if err != nil {
		return 0, fmt.Errorf("read cpu.max: %w", err)
	}
	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected cpu.max format: %q", string(data))
	}
	if fields[0] == "max" {
		slog.Debug("cgroup v2 cpu limit: unlimited")
		return 0, nil
	}
	quota, err1 := strconv.Atoi(fields[0])
	period, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("parse cpu.max quota=%q period=%q: %v / %v", fields[0], fields[1], err1, err2)
	}
	if period <= 0 {
		return 0, fmt.Errorf("cpu.max period is zero or negative")
	}
	mcpu := int(float64(quota) / float64(period) * 1000.0)
	slog.Debug("cgroup v2 cpu limit", "mCPU", mcpu)
	return mcpu, nil
}

// CPURequestMilli returns the CPU request of the current process's cgroup in
// millicores (mCPU). Reads cpu.weight and converts to millicores via the
// Kubernetes weight→shares→mCPU formula:
//
//	shares = (weight × 262144 + 5000) / 10000   [clamped to 2–262144]
//	mCPU   = (shares × 1000 + 512) / 1024
func CPURequestMilli() (int, error) {
	dir, err := cgroupV2Dir()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "cpu.weight"))
	if err != nil {
		return 0, fmt.Errorf("read cpu.weight: %w", err)
	}
	weight, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse cpu.weight: %w", err)
	}
	if weight < 1 {
		weight = 1
	}
	if weight > 10000 {
		weight = 10000
	}
	shares := (weight*262144 + 5000) / 10000
	if shares < 2 {
		shares = 2
	}
	if shares > 262144 {
		shares = 262144
	}
	mcpu := int((shares*1000 + 512) / 1024)
	slog.Debug("cgroup v2 cpu request", "weight", weight, "mCPU", mcpu)
	return mcpu, nil
}

// MemoryLimitBytes returns the memory limit of the current process's cgroup in
// bytes. Reads memory.max; "max" means unlimited → returns 0.
func MemoryLimitBytes() (int64, error) {
	dir, err := cgroupV2Dir()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "memory.max"))
	if err != nil {
		return 0, fmt.Errorf("read memory.max: %w", err)
	}
	s := strings.TrimSpace(string(data))
	if s == "max" {
		slog.Debug("cgroup v2 memory limit: unlimited")
		return 0, nil
	}
	limit, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.max: %w", err)
	}
	slog.Debug("cgroup v2 memory limit", "bytes", limit)
	return limit, nil
}

// MemoryRequestBytes returns the memory soft limit (request) of the current
// process's cgroup in bytes. Reads memory.low; "0" or "max" → returns 0.
func MemoryRequestBytes() (int64, error) {
	dir, err := cgroupV2Dir()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "memory.low"))
	if err != nil {
		return 0, fmt.Errorf("read memory.low: %w", err)
	}
	s := strings.TrimSpace(string(data))
	if s == "0" || s == "max" {
		slog.Debug("cgroup v2 memory request: no soft limit")
		return 0, nil
	}
	low, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.low: %w", err)
	}
	slog.Debug("cgroup v2 memory request", "bytes", low)
	return low, nil
}
