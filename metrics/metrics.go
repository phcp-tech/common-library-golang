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

package metrics

import (
	"runtime"
	"strconv"
	"time"

	"github.com/phcp-tech/common-library-golang/cgroup"
	"github.com/phcp-tech/common-library-golang/datetime"
	"github.com/phcp-tech/common-library-golang/process"
)

// NameValue holds a name-value pair used in metrics and chart data formats.
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var _startTime = time.Now().Unix()

// GetMetrics collects and returns a snapshot of key runtime and system metrics as a slice of
// NameValue pairs. The snapshot includes CPU and memory usage, thread and goroutine counts,
// GOMAXPROCS, NumCPU, cgroup CPU/memory request and limit values, and the process uptime age.
func GetMetrics() []NameValue {
	pInfo := process.GetProcessInfo()
	cpuRequest, _ := cgroup.CPURequestMilli()
	cpuLimit, _ := cgroup.CPULimitMilli()
	memoryLimit, _ := cgroup.MemoryLimitBytes()

	return []NameValue{
		// cpu percent
		{Name: "cpuPercent", Value: strconv.Itoa(int(pInfo.CpuPercent))},
		// CFS throttle of the latest 5 seconds: 0 = not held back by the CPU limit (always 0 without a limit)
		{Name: "cpuThrottledPercent", Value: strconv.Itoa(int(pInfo.CpuThrottledPercent))},
		{Name: "cpuThrottledMillis", Value: strconv.FormatInt(pInfo.CpuThrottledMillis, 10)},
		// CPU limits from cgroup (if any)
		{Name: "cpuRequest", Value: strconv.Itoa(cpuRequest)},
		{Name: "cpuLimit", Value: strconv.Itoa(cpuLimit)},

		// memory size of the process
		{Name: "memorySize", Value: strconv.Itoa(int(pInfo.MemorySize))},
		// Memory limits from cgroup (if any)
		// memoryRequest is not output: K8s does not write the memory request into the cgroup (memory.min and
		// memory.low are 0 unless MemoryQoS is on), so the value was always 0 and looked like "no request is set".
		// See cgroup.MemoryRequestBytes.
		{Name: "memoryLimit", Value: strconv.FormatInt(memoryLimit, 10)},

		// NumCPU, GOMAXPROCS and thread/goroutine counts
		{Name: "numCPU", Value: strconv.Itoa(runtime.NumCPU())},
		{Name: "threads", Value: strconv.Itoa(int(pInfo.Threads))},
		{Name: "goroutines", Value: strconv.Itoa(int(pInfo.Goroutines))},
		{Name: "gomaxprocs", Value: strconv.Itoa(runtime.GOMAXPROCS(0))},
		// process age
		{Name: "age", Value: datetime.GetAge(_startTime)},
	}
}
