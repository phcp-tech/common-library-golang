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

package cgroup

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseCPUStat(t *testing.T) {
	Convey("Test parseCPUStat", t, func() {
		Convey("Test cgroup v2 cpu.stat", func() {
			data := "usage_usec 123456\nuser_usec 100000\nsystem_usec 23456\n" +
				"nr_periods 500\nnr_throttled 120\nthrottled_usec 3500000\nnr_bursts 0\nburst_usec 0\n"
			So(parseCPUStat(data), ShouldResemble, CPUThrottle{Periods: 500, Throttled: 120, ThrottledNs: 3500000000})
		})

		Convey("Test cgroup v1 cpu.stat", func() {
			data := "nr_periods 500\nnr_throttled 120\nthrottled_time 3500000000\n"
			So(parseCPUStat(data), ShouldResemble, CPUThrottle{Periods: 500, Throttled: 120, ThrottledNs: 3500000000})
		})

		Convey("Test CRLF and no trailing newline", func() {
			So(parseCPUStat("nr_periods 7\r\nnr_throttled 3"), ShouldResemble, CPUThrottle{Periods: 7, Throttled: 3})
		})

		Convey("Test no CPU limit: counters are 0", func() {
			So(parseCPUStat("nr_periods 0\nnr_throttled 0\nthrottled_usec 0\n"), ShouldResemble, CPUThrottle{})
		})

		Convey("Test empty, malformed lines and unknown keys are ignored", func() {
			So(parseCPUStat(""), ShouldResemble, CPUThrottle{})
			So(parseCPUStat("nr_periods\nnr_throttled abc\nfoo 1\n\n"), ShouldResemble, CPUThrottle{})
		})
	})
}

// weightFromMilli is the new shares-to-weight formula of the container runtime (the forward direction of
// cpuWeightToMilli). K8s: shares = request(mCPU) * 1024 / 1000.
func weightFromMilli(milli int64) int64 {
	shares := float64(milli * 1024 / 1000)
	l := math.Log2(shares)
	return int64(math.Ceil(math.Pow(10, (l*l+125*l)/612.0-7.0/34.0)))
}

func TestCPUWeightToMilli(t *testing.T) {
	Convey("Test cpuWeightToMilli", t, func() {
		Convey("Test the case seen on UAT: request 500m has weight 59, it was shown as 1511m", func() {
			So(weightFromMilli(500), ShouldEqual, 59)
			got := cpuWeightToMilli(59)
			So(got, ShouldBeBetweenOrEqual, 495, 515)
			So(got, ShouldNotEqual, 1511)
		})

		Convey("Test the kernel default: weight 100 is 1024 shares, about 1000m", func() {
			So(cpuWeightToMilli(100), ShouldBeBetweenOrEqual, 985, 1015)
		})

		Convey("Test round requests come back within 2% (5% below 250m: the weight has fewer steps there)", func() {
			for _, milli := range []int64{100, 250, 500, 750, 1000, 1500, 2000, 4000, 8000} {
				tolerance := 0.02
				if milli < 250 {
					tolerance = 0.05
				}
				got := cpuWeightToMilli(weightFromMilli(milli))
				So(math.Abs(float64(got)-float64(milli))/float64(milli), ShouldBeLessThan, tolerance)
			}
		})

		Convey("Test the result grows with the weight", func() {
			prev := 0
			for w := int64(2); w <= 10000; w++ {
				got := cpuWeightToMilli(w)
				So(got, ShouldBeGreaterThanOrEqualTo, prev)
				prev = got
			}
		})

		Convey("Test weight 1 (BestEffort) and below is no request", func() {
			So(cpuWeightToMilli(1), ShouldEqual, 0)
			So(cpuWeightToMilli(0), ShouldEqual, 0)
		})
	})
}
