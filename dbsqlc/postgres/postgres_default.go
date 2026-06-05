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

package postgres

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	instance *pgxpool.Pool
	once     sync.Once
)

// InitDefault initializes the default singleton PostgreSQL connection pool.
func InitDefault(conf *Config) error {
	var err error
	once.Do(func() {
		var pool *pgxpool.Pool
		pool, err = NewPostgres(conf)
		if err == nil {
			instance = pool
		}
	})
	return err
}

// Default returns the default *pgxpool.Pool instance created by InitDefault.
func Default() *pgxpool.Pool {
	return instance
}
