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
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// cgroup_linux.go reads the fixed paths /sys/fs/cgroup, /proc/self/cgroup and /proc/self/mountinfo, so these tests
// run against the cgroup setup of the machine they run on (cgroup v1, v2, or none in a minimal container).
// They check what must hold in every setup: the parsing helpers, and the invariants of the returned values.
// An error from a reader is accepted, because the files it needs may not exist on the host; for example the root
// cgroup of a cgroup v2 host has no cpu.max.

// ---------------------------------------------------------------------------
// parseInt
// ---------------------------------------------------------------------------

func TestParseInt(t *testing.T) {
	Convey("Test parseInt", t, func() {
		Convey("Test valid numbers, with surrounding white space", func() {
			for in, want := range map[string]int{
				"0":         0,
				"42":        42,
				"-1":        -1,
				" 100000\n": 100000,
				"\t7 ":      7,
			} {
				got, err := parseInt(in)
				So(err, ShouldBeNil)
				So(got, ShouldEqual, want)
			}
		})

		Convey("Test invalid numbers", func() {
			for _, in := range []string{"", "  ", "abc", "1.5", "max", "12abc"} {
				_, err := parseInt(in)
				So(err, ShouldNotBeNil)
			}
		})
	})
}

// ---------------------------------------------------------------------------
// detectCgroupV2
// ---------------------------------------------------------------------------

// hasV2ProcEntry reports whether /proc/self/cgroup has the cgroup v2 line "0::<path>", and returns the path.
func hasV2ProcEntry(t *testing.T) (string, bool) {
	t.Helper()
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), ":", 3)
		if len(parts) == 3 && parts[0] == "0" && parts[1] == "" {
			return parts[2], true
		}
	}
	return "", false
}

func TestDetectCgroupV2(t *testing.T) {
	Convey("Test detectCgroupV2", t, func() {
		root, path, ok := detectCgroupV2()

		_, statErr := os.Stat("/sys/fs/cgroup/cgroup.controllers")
		wantPath, hasEntry := hasV2ProcEntry(t)

		Convey("Test it is v2 only when cgroup.controllers exists and /proc/self/cgroup has a 0:: line", func() {
			So(ok, ShouldEqual, statErr == nil && hasEntry)
		})

		if ok {
			Convey("Test the v2 root and path", func() {
				So(root, ShouldEqual, "/sys/fs/cgroup")
				So(path, ShouldEqual, wantPath)
			})
		} else {
			Convey("Test the results are empty when it is not v2", func() {
				So(root, ShouldBeEmpty)
				So(path, ShouldBeEmpty)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// cgroup v1 helpers
// ---------------------------------------------------------------------------

func TestFindCgroupV1Mountpoint(t *testing.T) {
	Convey("Test findCgroupV1Mountpoint", t, func() {
		Convey("Test a controller that does not exist is an error", func() {
			mnt, err := findCgroupV1Mountpoint("no-such-controller")
			So(err, ShouldNotBeNil)
			So(mnt, ShouldBeEmpty)
		})

		Convey("Test a found mount point is an absolute path", func() {
			for _, controller := range []string{"cpu", "memory"} {
				mnt, err := findCgroupV1Mountpoint(controller)
				if err == nil {
					So(filepath.IsAbs(mnt), ShouldBeTrue)
				}
			}
		})
	})
}

func TestCgroupV1Dirs(t *testing.T) {
	Convey("Test cgroupV1CpuDir and cgroupV1MemoryDir", t, func() {
		Convey("Test cgroupV1CpuDir returns an absolute path or an error", func() {
			dir, err := cgroupV1CpuDir()
			if err == nil {
				So(filepath.IsAbs(dir), ShouldBeTrue)
			} else {
				So(dir, ShouldBeEmpty)
			}
		})

		Convey("Test cgroupV1MemoryDir returns an absolute path or an error", func() {
			dir, err := cgroupV1MemoryDir()
			if err == nil {
				So(filepath.IsAbs(dir), ShouldBeTrue)
			} else {
				So(dir, ShouldBeEmpty)
			}
		})
	})
}

// ---------------------------------------------------------------------------
// CPULimitMilli / CPURequestMilli
// ---------------------------------------------------------------------------

func TestCPULimitMilli(t *testing.T) {
	Convey("Test CPULimitMilli", t, func() {
		v, err := CPULimitMilli()
		if err != nil {
			// Neither a readable v2 cpu.max nor a readable v1 cpu controller on this host.
			So(v, ShouldEqual, 0)
		} else {
			// 0 means unlimited.
			So(v, ShouldBeGreaterThanOrEqualTo, 0)
		}
	})
}

func TestCPURequestMilli(t *testing.T) {
	Convey("Test CPURequestMilli", t, func() {
		v, err := CPURequestMilli()
		if err != nil {
			So(v, ShouldEqual, 0)
		} else {
			// 0 means no request (BestEffort).
			So(v, ShouldBeGreaterThanOrEqualTo, 0)
		}
	})
}

// ---------------------------------------------------------------------------
// MemoryLimitBytes / MemoryRequestBytes
// ---------------------------------------------------------------------------

func TestMemoryLimitBytes(t *testing.T) {
	Convey("Test MemoryLimitBytes", t, func() {
		v, err := MemoryLimitBytes()
		if err != nil {
			So(v, ShouldEqual, 0)
		} else {
			// 0 means unlimited. A v1 "no limit" value (a huge number) is mapped to 0.
			So(v, ShouldBeGreaterThanOrEqualTo, 0)
			So(v, ShouldBeLessThanOrEqualTo, int64(1)<<62)
		}
	})
}

func TestMemoryRequestBytes(t *testing.T) {
	Convey("Test MemoryRequestBytes", t, func() {
		v, err := MemoryRequestBytes()
		if err != nil {
			So(v, ShouldEqual, 0)
		} else {
			// 0 means "no soft limit" or "can not be read".
			So(v, ShouldBeGreaterThanOrEqualTo, 0)
			So(v, ShouldBeLessThanOrEqualTo, int64(1)<<62)
		}
	})
}

// ---------------------------------------------------------------------------
// CPUThrottleStat
// ---------------------------------------------------------------------------

func TestCPUThrottleStat(t *testing.T) {
	Convey("Test CPUThrottleStat", t, func() {
		first, err := CPUThrottleStat()
		if err != nil {
			So(first, ShouldResemble, CPUThrottle{})
			return
		}

		Convey("Test a period is throttled at most once", func() {
			So(first.Throttled, ShouldBeLessThanOrEqualTo, first.Periods)
		})

		Convey("Test the counters never decrease between two reads", func() {
			second, err := CPUThrottleStat()
			So(err, ShouldBeNil)
			So(second.Periods, ShouldBeGreaterThanOrEqualTo, first.Periods)
			So(second.Throttled, ShouldBeGreaterThanOrEqualTo, first.Throttled)
			So(second.ThrottledNs, ShouldBeGreaterThanOrEqualTo, first.ThrottledNs)
		})
	})
}
