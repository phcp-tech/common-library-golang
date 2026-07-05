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

package math

import (
	"math/rand"
	"time"
)

// Random returns a pseudo-random integer in the half-open interval [min, max).
// A new random source seeded with the current time is created on every call.
func Random(min int, max int) int {
	rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand1.Intn(max-min) + min
}
