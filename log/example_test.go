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

// package log_test uses an external test package to demonstrate the public API
// from a caller's perspective, exactly as an application would use it.
package log_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/log"
)

// Example demonstrates the two usage modes of this package.
//
// Stdout mode: call [InitLog] with no arguments for stdout output at INFO level.
// File mode: call [InitLog] once at startup with a Config, and defer [Close]
// to flush all buffered entries before the process exits.
func Example() {
	log.InitLog(&log.Config{
		Level:      "info",
		FilePath:   "/var/log/app.log",
		MaxSizeMB:  100,
		MaxBackups: 7,
		MaxAgeDays: 30,
		Compress:   true,
	})
	defer log.Close()

	log.Info("application started")
	log.InfoWith("request handled",
		"method", "GET",
		"path", "/api/v1/users",
		"status", 200,
	)
}

// ExampleInitLog_file shows file logging mode: pass a *Config with FilePath set.
// For stdout at default INFO level, InitLog need not be called at all.
// To customise level without file output, use InitLog(&log.Config{Level: "debug"}).
func ExampleInitLog_file() {
	log.InitLog(&log.Config{
		Level:      "info",
		FilePath:   "/var/log/app.log",
		MaxSizeMB:  100,  // rotate after 100 MB
		MaxBackups: 7,    // keep at most 7 backup files
		MaxAgeDays: 30,   // delete backups older than 30 days; 0 = never delete
		Compress:   true, // gzip rotated files to save disk space
	})
	defer log.Close()

	log.Info("file logging enabled")
}

// ExampleClose shows that Close should be deferred immediately
// after InitLog. It flushes the async ring buffer and closes the underlying
// rotating file. It is a no-op when the logger writes to stdout.
func ExampleClose() {
	log.InitLog(&log.Config{
		Level:    "info",
		FilePath: "/var/log/app.log",
	})
	defer log.Close()

	log.Info("shutting down")
}

// ExampleInfo shows basic message logging to stdout.
// InitLog must always be called before any log function.
func ExampleInfo() {
	log.InitLog() // stdout + INFO
	log.Info("application started")
}

// ExampleSetLevel shows how to change the active log level at runtime without
// restarting the application. It also demonstrates the error returned for an
// unrecognised level string.
func ExampleSetLevel() {
	_ = log.SetLevel("debug") // enable verbose output temporarily
	_ = log.SetLevel("info")  // restore normal level

	if err := log.SetLevel("invalid"); err != nil {
		fmt.Println(err)
	}
	// Output: unknown log level: invalid
}

// ExampleDebugWith shows how to attach structured context to a debug log entry.
// args must alternate between a string key and its value; mismatched pairs
// are handled gracefully by the underlying slog handler.
func ExampleDebugWith() {
	log.DebugWith("cache lookup",
		"key", "user:42",
		"hit", false,
		"duration_us", 38,
	)
}

// ExampleInfoWith shows structured logging with additional key-value context.
// args must alternate between a string key and its value; mismatched pairs
// are handled gracefully by the underlying slog handler.
func ExampleInfoWith() {
	log.InfoWith("user login",
		"user_id", 42,
		"ip", "192.168.1.1",
		"success", true,
	)
}

// ExampleWarnWith shows how to attach structured context to a warning entry.
func ExampleWarnWith() {
	log.WarnWith("rate limit approaching",
		"user_id", 99,
		"requests", 980,
		"limit", 1000,
	)
}

// ExampleErrorWith shows how to attach error context and diagnostics to an
// error log entry.
func ExampleErrorWith() {
	log.ErrorWith("database query failed",
		"table", "orders",
		"duration_ms", 1523,
		"error", "connection timeout",
	)
}
