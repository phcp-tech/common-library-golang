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

package dbsqlx

import "time"

const (
	MaxOpenConns    = 100              // MaxOpenConns is the maximum number of open database connections in the pool.
	MaxIdleConns    = 25               // MaxIdleConns is the maximum number of idle database connections (1/4 of MaxOpenConns).
	ConnMaxLifetime = 60 * time.Minute // ConnMaxLifetime is the maximum lifetime of a database connection.
	ConnMaxIdletime = 10 * time.Minute // ConnMaxIdletime is the maximum idle lifetime of a database connection.
)

// PoolConfig contains connection-pool tuning parameters shared by all dbsqlx
// adapters. Zero-value fields fall back to the package-level defaults above.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
	ConnMaxIdletime int // minutes
}
