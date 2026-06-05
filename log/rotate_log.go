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

package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/phcp-tech/common-library-golang/ringbuf"

	"github.com/natefinch/lumberjack"
	slogFormatter "github.com/samber/slog-formatter"
)

const (
	// log file rotation defaults; used when the corresponding Config field is zero.
	defaultLogFileMaxSize    = 100 // defaultLogFileMaxSize is the maximum size of a single log file in megabytes before rotation.
	defaultLogFileMaxBackups = 100 // defaultLogFileMaxBackups is the maximum number of rotated log backup files to retain.
	defaultLogFileMaxAge     = 0   // defaultLogFileMaxAge is the maximum number of days to retain old log files; 0 means never delete.
)

// Config holds all configuration for the logger.
// Call Init with a Config before the first log call.
// If FilePath is empty, logs are written to stdout.
type Config struct {
	Level      string // "debug"|"info"|"warn"|"error"; default "info"
	FilePath   string // if non-empty, enables rotating file logging
	MaxSizeMB  int    // max size of a single log file in MB; default 100
	MaxBackups int    // max number of rotated backup files to retain; default 100
	MaxAgeDays int    // max age in days before deletion; 0 means never delete
	Compress   bool   // compress rotated files with gzip
}

// asyncWriter wraps an io.Writer with a RingMPSC buffer so callers return immediately
// and the actual Write is handled by the consumer goroutine.
type asyncWriter struct {
	rb *ringbuf.RingMPSC[[]byte]
}

func newAsyncWriter(w io.Writer) *asyncWriter {
	aw := &asyncWriter{}
	aw.rb = ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[[]byte]{
		ProcessFunc: func(b []byte) { w.Write(b) }, //nolint:errcheck
	})
	return aw
}

// Write copies p before pushing to avoid a race with slog's internal buffer reuse.
func (aw *asyncWriter) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	aw.rb.Push(buf)
	return len(p), nil
}

// Close drains the ring buffer and waits for all pending writes to complete.
func (aw *asyncWriter) Close() {
	aw.rb.Close()
}

// Log holds the structured logger, its dynamic level, and the optional rotating log file writer.
type Log struct {
	Logger      *slog.Logger
	logLevel    *slog.LevelVar     // default is INFO
	logFile     *lumberjack.Logger // non-nil only when file logging is enabled
	asyncWriter *asyncWriter       // non-nil only when file logging is enabled
}

var (
	instance *Log
	once     sync.Once
)

// InitLog configures the logger. It must be called once at application startup
// before any log function (Debug/Info/Warn/Error and their variants).
// Subsequent calls have no effect (only the first call takes effect).
// If called with no arguments, or with a nil Config pointer, the logger
// writes to stdout at INFO level. Pass a non-nil Config with FilePath set to enable
// rotating file logging.
func InitLog(cfg ...*Config) {
	once.Do(func() {
		c := Config{}
		if len(cfg) > 0 && cfg[0] != nil {
			c = *cfg[0]
		}
		instance = newLog(c)
	})
}

// Instance returns the singleton Log initialised by InitLog.
// Returns nil if InitLog has not been called.
func Instance() *Log {
	return instance
}

func newLog(cfg Config) *Log {
	l := &Log{logLevel: &slog.LevelVar{}}

	// resolve log level
	switch strings.ToLower(strings.TrimSpace(cfg.Level)) {
	case "error":
		l.logLevel.Set(slog.LevelError)
	case "warn":
		l.logLevel.Set(slog.LevelWarn)
	case "debug":
		l.logLevel.Set(slog.LevelDebug)
	default:
		l.logLevel.Set(slog.LevelInfo)
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     l.logLevel,
	}

	var w io.Writer
	if cfg.FilePath != "" {
		// resolve log file rotation settings, using defaults for any zero values
		maxSize := defaultLogFileMaxSize
		maxBackups := defaultLogFileMaxBackups
		maxAge := defaultLogFileMaxAge

		// If the user explicitly set a value (greater than 0), use it instead of the default.
		if cfg.MaxSizeMB > 0 {
			maxSize = cfg.MaxSizeMB
		}
		if cfg.MaxBackups > 0 {
			maxBackups = cfg.MaxBackups
		}
		if cfg.MaxAgeDays > 0 {
			maxAge = cfg.MaxAgeDays
		}
		l.logFile = &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   cfg.Compress,
			LocalTime:  false,
		}
		l.asyncWriter = newAsyncWriter(l.logFile)
		w = l.asyncWriter
	} else {
		w = os.Stdout
	}

	l.Logger = slog.New(
		slogFormatter.NewFormatterHandler(
			slogFormatter.TimezoneConverter(time.UTC),
			slogFormatter.TimeFormatter(time.RFC3339, nil),
		)(
			slog.NewJSONHandler(w, opts),
		),
	)

	// Routes standard library log output through this structured handler.
	// After this call, stdlib log.Print / log.Printf calls are emitted as structured JSON entries.
	slog.SetDefault(l.Logger)

	return l
}

