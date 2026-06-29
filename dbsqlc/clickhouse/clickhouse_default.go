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

package clickhouse

import (
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var (
	instance driver.Conn
	once     sync.Once
)

// InitDefault initializes the default singleton ClickHouse client.
// If you want to use any other instance, please call NewClickHouse.
func InitDefault(conf *Config) error {
	var err error
	once.Do(func() {
		instance, err = NewClickHouse(conf)
	})
	return err
}

// Default returns the default singleton instance of the ClickHouse client.
func Default() driver.Conn {
	return instance
}
