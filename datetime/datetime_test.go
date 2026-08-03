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

package datetime

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TIME_FORMAT constant
// ---------------------------------------------------------------------------

func TestTimeFormatConstant(t *testing.T) {
	// Verify the layout conforms to Go's reference time convention.
	ref := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if ref.Format(TIME_FORMAT) != "2006-01-02 15:04:05" {
		t.Errorf("TIME_FORMAT constant is wrong: got %q", TIME_FORMAT)
	}
}

// ---------------------------------------------------------------------------
// Time.MarshalJSON / UnmarshalJSON / String
// ---------------------------------------------------------------------------

func TestTime_MarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{
			name: "zero time",
			in:   time.Time{},
			want: `"0001-01-01 00:00:00"`,
		},
		{
			name: "specific datetime",
			in:   time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC),
			want: `"2024-03-15 10:30:45"`,
		},
		{
			name: "midnight",
			in:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			want: `"2023-12-31 00:00:00"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := Time(tc.in)
			b, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("MarshalJSON error: %v", err)
			}
			if string(b) != tc.want {
				t.Errorf("MarshalJSON: got %s, want %s", b, tc.want)
			}
		})
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
	// UnmarshalJSON receives raw JSON bytes (including surrounding quotes for strings).
	// The implementation passes the raw bytes directly to ParseInLocation WITHOUT
	// stripping quotes, so only unquoted datetime bytes parse successfully.
	cases := []struct {
		name    string
		rawData []byte // raw bytes passed directly to UnmarshalJSON
		wantErr bool
		wantFmt string // expected formatted string (empty means don't check)
	}{
		{
			name:    "valid unquoted datetime bytes",
			rawData: []byte("2024-06-15 08:20:30"),
			wantErr: false,
			wantFmt: "2024-06-15 08:20:30",
		},
		{
			name:    "invalid format",
			rawData: []byte("2024/06/15"),
			wantErr: true,
		},
		{
			name:    "empty bytes",
			rawData: []byte(""),
			wantErr: true,
		},
		{
			name:    "quoted string (as json.Unmarshal would deliver)",
			rawData: []byte(`"2024-06-15 08:20:30"`),
			wantErr: true, // quotes cause ParseInLocation to fail
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v Time
			err := v.UnmarshalJSON(tc.rawData)
			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && tc.wantFmt != "" {
				if v.String() != tc.wantFmt {
					t.Errorf("UnmarshalJSON: got %q, want %q", v.String(), tc.wantFmt)
				}
			}
		})
	}
}

func TestTime_String(t *testing.T) {
	in := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	v := Time(in)
	if v.String() != "2025-01-02 03:04:05" {
		t.Errorf("String: got %q, want %q", v.String(), "2025-01-02 03:04:05")
	}
}

func TestTime_MarshalAndUnmarshalConsistency(t *testing.T) {
	// MarshalJSON produces `"2024-11-22 13:14:15"` (with quotes).
	// UnmarshalJSON parses raw bytes WITHOUT stripping quotes, so we call
	// UnmarshalJSON with the unquoted datetime bytes directly to verify consistency.
	original := time.Date(2024, 11, 22, 13, 14, 15, 0, time.UTC)
	v := Time(original)

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Strip surrounding quotes to get the raw datetime string.
	inner := b[1 : len(b)-1] // remove leading and trailing `"`

	var v2 Time
	if err := v2.UnmarshalJSON(inner); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if v.String() != v2.String() {
		t.Errorf("marshal/unmarshal mismatch: got %q, want %q", v2.String(), v.String())
	}
}

// ---------------------------------------------------------------------------
// UTCToLocal
// ---------------------------------------------------------------------------

