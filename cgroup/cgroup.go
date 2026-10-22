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
	"strconv"
	"strings"
)

// CPUThrottle is the cumulative CPU throttle counters of a cgroup (from cpu.stat).
// The counters only increase, so the caller must take the difference of two reads.
// All fields are 0 when the cgroup has no CPU limit, or the counters can not be read.
type CPUThrottle struct {
	// Periods is nr_periods: the number of CFS periods that have elapsed.
	Periods uint64
	// Throttled is nr_throttled: the number of periods in which the cgroup used all its quota and was throttled.
	Throttled uint64
	// ThrottledNs is the total time the cgroup was throttled, in nanoseconds.
	ThrottledNs uint64
}

// parseCPUStat parses the content of cpu.stat of cgroup v1 and v2. Lines have the format "key value".
//
//	v2: nr_periods, nr_throttled, throttled_usec (microseconds)
//	v1: nr_periods, nr_throttled, throttled_time (nanoseconds)
//
// Unknown keys and malformed lines are ignored.
func parseCPUStat(data string) CPUThrottle {
	var t CPUThrottle
	for _, line := range strings.Split(data, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		v, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "nr_periods":
			t.Periods = v
		case "nr_throttled":
			t.Throttled = v
		case "throttled_usec":
			t.ThrottledNs = v * 1000
		case "throttled_time":
			t.ThrottledNs = v
		}
	}
	return t
}

// cpuWeightToMilli converts the cgroup v2 cpu.weight of a container to its CPU request (mCPU).
//
// Kubernetes sets cpu.shares = request(mCPU) * 1024 / 1000, and the container runtime converts shares to the
// cgroup v2 cpu.weight. Newer runtimes use a formula that maps 1024 shares to weight 100 (the kernel default):
//
//	weight = ceil(10 ^ ((log2(shares)^2 + 125*log2(shares)) / 612 - 7/34))
//
// The old linear formula (weight = 1 + (shares-2)*9999/262142) maps 1024 shares to weight 39. The old code of
// CPURequestMilli assumed the linear formula: a 500m request has weight 59 in the new formula, and was shown as
// 1511m (59 * 262144 / 10000 = 1546 shares). This function inverts the new formula.
//
// Limits of the result:
//   - ceil() loses information: several close requests have the same weight (for example 500m to 509m have
//     weight 59). The shares of the weight are in the range (inv(weight-1), inv(weight)], so the middle of the range
//     is returned. The error is up to about 2% for requests of 250m and more (500m is shown as about 505m), and up to
//     about 4% for 100m.
//   - On a runtime that still uses the old linear formula the result is wrong (a 500m request is shown as about
//     190m). There is no way to know the formula from the weight alone.
//
// It returns 0 for weight <= 1: shares = 2 is BestEffort (no request).
func cpuWeightToMilli(weight int64) int {
	if weight <= 1 {
		return 0
	}
	lo := cpuWeightToShares(float64(weight - 1)) // weight-1 >= 1, and weight 1 gives 2 shares
	hi := cpuWeightToShares(float64(weight))
	shares := (lo + hi) / 2
	return int(math.Round(shares * 1000 / 1024))
}

// cpuWeightToShares is the inverse of the new shares-to-weight formula, without the ceil().
// Solve l^2 + 125*l - 612*(log10(weight) + 7/34) = 0 for l = log2(shares).
func cpuWeightToShares(weight float64) float64 {
	e := math.Log10(weight) + 7.0/34.0
	l := (-125 + math.Sqrt(125*125+4*612*e)) / 2
	return math.Pow(2, l)
}
