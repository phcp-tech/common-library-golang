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

package loader_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/httpserver/loader"
)

// ExampleLoadFromEnv shows how LoadFromEnv is used in the composition root.
// It reads app.runmode and http.server.port from the koanf env singleton and
// returns the appropriate Runner. The caller owns the goroutine and shutdown:
//
//	runner := loader.LoadFromEnv()
//	go func() {
//	    if err := runner.Start(ginRouter); err != nil {
//	        slog.Error("Http server stopped with error", "error", err)
//	    }
//	}()
//	shutdown.Wait()
//	runner.Shutdown(ctx)
func ExampleLoadFromEnv() {
	runner := loader.LoadFromEnv()
	fmt.Println(runner != nil)
	// Output:
	// true
}
