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

package metrics_test

import (
	"encoding/json"
	"fmt"

	"github.com/phcp-tech/common-library-golang/metrics"
)

// ExampleNameValue shows how to construct a NameValue pair directly.
// NameValue is the element type returned by [GetMetrics].
func ExampleNameValue() {
	nv := metrics.NameValue{Name: "goroutines", Value: "42"}
	fmt.Printf("%s=%s\n", nv.Name, nv.Value)
	// Output:
	// goroutines=42
}

// ExampleGetMetrics shows how to call GetMetrics and look up a specific
// entry by name. GetMetrics samples CPU usage over one second internally,
// so this example is compiled but not executed during automated tests.
func ExampleGetMetrics() {
	snapshot := metrics.GetMetrics()

	for _, nv := range snapshot {
		if nv.Name == "goroutines" {
			fmt.Println("goroutines:", nv.Value)
			break
		}
	}
}

// ExampleGetMetrics_json shows how to serialize the metrics snapshot to JSON,
// suitable for a /metrics or /health HTTP endpoint response.
func ExampleGetMetrics_json() {
	snapshot := metrics.GetMetrics()

	b, err := json.Marshal(snapshot)
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}
	fmt.Println(string(b) != "")
}

// ExampleGetMetrics_iterate shows how to iterate over all metrics entries and
// print each name-value pair — useful for logging the full snapshot at startup.
func ExampleGetMetrics_iterate() {
	snapshot := metrics.GetMetrics()

	for _, nv := range snapshot {
		fmt.Printf("%-14s %s\n", nv.Name, nv.Value)
	}
}
