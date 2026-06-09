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

package cgroup

import (
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------
// test helpers
// -----------------------------------------------------------------------

// setupFakeV2 builds a minimal cgroup v2 directory tree under t.TempDir()
// and overrides the package-level cgroupRoot / procSelfCgroup vars so all
// cgroup functions operate on the fake tree. The returned restore function
// must be deferred by the caller.
//
// files maps filenames (relative to cgroupRoot) to their contents.
// The v2 sentinel file "cgroup.controllers" is created automatically.
// /proc/self/cgroup is faked with a single "0::/" entry (process at root).
func setupFakeV2(t *testing.T, files map[string]string) (restore func()) {
	t.Helper()
	root := t.TempDir()

	// cgroup v2 sentinel
	writeTestFile(t, filepath.Join(root, "cgroup.controllers"), "cpu memory")

	for name, content := range files {
		writeTestFile(t, filepath.Join(root, name), content)
	}

	// fake /proc/self/cgroup: v2 entry "0::/" → process lives at cgroupRoot itself
	procFile := filepath.Join(t.TempDir(), "cgroup")
	writeTestFile(t, procFile, "0::/\n")

	origRoot, origProc := cgroupRoot, procSelfCgroup
	cgroupRoot = root
	procSelfCgroup = procFile
	return func() {
		cgroupRoot = origRoot
		procSelfCgroup = origProc
	}
}

// setupFakeV2Nested is like setupFakeV2 but the process cgroup entry points
// to a subdirectory, as in a real Kubernetes pod environment.
func setupFakeV2Nested(t *testing.T, cgroupPath string, files map[string]string) (restore func()) {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "cgroup.controllers"), "cpu memory")

	for name, content := range files {
		writeTestFile(t, filepath.Join(root, cgroupPath, name), content)
	}

	procFile := filepath.Join(t.TempDir(), "cgroup")
	writeTestFile(t, procFile, "0::"+cgroupPath+"\n")

	origRoot, origProc := cgroupRoot, procSelfCgroup
	cgroupRoot = root
	procSelfCgroup = procFile
	return func() {
		cgroupRoot = origRoot
		procSelfCgroup = origProc
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// -----------------------------------------------------------------------
// cgroupV2Dir
// -----------------------------------------------------------------------

func TestCgroupV2Dir_NoCgroupControllers(t *testing.T) {
	origRoot := cgroupRoot
	cgroupRoot = t.TempDir() // directory exists but no cgroup.controllers inside
	defer func() { cgroupRoot = origRoot }()

	_, err := cgroupV2Dir()
	if err == nil {
		t.Error("want error when cgroup.controllers is absent, got nil")
	}
}

func TestCgroupV2Dir_NoV2EntryInProcFile(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "cgroup.controllers"), "cpu memory")

	// Only v1 entries, no "0::" line
	procFile := filepath.Join(t.TempDir(), "cgroup")
	writeTestFile(t, procFile, "1:cpu,cpuacct:/kubepods\n2:memory:/kubepods\n")

	origRoot, origProc := cgroupRoot, procSelfCgroup
	cgroupRoot = root
	procSelfCgroup = procFile
	defer func() { cgroupRoot = origRoot; procSelfCgroup = origProc }()

	_, err := cgroupV2Dir()
	if err == nil {
		t.Error("want error when no v2 entry in proc file, got nil")
	}
}

func TestCgroupV2Dir_NestedPath(t *testing.T) {
	const cgPath = "/kubepods/burstable/pod123/ctr456"
	restore := setupFakeV2Nested(t, cgPath, map[string]string{
		"cpu.max": "max 100000\n",
	})
	defer restore()

	dir, err := cgroupV2Dir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty dir")
	}
}

// -----------------------------------------------------------------------
// CPULimitMilli
// -----------------------------------------------------------------------

func TestCPULimitMilli_Unlimited(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"cpu.max": "max 100000\n"})
	defer restore()

	v, err := CPULimitMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("unlimited: want 0, got %d", v)
	}
}

