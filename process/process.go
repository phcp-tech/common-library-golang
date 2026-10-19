package process

import (
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// cpuSampleInterval is the length of one CPU sample window.
const cpuSampleInterval = 5 * time.Second

var (
	cpuSamplerOnce sync.Once
	// cpuPercentBits holds the float64 bits of the CPU percent of the latest sample window.
	cpuPercentBits atomic.Uint64
)

type ProcessInfo struct {
	// CpuPercent is the average CPU percent of the latest 5 seconds (100% = 1 core). It is 0 in the first 5 seconds.
	CpuPercent float64
	MemorySize uint64
	Threads    int
	Goroutines int
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
func startCpuSampler() {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return
	}
	_, _ = p.Percent(0)

	go func() {
		ticker := time.NewTicker(cpuSampleInterval)
		defer ticker.Stop()
		for range ticker.C {
			if v, err := p.Percent(0); err == nil {
				cpuPercentBits.Store(math.Float64bits(v))
			}
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

	// vm memory
	m, _ := p.MemoryInfo()
	pInfo.MemorySize = m.RSS / 1024 / 1024

	// threads
	pInfo.Threads = pprof.Lookup("threadcreate").Count()
	// goroutines
	pInfo.Goroutines = runtime.NumGoroutine()

	return pInfo
}
