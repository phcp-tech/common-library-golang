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

package httpserver_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/phcp-tech/common-library-golang/httpserver"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// freePort finds a free TCP port on localhost by briefly opening and closing a listener.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	return port
}

// newTestHandler returns a minimal http.Handler with a /ping route.
func newTestHandler() http.Handler {
	r := gin.New()
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// -----------------------------------------------------------------------
// Config — resolve() fills in defaults for zero-value fields
// -----------------------------------------------------------------------

func TestConfig_Resolve_DefaultsApplied(t *testing.T) {
	cfg := httpserver.Config{Port: "8080"} // all durations are zero
	runner := httpserver.NewHttpServer(cfg)
	// Verify NewHttpServer does not panic and returns a non-nil IRunner.
	if runner == nil {
		t.Fatal("NewHttpServer() returned nil runner")
	}
}

func TestConfig_Resolve_CustomValuesPreserved(t *testing.T) {
	cfg := httpserver.Config{
		Port:         "9090",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	runner := httpserver.NewHttpServer(cfg)
	if runner == nil {
		t.Fatal("NewHttpServer() returned nil runner")
	}
}

// -----------------------------------------------------------------------
// NewHttpServer — runner selection
// -----------------------------------------------------------------------

func TestNewHttpServer_ReturnsNonNilRunner(t *testing.T) {
	runner := httpserver.NewHttpServer(httpserver.Config{Port: "8080"})
	if runner == nil {
		t.Fatal("NewHttpServer() returned nil")
	}
}

// -----------------------------------------------------------------------
// httpRunner — plain HTTP: start, verify, shutdown
// -----------------------------------------------------------------------

func TestHTTPRunner_PlainHTTP_StartAndShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real HTTP server test in short mode (-short)")
	}
	port := freePort(t)
	runner := httpserver.NewHttpServer(httpserver.Config{Port: port})

	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Start(newTestHandler())
	}()

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	waitUntilReachable(t, addr)

	// Verify the server responds.
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", addr))
	if err != nil {
		t.Fatalf("GET /ping: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /ping status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Start goroutine should return nil after ErrServerClosed is filtered out.
	select {
	case startErr := <-errCh:
		if startErr != nil {
			t.Errorf("Start returned error after Shutdown: %v", startErr)
		}
	case <-time.After(3 * time.Second):
		t.Error("Start goroutine did not return within 3 seconds of Shutdown")
	}
}

// -----------------------------------------------------------------------
// httpRunner — TLS with invalid cert files returns error
// -----------------------------------------------------------------------

func TestHTTPRunner_TLS_InvalidCerts_ReturnsError(t *testing.T) {
	port := freePort(t)
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port:    port,
		CrtFile: "/no/such/cert.crt",
		KeyFile: "/no/such/key.key",
	})
	// ListenAndServeTLS returns immediately when cert files don't exist.
	err := runner.Start(newTestHandler())
	if err == nil {
		t.Error("Start with non-existent TLS cert files should return error, got nil")
	}
}

// -----------------------------------------------------------------------
// httpRunner — invalid port returns error
// -----------------------------------------------------------------------

func TestHTTPRunner_InvalidPort_ReturnsError(t *testing.T) {
	// Port 99999 > 65535 is out of range on all platforms.
	runner := httpserver.NewHttpServer(httpserver.Config{Port: "99999"})
	err := runner.Start(newTestHandler())
	if err == nil {
		t.Error("Start with out-of-range port should return error, got nil")
	}
}

// -----------------------------------------------------------------------
// httpRunner — Shutdown before Start is a no-op
// -----------------------------------------------------------------------

func TestHTTPRunner_ShutdownBeforeStart_IsNoOp(t *testing.T) {
	runner := httpserver.NewHttpServer(httpserver.Config{Port: "8080"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown before Start should return nil, got %v", err)
	}
}

// -----------------------------------------------------------------------
// Config.WriteTimeout / NoWriteTimeout — end-to-end through a real
// http.Server, not just checking the resolved struct field. A handler
// writes a first chunk, flushes, sleeps past the configured WriteTimeout,
// then writes a second chunk — proving whether the server actually killed
// the connection mid-response, not just whether resolve() stored the right
// duration.
// -----------------------------------------------------------------------

// slowHandler returns a handler that flushes "first-chunk", sleeps for
// delay, then writes "second-chunk". Whether the client ever receives the
// second chunk is exactly what distinguishes a killed connection from an
// unlimited one.
func slowHandler(delay time.Duration) http.Handler {
	r := gin.New()
	r.GET("/slow", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte("first-chunk")) //nolint:errcheck
		c.Writer.Flush()
		time.Sleep(delay)
		c.Writer.Write([]byte("second-chunk")) //nolint:errcheck
	})
	return r
}

func TestHTTPRunner_WriteTimeout_KillsSlowWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real HTTP server test in short mode (-short)")
	}
	port := freePort(t)
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port:         port,
		WriteTimeout: 100 * time.Millisecond, // shorter than slowHandler's 300ms delay
	})

	errCh := make(chan error, 1)
	go func() { errCh <- runner.Start(slowHandler(300 * time.Millisecond)) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		runner.Shutdown(ctx) //nolint:errcheck
	}()

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	waitUntilReachable(t, addr)

	resp, err := http.Get(fmt.Sprintf("http://%s/slow", addr))
	if err != nil {
		t.Fatalf("GET /slow: %v", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)

	// A WriteTimeout firing mid-response is observed client-side as a
	// truncated body (the connection was cut before "second-chunk" was
	// written), typically surfaced as a read error since the chunked
	// encoding never gets its terminating chunk.
	if readErr == nil && strings.Contains(string(body), "second-chunk") {
		t.Fatalf("expected the connection to be killed before the second chunk, got full body %q with no error", body)
	}
}

func TestHTTPRunner_NoWriteTimeout_AllowsSlowWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real HTTP server test in short mode (-short)")
	}
	port := freePort(t)
	runner := httpserver.NewHttpServer(httpserver.Config{
		Port:         port,
		WriteTimeout: httpserver.NoWriteTimeout,
	})

	errCh := make(chan error, 1)
	go func() { errCh <- runner.Start(slowHandler(300 * time.Millisecond)) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		runner.Shutdown(ctx) //nolint:errcheck
	}()

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	waitUntilReachable(t, addr)

	resp, err := http.Get(fmt.Sprintf("http://%s/slow", addr))
	if err != nil {
		t.Fatalf("GET /slow: %v", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("ReadAll: %v (want no error — NoWriteTimeout should let the slow write complete)", readErr)
	}
	if got := string(body); got != "first-chunksecond-chunk" {
		t.Fatalf("body = %q, want %q (both chunks, uninterrupted)", got, "first-chunksecond-chunk")
	}
}

// waitUntilReachable polls addr until a TCP connection succeeds, up to 2 seconds.
func waitUntilReachable(t *testing.T, addr string) {
	t.Helper()
	for range 40 {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("server did not become reachable within 2 seconds")
}