func TestCPULimitMilli_OneCPU(t *testing.T) {
	// quota=100000, period=100000 → 1 core → 1000 mCPU
	restore := setupFakeV2(t, map[string]string{"cpu.max": "100000 100000\n"})
	defer restore()

	v, err := CPULimitMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1000 {
		t.Errorf("1 CPU: want 1000 mCPU, got %d", v)
	}
}

func TestCPULimitMilli_HalfCPU(t *testing.T) {
	// quota=50000, period=100000 → 500 mCPU
	restore := setupFakeV2(t, map[string]string{"cpu.max": "50000 100000\n"})
	defer restore()

	v, err := CPULimitMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 500 {
		t.Errorf("0.5 CPU: want 500 mCPU, got %d", v)
	}
}

func TestCPULimitMilli_FourCPU(t *testing.T) {
	// quota=400000, period=100000 → 4000 mCPU
	restore := setupFakeV2(t, map[string]string{"cpu.max": "400000 100000\n"})
	defer restore()

	v, err := CPULimitMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 4000 {
		t.Errorf("4 CPU: want 4000 mCPU, got %d", v)
	}
}

func TestCPULimitMilli_MissingFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{}) // no cpu.max
	defer restore()

	_, err := CPULimitMilli()
	if err == nil {
		t.Error("want error for missing cpu.max, got nil")
	}
}

func TestCPULimitMilli_MalformedFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"cpu.max": "notanumber 100000\n"})
	defer restore()

	_, err := CPULimitMilli()
	if err == nil {
		t.Error("want error for malformed cpu.max, got nil")
	}
}

func TestCPULimitMilli_MissingCgroupV2(t *testing.T) {
	origRoot := cgroupRoot
	cgroupRoot = t.TempDir() // no cgroup.controllers
	defer func() { cgroupRoot = origRoot }()

	_, err := CPULimitMilli()
	if err == nil {
		t.Error("want error when cgroup v2 unavailable, got nil")
	}
}

// -----------------------------------------------------------------------
// CPURequestMilli
// -----------------------------------------------------------------------

// Expected values computed from the formula:
//   shares = (weight*262144 + 5000) / 10000   [integer division]
//   mCPU   = (shares*1000 + 512) / 1024       [integer division]
//
//   weight=1    → shares=26   → mCPU=25
//   weight=100  → shares=2621 → mCPU=2560
//   weight=10000 → shares=262144 → mCPU=256000

func TestCPURequestMilli_Weight1(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"cpu.weight": "1\n"})
	defer restore()

	v, err := CPURequestMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 25 {
		t.Errorf("weight=1: want 25 mCPU, got %d", v)
	}
}

func TestCPURequestMilli_Weight100(t *testing.T) {
	// weight=100 is the default (no explicit request) in Kubernetes
	restore := setupFakeV2(t, map[string]string{"cpu.weight": "100\n"})
	defer restore()

	v, err := CPURequestMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 2560 {
		t.Errorf("weight=100: want 2560 mCPU, got %d", v)
	}
}

func TestCPURequestMilli_Weight10000(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"cpu.weight": "10000\n"})
	defer restore()

	v, err := CPURequestMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 256000 {
		t.Errorf("weight=10000: want 256000 mCPU, got %d", v)
	}
}

func TestCPURequestMilli_WeightClamped(t *testing.T) {
	// weight > 10000 is clamped to 10000
	restore := setupFakeV2(t, map[string]string{"cpu.weight": "99999\n"})
	defer restore()

	v, err := CPURequestMilli()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 256000 {
		t.Errorf("weight=99999 (clamped to 10000): want 256000 mCPU, got %d", v)
	}
}

func TestCPURequestMilli_MissingFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{})
	defer restore()

	_, err := CPURequestMilli()
	if err == nil {
		t.Error("want error for missing cpu.weight, got nil")
	}
}

func TestCPURequestMilli_MalformedFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"cpu.weight": "abc\n"})
	defer restore()

	_, err := CPURequestMilli()
	if err == nil {
		t.Error("want error for malformed cpu.weight, got nil")
	}
}

// -----------------------------------------------------------------------
// MemoryLimitBytes
// -----------------------------------------------------------------------

func TestMemoryLimitBytes_Unlimited(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"memory.max": "max\n"})
	defer restore()

	v, err := MemoryLimitBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("unlimited: want 0, got %d", v)
	}
}