// SetLevel dynamically changes the log level of the singleton logger.
// Accepted values are "error", "warn", "info", and "debug" (case-insensitive).
// Returns an error if the provided level string is unrecognised.
func SetLevel(level string) error {
	level = strings.ToLower(level)
	switch level {
	case "error":
		Instance().logLevel.Set(slog.LevelError)
	case "warn":
		Instance().logLevel.Set(slog.LevelWarn)
	case "info":
		Instance().logLevel.Set(slog.LevelInfo)
	case "debug":
		Instance().logLevel.Set(slog.LevelDebug)
	default:
		return fmt.Errorf("unknown log level: %s", level)
	}

	return nil
}

// CloseLogFile flushes all pending async log entries and closes the underlying
// rotating log file. It should be called during graceful shutdown.
// No-op when logging to stdout.
func CloseLogFile() {
	l := Instance()
	if l.asyncWriter != nil {
		l.asyncWriter.Close() // drain ring buffer before closing the file
	}
	if l.logFile != nil {
		l.logFile.Close() //nolint:errcheck
	}
}

// Debug logs a message at DEBUG level using the singleton logger.
func Debug(msg string) { Instance().Logger.Debug(msg) }

// Info logs a message at INFO level using the singleton logger.
func Info(msg string) { Instance().Logger.Info(msg) }

// Warn logs a message at WARN level using the singleton logger.
func Warn(msg string) { Instance().Logger.Warn(msg) }

// Error logs a message at ERROR level using the singleton logger.
func Error(msg string) { Instance().Logger.Error(msg) }

// Debugf logs a formatted message at DEBUG level using the singleton logger.
func Debugf(format string, args ...any) { Instance().Logger.Debug(fmt.Sprintf(format, args...)) }

// Infof logs a formatted message at INFO level using the singleton logger.
func Infof(format string, args ...any) { Instance().Logger.Info(fmt.Sprintf(format, args...)) }

// Warnf logs a formatted message at WARN level using the singleton logger.
func Warnf(format string, args ...any) { Instance().Logger.Warn(fmt.Sprintf(format, args...)) }

// Errorf logs a formatted message at ERROR level using the singleton logger.
func Errorf(format string, args ...any) { Instance().Logger.Error(fmt.Sprintf(format, args...)) }

// DebugWith logs a message at DEBUG level with additional structured key-value fields.
// args must be alternating key-value pairs, e.g. DebugWith("msg", "key", value).
func DebugWith(msg string, args ...any) { Instance().Logger.Debug(msg, args...) }

// InfoWith logs a message at INFO level with additional structured key-value fields.
// args must be alternating key-value pairs, e.g. InfoWith("msg", "key", value).
func InfoWith(msg string, args ...any) { Instance().Logger.Info(msg, args...) }

// WarnWith logs a message at WARN level with additional structured key-value fields.
// args must be alternating key-value pairs, e.g. WarnWith("msg", "key", value).
func WarnWith(msg string, args ...any) { Instance().Logger.Warn(msg, args...) }

// ErrorWith logs a message at ERROR level with additional structured key-value fields.
// args must be alternating key-value pairs, e.g. ErrorWith("msg", "key", value).
func ErrorWith(msg string, args ...any) { Instance().Logger.Error(msg, args...) }
