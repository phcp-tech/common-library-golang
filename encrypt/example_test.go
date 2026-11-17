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

// package encrypt_test demonstrates the public API from a caller's perspective.
package encrypt_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/encrypt"
)

// ExampleEncryptPassword hashes a password with a per-user salt. Store the
// salt next to the hash; to check a login, hash the entered password with the
// same salt and compare the two strings.
func ExampleEncryptPassword() {
	stored := encrypt.EncryptPassword("password", "salt")
	fmt.Println(stored)
	fmt.Println(encrypt.EncryptPassword("password", "salt") == stored)
	fmt.Println(encrypt.EncryptPassword("wrong", "salt") == stored)
	// Output:
	// xeR41ZKIyEGqUw22hFxMjZYok6ABzk4RpJY4c6qYE0o=
	// true
	// false
}
