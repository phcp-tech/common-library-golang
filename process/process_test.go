package process

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetProcessInfo(t *testing.T) {
	Convey("Test GetProcessInfo", t, func() {
		Convey("Test it does not block the caller", func() {
			// the old code called p.Percent(time.Second), which slept 1 second in every call
			GetProcessInfo() // the first call starts the sampler
			start := time.Now()
			for i := 0; i < 5; i++ {
				GetProcessInfo()
			}
			So(time.Since(start), ShouldBeLessThan, 500*time.Millisecond)
		})

		Convey("Test process fields", func() {
			info := GetProcessInfo()
			So(info.MemorySize, ShouldBeGreaterThan, 0)
			So(info.Goroutines, ShouldBeGreaterThan, 0)
			So(info.Threads, ShouldBeGreaterThan, 0)
			So(info.CpuPercent, ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

// TestGetProcessInfoCpuPercent keeps one goroutine busy for longer than one sample window
// and checks that CpuPercent shows it. It takes about 6 seconds, so it is skipped in short mode.
func TestGetProcessInfoCpuPercent(t *testing.T) {
	if testing.Short() {
		t.Skip("takes more than one sample window")
	}

	Convey("Test CpuPercent follows the CPU usage of the latest sample window", t, func() {
		GetProcessInfo() // start the sampler and set the base point

		stop := make(chan struct{})
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				select {
				case <-stop:
					return
				default:
				}
			}
		}()

		// wait for more than one window, so at least one full window is sampled while the goroutine is busy
		time.Sleep(cpuSampleInterval + 1500*time.Millisecond)
		busy := GetProcessInfo().CpuPercent
		close(stop)
		<-done

		// one busy goroutine is about 100%. Use a low threshold, because a shared CI machine may give less CPU.
		So(busy, ShouldBeGreaterThan, 30)
	})
}
