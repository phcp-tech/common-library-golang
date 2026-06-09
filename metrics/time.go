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
	"strconv"
	"time"
)

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