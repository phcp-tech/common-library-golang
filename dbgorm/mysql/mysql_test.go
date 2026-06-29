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

package mysql_test

import (
	"errors"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/mysql"
)

func TestDialectorFromStructuredFields(t *testing.T) {
	dialector, err := mysql.Dialector(&mysql.Config{
		Host:     "localhost",
		Port:     "3306",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("dialector: %v", err)
	}
	if dialector == nil {
		t.Fatalf("expected dialector")
	}
}

func TestDialectorRequiresStructuredFields(t *testing.T) {
	if _, err := mysql.Dialector(&mysql.Config{}); err != dbgorm.ErrMissingConfig {
		t.Fatalf("expected ErrMissingConfig, got %v", err)
	}
}

func TestDialectorDoesNotConnect(t *testing.T) {
	dialector, err := mysql.Dialector(&mysql.Config{
		Host:     "localhost",
		Port:     "3306",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("dialector: %v", err)
	}
	if dialector == nil {
		t.Fatalf("expected dialector")
	}
}

// TestNewMySQL_ErrMissingConfig covers the Dialector-error path in NewMySQL.
func TestNewMySQL_ErrMissingConfig(t *testing.T) {
	_, err := mysql.NewMySQL(&mysql.Config{}) // all required fields empty
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("NewMySQL empty config: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Error covers the error-return path in InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := mysql.InitDefault(&mysql.Config{}) // Dialector fails → error propagated
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: want ErrMissingConfig, got %v", err)
	}
}
