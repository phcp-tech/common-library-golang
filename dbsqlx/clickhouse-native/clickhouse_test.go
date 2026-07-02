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

// Tests must be run with: go test -gcflags=all=-l ./dbsqlx/clickhouse-native/...
// The -gcflags=all=-l flag disables inlining, required by gomonkey for function patching.
package clickhousenative

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	clickhousedriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/agiledragon/gomonkey/v2"
)

// mockConn is a test double for driver.Conn.
type mockConn struct {
	pingErr     error
	pingCalled  bool
	closeCalled bool
}

func (m *mockConn) Contributors() []string                                           { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error)                    { return nil, nil }
func (m *mockConn) Select(_ context.Context, _ any, _ string, _ ...any) error        { return nil }
func (m *mockConn) Query(_ context.Context, _ string, _ ...any) (driver.Rows, error) { return nil, nil }
func (m *mockConn) QueryRow(_ context.Context, _ string, _ ...any) driver.Row        { return nil }
func (m *mockConn) PrepareBatch(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}
func (m *mockConn) Exec(_ context.Context, _ string, _ ...any) error { return nil }
func (m *mockConn) AsyncInsert(_ context.Context, _ string, _ bool, _ ...any) error {
	return nil
}
func (m *mockConn) Ping(_ context.Context) error { m.pingCalled = true; return m.pingErr }
func (m *mockConn) Stats() driver.Stats          { return driver.Stats{} }
func (m *mockConn) Close() error                 { m.closeCalled = true; return nil }

// resetSingleton resets the singleton state for test isolation.
// sync.Once must not be copied — reset via composite literal assignment.
func resetSingleton(t *testing.T) {
	t.Helper()
	if instance != nil {
		instance.Close() //nolint:errcheck
	}
	instance = nil
	once = sync.Once{}
	t.Cleanup(func() {
		if instance != nil {
			instance.Close() //nolint:errcheck
		}
		instance = nil
		once = sync.Once{}
		once.Do(func() {}) // mark once as "already run" to match pre-test state
	})
}

func TestDefault_BeforeInit_IsNil(t *testing.T) {
	resetSingleton(t)
	if Default() != nil {
		t.Skip("singleton already initialised in this process — cannot test pre-init state")
	}
}

// ─── Singleton lifecycle ──────────────────────────────────────────────────────

func TestSingleton_Lifecycle(t *testing.T) {
	resetSingleton(t)

	mock := &mockConn{}
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(_ *clickhousedriver.Options) (driver.Conn, error) {
		return mock, nil
	})
	defer patches.Reset()

	err := InitDefault(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})
	if err != nil {
		t.Fatalf("InitDefault error = %v", err)
	}

	if Default() == nil {
		t.Fatal("Default() is nil after successful InitDefault")
	}

	// Second InitDefault is a no-op (sync.Once).
	if err := InitDefault(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "other",
		Username: "nobody",
		Password: "nopass",
	}); err != nil {
		t.Errorf("second InitDefault should return nil; got %v", err)
	}
}

// ─── InitDefault — Open error ────────────────────────────────────────────────

func TestInitDefault_OpenError(t *testing.T) {
	resetSingleton(t)

	openErr := errors.New("connection refused")
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(_ *clickhousedriver.Options) (driver.Conn, error) {
		return nil, openErr
	})
	defer patches.Reset()

	err := InitDefault(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})
	if err == nil {
		t.Fatal("InitDefault should return error when Open fails")
	}
	if !errors.Is(err, openErr) {
		t.Errorf("InitDefault error = %v, want %v", err, openErr)
	}
	if Default() != nil {
		t.Error("Default() should be nil when InitDefault fails")
	}
}

// ─── NewClickHouse ───────────────────────────────────────────────────────────

func TestNewClickHouse_OpenError(t *testing.T) {
	openErr := errors.New("dial tcp: connection refused")
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(_ *clickhousedriver.Options) (driver.Conn, error) {
		return nil, openErr
	})
	defer patches.Reset()

	conn, err := NewClickHouse(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})
	if conn != nil {
		t.Error("NewClickHouse should return nil conn on Open error")
	}
	if !errors.Is(err, openErr) {
		t.Errorf("NewClickHouse error = %v, want %v", err, openErr)
	}
}

func TestNewClickHouse_Success(t *testing.T) {
	mock := &mockConn{}
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(_ *clickhousedriver.Options) (driver.Conn, error) {
		return mock, nil
	})
	defer patches.Reset()

	conn, err := NewClickHouse(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})
	if err != nil {
		t.Fatalf("NewClickHouse error = %v", err)
	}
	if conn == nil {
		t.Fatal("NewClickHouse returned nil conn")
	}
}

func TestNewClickHouse_CustomPoolSettings(t *testing.T) {
	var gotMaxOpen, gotMaxIdle int
	var gotLifetime time.Duration
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(opt *clickhousedriver.Options) (driver.Conn, error) {
		gotMaxOpen = opt.MaxOpenConns
		gotMaxIdle = opt.MaxIdleConns
		gotLifetime = opt.ConnMaxLifetime
		return &mockConn{}, nil
	})
	defer patches.Reset()

	_, _ = NewClickHouse(&Config{
		Host:            "127.0.0.1",
		Port:            "19000",
		Database:        "testdb",
		Username:        "nobody",
		Password:        "nopass",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30, // minutes
	})

	if gotMaxOpen != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", gotMaxOpen)
	}
	if gotMaxIdle != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", gotMaxIdle)
	}
	if gotLifetime != 30*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 30m", gotLifetime)
	}
}

func TestNewClickHouse_UsesGlobalDefaults(t *testing.T) {
	var gotMaxOpen, gotMaxIdle int
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(opt *clickhousedriver.Options) (driver.Conn, error) {
		gotMaxOpen = opt.MaxOpenConns
		gotMaxIdle = opt.MaxIdleConns
		return &mockConn{}, nil
	})
	defer patches.Reset()

	_, _ = NewClickHouse(&Config{
		Host:     "127.0.0.1",
		Port:     "19000",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})

	if gotMaxOpen != 100 {
		t.Errorf("MaxOpenConns = %d, want 100 (global default)", gotMaxOpen)
	}
	if gotMaxIdle != 25 {
		t.Errorf("MaxIdleConns = %d, want 25 (global default)", gotMaxIdle)
	}
}

func TestNewClickHouse_PassesConfigToOptions(t *testing.T) {
	var gotAddr, gotDatabase, gotUsername, gotPassword string
	patches := gomonkey.ApplyFunc(clickhousedriver.Open, func(opt *clickhousedriver.Options) (driver.Conn, error) {
		if len(opt.Addr) > 0 {
			gotAddr = opt.Addr[0]
		}
		gotDatabase = opt.Auth.Database
		gotUsername = opt.Auth.Username
		gotPassword = opt.Auth.Password
		return &mockConn{}, nil
	})
	defer patches.Reset()

	_, _ = NewClickHouse(&Config{
		Host:     "ch-host",
		Port:     "9440",
		Database: "mydb",
		Username: "admin",
		Password: "secret",
	})

	if gotAddr != "ch-host:9440" {
		t.Errorf("Addr = %q, want %q", gotAddr, "ch-host:9440")
	}
	if gotDatabase != "mydb" {
		t.Errorf("Database = %q, want %q", gotDatabase, "mydb")
	}
	if gotUsername != "admin" {
		t.Errorf("Username = %q, want %q", gotUsername, "admin")
	}
	if gotPassword != "secret" {
		t.Errorf("Password = %q, want %q", gotPassword, "secret")
	}
}
