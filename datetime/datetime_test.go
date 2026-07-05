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

// Package datetime contains functions for calculating age based on timestamps.
package datetime

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// GetAge
// ---------------------------------------------------------------------------

func TestGetAge(t *testing.T) {
	// Verify the output format matches "Xd Xh Xm Xs"
	now := time.Now().Unix()

	t.Run("just_now", func(t *testing.T) {
		age := GetAge(now)
		matched, _ := regexp.MatchString(`^\d+d \d+h \d+m \d+s$`, age)
		if !matched {
			t.Errorf("GetAge format wrong: %q", age)
		}
	})

	t.Run("one_day_ago", func(t *testing.T) {
		oneDayAgo := now - 86400
		age := GetAge(oneDayAgo)
		if !strings.HasPrefix(age, "1d ") {
			t.Errorf("GetAge for 1 day ago should start with '1d ', got: %q", age)
		}
	})

	t.Run("one_hour_ago", func(t *testing.T) {
		oneHourAgo := now - 3600
		age := GetAge(oneHourAgo)
		if !strings.HasPrefix(age, "0d 1h ") {
			t.Errorf("GetAge for 1 hour ago should start with '0d 1h ', got: %q", age)
		}
	})

	t.Run("one_minute_ago", func(t *testing.T) {
		oneMinAgo := now - 60
		age := GetAge(oneMinAgo)
		if !strings.HasPrefix(age, "0d 0h 1m ") {
			t.Errorf("GetAge for 1 minute ago should start with '0d 0h 1m ', got: %q", age)
		}
	})

	t.Run("two_days_and_three_hours", func(t *testing.T) {
		past := now - 2*86400 - 3*3600
		age := GetAge(past)
		if !strings.HasPrefix(age, "2d 3h ") {
			t.Errorf("GetAge for 2d3h ago should start with '2d 3h ', got: %q", age)
		}
	})
}

func TestGetAge_ZeroTime(t *testing.T) {
	// Very old timestamp; result should have large day count.
	age := GetAge(0)
	matched, _ := regexp.MatchString(`^\d+d \d+h \d+m \d+s$`, age)
	if !matched {
		t.Errorf("GetAge format wrong for zero time: %q", age)
	}
}
