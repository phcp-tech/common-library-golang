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

// package file_test demonstrates the public API from a caller's perspective.
package file_test

import (
	"errors"
	"fmt"

	"github.com/phcp-tech/common-library-golang/file"
)

// ExamplePolicy_Check checks two uploads against a document policy.
func ExamplePolicy_Check() {
	docs := file.Policy{Types: file.DocumentTypes(), MaxSize: 20_000_000, Disposition: "attachment"}

	ext, err := docs.Check("application/pdf", 123456, "report.pdf")
	fmt.Println(ext, err)

	_, err = docs.Check("application/msword", 123456, "old.doc")
	fmt.Println(errors.Is(err, file.ErrInvalid))
	// Output:
	// pdf <nil>
	// true
}