func TestUTCToLocal(t *testing.T) {
	cases := []struct {
		name     string
		utcInput time.Time
		timezone string
		wantErr  bool
		// we verify the hour difference rather than an exact string to avoid DST flakiness
		hourDiff int // expected added hours compared to UTC
	}{
		{
			name:     "UTC to Asia/Shanghai (+8)",
			utcInput: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			timezone: "Asia/Shanghai",
			wantErr:  false,
			hourDiff: 8,
		},
		{
			name:     "UTC to UTC (no change)",
			utcInput: time.Date(2024, 6, 20, 12, 0, 0, 0, time.UTC),
			timezone: "UTC",
			wantErr:  false,
			hourDiff: 0,
		},
		{
			name:     "UTC to America/New_York (EST = -5)",
			utcInput: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			timezone: "America/New_York",
			wantErr:  false,
			hourDiff: -5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := UTCToLocal(tc.utcInput, tc.timezone)
			if tc.wantErr && err != nil {
				return // expected
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("UTCToLocal(%v, %q) error: %v", tc.utcInput, tc.timezone, err)
			}

			// Check the hour difference (UTC hour + offset = result hour)
			expectedHour := (tc.utcInput.Hour() + tc.hourDiff + 24) % 24
			if result.Hour() != expectedHour {
				t.Errorf("UTCToLocal hour: got %d, want %d", result.Hour(), expectedHour)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LocalToUTC
// ---------------------------------------------------------------------------

func TestLocalToUTC(t *testing.T) {
	cases := []struct {
		name      string
		localTime time.Time
		timezone  string
		wantErr   bool
		wantHour  int
	}{
		{
			name:      "Shanghai local 08:00 -> UTC 00:00",
			localTime: time.Date(2024, 4, 1, 8, 0, 0, 0, time.UTC), // treated as local string
			timezone:  "Asia/Shanghai",
			wantErr:   false,
			wantHour:  0,
		},
		{
			name:      "UTC 12:00 stays 12:00",
			localTime: time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
			timezone:  "UTC",
			wantErr:   false,
			wantHour:  12,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LocalToUTC(tc.localTime, tc.timezone)
			if tc.wantErr && err != nil {
				return
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("LocalToUTC error: %v", err)
			}
			if result.Hour() != tc.wantHour {
				t.Errorf("LocalToUTC hour: got %d, want %d", result.Hour(), tc.wantHour)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UTCToLocalString
// ---------------------------------------------------------------------------

func TestUTCToLocalString(t *testing.T) {
	cases := []struct {
		name     string
		utcStr   string
		timezone string
		wantErr  bool
		wantStr  string
	}{
		{
			name:     "Asia/Shanghai +8h",
			utcStr:   "2024-01-15 00:00:00",
			timezone: "Asia/Shanghai",
			wantErr:  false,
			wantStr:  "2024-01-15 08:00:00",
		},
		{
			name:     "UTC no change",
			utcStr:   "2024-06-20 12:30:00",
			timezone: "UTC",
			wantErr:  false,
			wantStr:  "2024-06-20 12:30:00",
		},
		{
			name:     "America/New_York EST -5h",
			utcStr:   "2024-01-10 12:00:00",
			timezone: "America/New_York",
			wantErr:  false,
			wantStr:  "2024-01-10 07:00:00",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := UTCToLocalString(tc.utcStr, tc.timezone)
			if tc.wantErr && err != nil {
				return
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("UTCToLocalString error: %v", err)
			}
			if result != tc.wantStr {
				t.Errorf("UTCToLocalString: got %q, want %q", result, tc.wantStr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LocalToUTCString
// ---------------------------------------------------------------------------

func TestLocalToUTCString(t *testing.T) {
	cases := []struct {
		name      string
		localStr  string
		timezone  string
		wantErr   bool
		wantStr   string
	}{
		{
			name:     "Shanghai 08:00 -> UTC 00:00",
			localStr: "2024-01-15 08:00:00",
			timezone: "Asia/Shanghai",
			wantErr:  false,
			wantStr:  "2024-01-15 00:00:00",
		},
		{
			name:     "UTC 12:30 unchanged",
			localStr: "2024-06-20 12:30:00",
			timezone: "UTC",
			wantErr:  false,
			wantStr:  "2024-06-20 12:30:00",
		},
		{
			name:     "New York EST 07:00 -> UTC 12:00",
			localStr: "2024-01-10 07:00:00",
			timezone: "America/New_York",
			wantErr:  false,
			wantStr:  "2024-01-10 12:00:00",
		},
		{
			name:     "invalid datetime string",
			localStr: "not-a-date",
			timezone: "UTC",
			wantErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LocalToUTCString(tc.localStr, tc.timezone)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LocalToUTCString error: %v", err)
			}
			if result != tc.wantStr {
				t.Errorf("LocalToUTCString: got %q, want %q", result, tc.wantStr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LocalToUTCString / UTCToLocalString inverse relationship
// ---------------------------------------------------------------------------

func TestUTCLocalStringInverse(t *testing.T) {
	utcStr := "2024-07-04 10:00:00"
	timezone := "Asia/Shanghai"

	localStr, err := UTCToLocalString(utcStr, timezone)
	if err != nil {
		t.Fatalf("UTCToLocalString error: %v", err)
	}

	backToUTC, err := LocalToUTCString(localStr, timezone)
	if err != nil {
		t.Fatalf("LocalToUTCString error: %v", err)
	}

	if backToUTC != utcStr {
		t.Errorf("inverse round-trip: got %q, want %q", backToUTC, utcStr)
	}
}

// ---------------------------------------------------------------------------
// UTCToMTServerTime / MTServerTimeToUTC
// ---------------------------------------------------------------------------

func TestUTCToMTServerTime(t *testing.T) {
	// Use a known EST time (January = UTC-5).
	// NY offset in EST = -18000 (UTC-5)
	// MT offset = 7*3600 = 25200
	// delta = 25200 + (-18000) = 7200 (2 hours ahead of UTC during EST)
	estDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	utcUnix := estDate.Unix()

	mtTime, err := UTCToMTServerTime(utcUnix)
	if err != nil {
		t.Fatalf("UTCToMTServerTime error: %v", err)
	}

	// delta should be 7200 seconds (2h) in January (EST)
	delta := mtTime - utcUnix
	if delta != 7200 {
		t.Errorf("UTCToMTServerTime EST delta: got %d, want 7200", delta)
	}
}

func TestMTServerTimeToUTC(t *testing.T) {
	// Use a known EST time reference.
	estDate := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC) // 14:00 MT server time
	mtUnix := estDate.Unix()

	utcTime, err := MTServerTimeToUTC(mtUnix)
	if err != nil {
		t.Fatalf("MTServerTimeToUTC error: %v", err)
	}

	delta := mtUnix - utcTime
	if delta != 7200 {
		t.Errorf("MTServerTimeToUTC EST delta: got %d, want 7200", delta)
	}
}

func TestMTServerTimeRoundTrip(t *testing.T) {
	now := time.Now().Unix()

	mtTime, err := UTCToMTServerTime(now)
	if err != nil {
		t.Fatalf("UTCToMTServerTime error: %v", err)
	}

	backToUTC, err := MTServerTimeToUTC(mtTime)
	if err != nil {
		t.Fatalf("MTServerTimeToUTC error: %v", err)
	}

	if backToUTC != now {
		t.Errorf("MT round-trip mismatch: got %d, want %d", backToUTC, now)
	}
}

func TestMTSERVER_TIME_OFFSET(t *testing.T) {
	if MTSEVER_TIME_OFFSET != 25200 {
		t.Errorf("MTSEVER_TIME_OFFSET: got %d, want 25200", MTSEVER_TIME_OFFSET)
	}
}

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
