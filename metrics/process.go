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
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo holds a snapshot of the current process's resource usage.
type ProcessInfo struct {
	CpuPercent float64 // CPU usage as a percentage across all cores over the last second.
	MemorySize uint64  // Resident set size (RSS) memory usage in megabytes.
	Threads    int     // Number of OS threads created by the process.
	Goroutines int     // Number of live goroutines.
}

// GetProcessInfo returns a ProcessInfo snapshot for the currently running process,
// including CPU percentage (measured over one second), RSS memory in megabytes,
// OS thread count, and goroutine count.
func GetProcessInfo() ProcessInfo {
	p, _ := process.NewProcess(int32(os.Getpid()))
	pInfo := ProcessInfo{}

	// total core cpu. p.CPUPercent is not correct, use Percent instead
	//pInfo.CpuPercent, _ = p.CPUPercent()
	pInfo.CpuPercent, _ = p.Percent(time.Second)
	// single core cpu
	//pInfo.CpuPercent = pInfo.CpuPercent / float64(runtime.NumCPU())

	// vm memory
	m, _ := p.MemoryInfo()
	pInfo.MemorySize = m.RSS / 1024 / 1024

	// threads
	pInfo.Threads = pprof.Lookup("threadcreate").Count()
	// goroutines
	pInfo.Goroutines = runtime.NumGoroutine()

	return pInfo
}
