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

package component_test

import (
	"fmt"

	storageComp "github.com/phcp-tech/common-library-golang/storage/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads the storage.* keys from env during Init() and stores the result in
// storage.Default(). Init fails when a required key is empty or still
// "TOBE_REPLACED"; it does not contact the store, so call storage.Default().Check
// to test connectivity.
//
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)).
//	    Add(logComp.Component()).
//	    Add(storageComp.Component()).
//	    Run()
func ExampleComponent() {
	c := storageComp.Component()
	fmt.Println(c.Name())
	// Output:
	// storage
}
