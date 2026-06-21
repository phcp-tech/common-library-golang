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

package health_test

import (
	"context"
	"fmt"

	"github.com/phcp-tech/common-library-golang/health"
)

// ExampleCheck shows how to compose multiple health checkers for a /health endpoint.
// Each checker reports its own component name and status independently.
func ExampleCheck() {
	mockDB := func(ctx context.Context) health.Result {
		return health.Result{Name: "postgres", Status: health.StatusHealthy}
	}
	mockRedis := func(ctx context.Context) health.Result {
		return health.Result{Name: "redis", Status: health.StatusHealthy}
	}

	results := health.Check(context.Background(), mockDB, mockRedis)
	for _, r := range results {
		fmt.Printf("%s: %d\n", r.Name, r.Status)
	}
	// Output:
	// postgres: 1
	// redis: 1
}
