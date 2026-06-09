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

// Package shutdown provides two primitives for application shutdown:
//
//   - [Wait] blocks the calling goroutine until an OS signal or [Trigger] is called.
//   - [Trigger] unblocks [Wait] programmatically from any goroutine.
//
// Typical usage in main:
//
//	// start services ...
//	shutdown.Wait()
//	// cleanup ...
package shutdown

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	stopCh = make(chan struct{})
	once   sync.Once
)

// Wait blocks until SIGINT, SIGTERM, SIGHUP, SIGQUIT is received, or [Trigger]
// is called. After Wait returns the caller should perform cleanup and exit.
func Wait() {
	// on Windows
	//signal.Notify(ch, os.Interrupt)
	// on Linux. SIGINT for ctrl_c, SIGTERM for kill pid
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		slog.Info(fmt.Sprintf("Received shutdown signal: %v", sig))
	case <-stopCh:
		slog.Info("Programmatic shutdown triggered")
	}
}

// Trigger unblocks [Wait] programmatically — useful for a /shutdown HTTP
// endpoint or a metrics failure handler. Safe to call multiple times and from
// any goroutine. Has no effect when the process is forcibly terminated by the OS.
func Trigger() {
	once.Do(func() { close(stopCh) })
}
