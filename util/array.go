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

package util

import (
	"errors"
	"reflect"
	"strings"
)

// Contain reports whether obj is present in target. target may be a slice, array, or map.
// It returns an error when obj is not found. Originally provided by zsbfree:
// https://www.cnblogs.com/zsbfree/archive/2013/05/23/3094993.html
func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

// URIContain reports whether path contains any of the strings in nonjwtapi.
// It iterates over each entry in nonjwtapi and checks if it is a substring of path,
// returning an error when no match is found.
func URIContain(path string, nonjwtapi []string) (bool, error) {
	for i := 0; i < len(nonjwtapi); i++ {
		if strings.Contains(path, nonjwtapi[i]) {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}
