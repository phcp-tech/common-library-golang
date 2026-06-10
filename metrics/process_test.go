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
	"testing"
)

// ---------------------------------------------------------------------------
// process.go – GetProcessInfo
// ---------------------------------------------------------------------------

func TestGetProcessInfo(t *testing.T) {
	info := GetProcessInfo()
	// Goroutines must be at least 1 (the test goroutine itself)
	if info.Goroutines < 1 {
		t.Errorf("GetProcessInfo Goroutines = %d; want >= 1", info.Goroutines)
	}
	// Threads must be at least 1
	if info.Threads < 1 {
		t.Errorf("GetProcessInfo Threads = %d; want >= 1", info.Threads)
	}
	// CpuPercent must be >= 0
	if info.CpuPercent < 0 {
		t.Errorf("GetProcessInfo CpuPercent = %f; want >= 0", info.CpuPercent)
	}
}
