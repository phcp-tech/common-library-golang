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

package shutdown_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/shutdown"
)

// ExampleWait shows the typical main-function usage: block until a signal or
// Trigger is received, then perform cleanup.
func ExampleWait() {
	// Trigger immediately so the example does not block.
	shutdown.Trigger()

	shutdown.Wait()
	fmt.Println("cleanup done")
	// Output:
	// cleanup done
}

// ExampleTrigger shows programmatic shutdown — e.g. from a /shutdown HTTP
// endpoint or a metrics failure handler.
func ExampleTrigger() {
	shutdown.Trigger() // idempotent: safe to call multiple times
	shutdown.Trigger()
	fmt.Println("ok")
	// Output:
	// ok
}