func TestMemoryLimitBytes_1GiB(t *testing.T) {
	const oneGiB = int64(1 << 30)
	restore := setupFakeV2(t, map[string]string{"memory.max": "1073741824\n"})
	defer restore()

	v, err := MemoryLimitBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != oneGiB {
		t.Errorf("1 GiB: want %d, got %d", oneGiB, v)
	}
}

func TestMemoryLimitBytes_512MiB(t *testing.T) {
	const fiveTwelveMiB = int64(512 << 20)
	restore := setupFakeV2(t, map[string]string{"memory.max": "536870912\n"})
	defer restore()

	v, err := MemoryLimitBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != fiveTwelveMiB {
		t.Errorf("512 MiB: want %d, got %d", fiveTwelveMiB, v)
	}
}

func TestMemoryLimitBytes_MissingFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{})
	defer restore()

	_, err := MemoryLimitBytes()
	if err == nil {
		t.Error("want error for missing memory.max, got nil")
	}
}

func TestMemoryLimitBytes_MalformedFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"memory.max": "notanumber\n"})
	defer restore()

	_, err := MemoryLimitBytes()
	if err == nil {
		t.Error("want error for malformed memory.max, got nil")
	}
}

// -----------------------------------------------------------------------
// MemoryRequestBytes
// -----------------------------------------------------------------------

func TestMemoryRequestBytes_Zero(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"memory.low": "0\n"})
	defer restore()

	v, err := MemoryRequestBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("no soft limit (0): want 0, got %d", v)
	}
}

func TestMemoryRequestBytes_Max(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"memory.low": "max\n"})
	defer restore()

	v, err := MemoryRequestBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("max: want 0, got %d", v)
	}
}

func TestMemoryRequestBytes_256MiB(t *testing.T) {
	const twoFiftySixMiB = int64(256 << 20)
	restore := setupFakeV2(t, map[string]string{"memory.low": "268435456\n"})
	defer restore()

	v, err := MemoryRequestBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != twoFiftySixMiB {
		t.Errorf("256 MiB: want %d, got %d", twoFiftySixMiB, v)
	}
}

func TestMemoryRequestBytes_MissingFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{})
	defer restore()

	_, err := MemoryRequestBytes()
	if err == nil {
		t.Error("want error for missing memory.low, got nil")
	}
}

func TestMemoryRequestBytes_MalformedFile(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{"memory.low": "bad\n"})
	defer restore()

	_, err := MemoryRequestBytes()
	if err == nil {
		t.Error("want error for malformed memory.low, got nil")
	}
}

// -----------------------------------------------------------------------
// return type and non-negative sanity (mirrors Windows tests)
// -----------------------------------------------------------------------

func TestAllFunctions_ReturnTypesAndNonNegative(t *testing.T) {
	restore := setupFakeV2(t, map[string]string{
		"cpu.max":    "max 100000\n",
		"cpu.weight": "100\n",
		"memory.max": "max\n",
		"memory.low": "0\n",
	})
	defer restore()

	cpuLimit, err := CPULimitMilli()
	if err != nil {
		t.Errorf("CPULimitMilli: %v", err)
	}
	var _ int = cpuLimit
	if cpuLimit < 0 {
		t.Errorf("CPULimitMilli must be >= 0, got %d", cpuLimit)
	}

	cpuReq, err := CPURequestMilli()
	if err != nil {
		t.Errorf("CPURequestMilli: %v", err)
	}
	var _ int = cpuReq
	if cpuReq < 0 {
		t.Errorf("CPURequestMilli must be >= 0, got %d", cpuReq)
	}

	memLimit, err := MemoryLimitBytes()
	if err != nil {
		t.Errorf("MemoryLimitBytes: %v", err)
	}
	var _ int64 = memLimit
	if memLimit < 0 {
		t.Errorf("MemoryLimitBytes must be >= 0, got %d", memLimit)
	}

	memReq, err := MemoryRequestBytes()
	if err != nil {
		t.Errorf("MemoryRequestBytes: %v", err)
	}
	var _ int64 = memReq
	if memReq < 0 {
		t.Errorf("MemoryRequestBytes must be >= 0, got %d", memReq)
	}
}
