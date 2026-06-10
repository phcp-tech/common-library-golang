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

package app

import (
	"context"

	db "github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
	"github.com/phcp-tech/common-library-golang/env"
)

// Health represents the health status of a named service or component.
type Health struct {
	Name   string `json:"name,omitempty"`
	Status int    `json:"status,omitempty"` // status code indicating component health
}

// GetHealth returns a Health snapshot reflecting the current PostgreSQL connectivity
// status. It is intended for API health-check endpoints only; CLI processes should
// not call this function because the required environment variables are unavailable
// in that context. The returned Status field is 2 when the database is reachable and
// 0 when it is not.
func GetHealth() Health {
	// redisStatus: 0=not connected, 1=connected
	// dbStatus: 0=not connected, 2=connected
	dbStatus := 0
	if db.Default() != nil && db.Default().Ping(context.Background()) == nil {
		dbStatus = 2
	}

	return Health{
		Name:   env.Env().String("app.name"),
		Status: dbStatus,
	}
}
