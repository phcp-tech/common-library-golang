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
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// TIME_FORMAT is the standard datetime layout used throughout this package ("2006-01-02 15:04:05").
const TIME_FORMAT = "2006-01-02 15:04:05"

// Time is a custom time type that wraps the standard library time.Time and provides
// JSON marshaling/unmarshaling using the TIME_FORMAT layout.
type Time time.Time

// MarshalJSON implements the json.Marshaler interface, encoding the time value
// as a quoted string using TIME_FORMAT.
func (t Time) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format(TIME_FORMAT))
	return []byte(stamp), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface, parsing a quoted string
// in TIME_FORMAT into the Time value using the local timezone.
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(TIME_FORMAT, string(data), time.Local)
	*t = Time(now)
	return err
}

// String returns the time formatted as a string using TIME_FORMAT.
func (t Time) String() string {
	return time.Time(t).Format(TIME_FORMAT)
}

// UTCToLocal converts a UTC time.Time value to the specified timezone (e.g. "Asia/Shanghai")
// and returns the equivalent local time.
// Usage: 1, get UTC with time.Now().UTC(); 2, call In to the UTC time.
func UTCToLocal(utctime time.Time, timezone string) (time.Time, error) {
	location, _ := time.LoadLocation(timezone)
	return time.Parse(TIME_FORMAT, utctime.In(location).Format(TIME_FORMAT))
}

// LocalToUTC converts a local time.Time value from the specified timezone to UTC.
// Usage: 1, format timezone to local time; 2, call UTC to the localtime with timezone.
func LocalToUTC(localtime time.Time, timezone string) (time.Time, error) {
	location, _ := time.LoadLocation(timezone)
	if utctime, err := time.ParseInLocation(TIME_FORMAT, localtime.Format(TIME_FORMAT), location); err == nil {
		return utctime.UTC(), nil
	} else {
		return localtime, errors.Wrap(err, "LocalToUTC error")
	}
}

// UTCToLocalString converts a UTC time string (formatted with TIME_FORMAT) to a local time string
// in the specified timezone. It returns the formatted local time string or an error.
func UTCToLocalString(utctime string, timezone string) (string, error) {
	location, _ := time.LoadLocation(timezone)
	utime, _ := time.Parse(TIME_FORMAT, utctime)
	if localtime, err := time.Parse(TIME_FORMAT, utime.In(location).Format(TIME_FORMAT)); err == nil {
		return localtime.Format(TIME_FORMAT), nil
	} else {
		return utctime, errors.Wrap(err, "UTCToLocalString error")
	}
}

// LocalToUTCString converts a local time string (formatted with TIME_FORMAT) in the specified
// timezone to a UTC time string. It returns the formatted UTC time string or an error.
func LocalToUTCString(localtime string, timezone string) (string, error) {
	location, _ := time.LoadLocation(timezone)
	if utctime, err := time.ParseInLocation(TIME_FORMAT, localtime, location); err == nil {
		return utctime.UTC().Format(TIME_FORMAT), nil
	} else {
		return localtime, errors.Wrap(err, "LocalToUTCString error")
	}
}

// MTSEVER_TIME_OFFSET is the MT server time offset in seconds relative to New York Time (7 hours * 3600 seconds).
const MTSEVER_TIME_OFFSET = 25200 // 7(hours)*3600 seconds

// zoneinfo: /usr/share/zoneinfo/America/New_York in Linux,
// /usr/share/zoneinfo/zoneinfo/America/New_York in MacOS,
// Computer\HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Time Zones in Windows

// UTCToMTServerTime converts UTC time to MTServerTime
// UTC --> UTC time + 2/3*3600(2/3*3600 = 7*3600 - New York timezone offset) -> MTServerTime
// EST: UTC-5, New York timezone offset = -18000s; EDT: UTC-4, New York timezone offset = -14400s
func UTCToMTServerTime(utcTime int64) (int64, error) {
	nyTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		return 0, errors.Wrap(err, "failed to load New York timezone")
	}

	_, nyOffset := time.Unix(utcTime, 0).In(nyTZ).Zone()
	return utcTime + int64((MTSEVER_TIME_OFFSET + nyOffset)), nil
}

// MTServerTimeToUTC converts MTServerTime to UTC time
// MTServerTime --> MTServerTime - 2/3*3600(2/3*3600 = 7*3600 - New York timezone offset) --> UTC
// EST: UTC-5, New York timezone offset = -18000s; EDT: UTC-4, New York timezone offset = -14400s
func MTServerTimeToUTC(mtServerTime int64) (int64, error) {
	nyTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		return 0, errors.Wrap(err, "failed to load New York timezone")
	}

	_, nyOffset := time.Unix(mtServerTime, 0).In(nyTZ).Zone()
	return mtServerTime - int64((MTSEVER_TIME_OFFSET + nyOffset)), nil
}

// GetAge returns a human-readable duration string representing the elapsed time
// since the given Unix timestamp (in seconds), formatted as "Xd Xh Xm Xs".
func GetAge(startTime int64) string {
	diff := time.Now().Unix() - startTime
	day := diff / 86400
	hour := (diff % 86400) / 3600
	minute := (diff % 3600) / 60
	second := diff % 60
	return strconv.Itoa(int(day)) + "d " + strconv.Itoa(int(hour)) + "h " + strconv.Itoa(int(minute)) + "m " + strconv.Itoa(int(second)) + "s"
}
