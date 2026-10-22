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

package process

import (
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/phcp-tech/common-library-golang/cgroup"
	"github.com/shirou/gopsutil/v4/process"
)

// cpuSampleInterval is the length of one CPU sample window.
const cpuSampleInterval = 5 * time.Second

var (
	cpuSamplerOnce sync.Once
	// cpuPercentBits holds the float64 bits of the CPU percent of the latest sample window.
	cpuPercentBits atomic.Uint64
	// cpuThrottledPercentBits holds the float64 bits of the throttled period percent of the latest sample window.
	cpuThrottledPercentBits atomic.Uint64
	// cpuThrottledMillis holds the throttled time (milliseconds) of the latest sample window.
	cpuThrottledMillis atomic.Int64
)

type ProcessInfo struct {
	// CpuPercent is the average CPU percent of the latest 5 seconds (100% = 1 core). It is 0 in the first 5 seconds.
	CpuPercent float64
	// CpuThrottledPercent is the percent of CFS periods in the latest 5 seconds in which the cgroup was throttled
	// by its CPU limit. 0 = not throttled. It is always 0 without a CPU limit, or when cpu.stat can not be read.
	CpuThrottledPercent float64
	// CpuThrottledMillis is the total time (milliseconds) the cgroup was throttled in the latest 5 seconds.
	// The time of all threads is added, so it can be larger than 5000.
	CpuThrottledMillis int64
	MemorySize         uint64
	Threads            int
	Goroutines         int
}

// throttleDelta calculates the throttle values of one sample window from two cumulative readings.
// percent is the throttled periods as a percentage of all periods in the window. millis is the throttled time.
// It returns 0, 0 when no period elapsed, or when a counter went backward (the cgroup was recreated).
func throttleDelta(prev, cur cgroup.CPUThrottle) (percent float64, millis int64) {
	if cur.Periods <= prev.Periods || cur.Throttled < prev.Throttled || cur.ThrottledNs < prev.ThrottledNs {
		return 0, 0
	}
	percent = float64(cur.Throttled-prev.Throttled) / float64(cur.Periods-prev.Periods) * 100
	millis = int64((cur.ThrottledNs - prev.ThrottledNs) / uint64(time.Millisecond))
	return percent, millis
}

// startCpuSampler samples the CPU percent of this process in the background every cpuSampleInterval.
//
// It replaces p.Percent(time.Second) in GetProcessInfo. Percent(time.Second) sleeps 1 second in the caller, and
// the caller is the single Kafka consumer goroutine of TMT (ProcessQuotes calls metrics.GetMetrics every
// processedBase ticks), so the consumer stopped for 1 second at every call. Also, NewProcess was called every time,
// so there was no previous sample and the 1 second sleep was the only way to get a time window.
//
// Now the Process object is created once. Percent(0) returns the CPU percent since the previous call and does not
// sleep. The first Percent(0) call only sets the base point.
//
// The same sampler also reads the CFS throttle counters of the cgroup (cpu.stat), which are cumulative. The
// difference between two samples shows whether the CPU limit held the process back in the window. A cpuPercent
// that stops at the limit (for example 100% with a 1000m limit) can not tell "needs exactly 1 core" from "needs
// more but is throttled"; the throttled percent can.
func startCpuSampler() {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return
	}
	_, _ = p.Percent(0)
	prevThrottle, throttleErr := cgroup.CPUThrottleStat()

	go func() {
		ticker := time.NewTicker(cpuSampleInterval)
		defer ticker.Stop()
		for range ticker.C {
			if v, err := p.Percent(0); err == nil {
				cpuPercentBits.Store(math.Float64bits(v))
			}

			// when cpu.stat can not be read (not Linux, or no permission), keep the values at 0
			curThrottle, err := cgroup.CPUThrottleStat()
			if err != nil {
				throttleErr = err
				cpuThrottledPercentBits.Store(0)
				cpuThrottledMillis.Store(0)
				continue
			}
			if throttleErr != nil {
				// the previous read failed, there is no base point for this window
				throttleErr = nil
				prevThrottle = curThrottle
				continue
			}
			percent, millis := throttleDelta(prevThrottle, curThrottle)
			prevThrottle = curThrottle
			cpuThrottledPercentBits.Store(math.Float64bits(percent))
			cpuThrottledMillis.Store(millis)
		}
	}()
}

func GetProcessInfo() ProcessInfo {
	cpuSamplerOnce.Do(startCpuSampler)

	p, _ := process.NewProcess(int32(os.Getpid()))
	pInfo := ProcessInfo{}

	// total core cpu. p.CPUPercent is not correct, use Percent instead.
	// The value comes from the background sampler, see startCpuSampler.
	//pInfo.CpuPercent, _ = p.CPUPercent()
	pInfo.CpuPercent = math.Float64frombits(cpuPercentBits.Load())
	// single core cpu
	//pInfo.CpuPercent = pInfo.CpuPercent / float64(runtime.NumCPU())

	// cpu throttle of the cgroup, from the same background sampler
	pInfo.CpuThrottledPercent = math.Float64frombits(cpuThrottledPercentBits.Load())
	pInfo.CpuThrottledMillis = cpuThrottledMillis.Load()

	// vm memory
	m, _ := p.MemoryInfo()
	pInfo.MemorySize = m.RSS / 1024 / 1024

	// threads
	pInfo.Threads = pprof.Lookup("threadcreate").Count()
	// goroutines
	pInfo.Goroutines = runtime.NumGoroutine()

	return pInfo
}
