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

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/ringbuf"

	"github.com/natefinch/lumberjack"
	slogFormatter "github.com/samber/slog-formatter"
)

const (
	// log file rotation settings; can be overridden by environment config
	defaultLogFileMaxSize    = 100  // defaultLogFileMaxSize is the maximum size of a single log file in megabytes before rotation.
	defaultLogFileMaxBackups = 100  // defaultLogFileMaxBackups is the maximum number of rotated log backup files to retain.
	defaultLogFileMaxAge     = 0    // defaultLogFileMaxAge is the maximum number of days to retain old log files; 0 means never delete.
	defaultLogFileCompress   = true // defaultLogFileCompress indicates whether to compress rotated log files using gzip.
)

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
	logLevel    *slog.LevelVar // default is INFO
	logFile     *lumberjack.Logger
	asyncWriter *asyncWriter // non-nil only when file logging is enabled
}

// Instance is the package-level singleton Log, initialized exactly once on first access.
// It reads log level and file rotation settings from the environment configuration.
var Instance = sync.OnceValue(func() *Log {
	// Log instance
	log := &Log{
		Logger:   nil,
		logLevel: &slog.LevelVar{},
		logFile:  &lumberjack.Logger{},
	}

	// set log level from env. If there no env, set to info
	level := "info"
	if env.Env() != nil {
		level = strings.ToLower(strings.TrimSpace(env.Env().String("log.level")))
	}
	switch level {
	case "error":
		log.logLevel.Set(slog.LevelError)
	case "warn":
		log.logLevel.Set(slog.LevelWarn)
	case "debug":
		log.logLevel.Set(slog.LevelDebug)
	default:
		log.logLevel.Set(slog.LevelInfo)
	}

	// slog options
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     log.logLevel,
	}

	// write to log file for rotate. If there no env, write to stdout
	if env.Env() != nil && env.Env().Bool("log.writefile") {
		log.logFile.Filename = env.Env().String("log.path")
		log.logFile.LocalTime = false

		log.logFile.MaxSize = env.Env().Int("log.file.max.size")
		if log.logFile.MaxSize == 0 {
			log.logFile.MaxSize = defaultLogFileMaxSize
		}
		log.logFile.MaxBackups = env.Env().Int("log.file.max.backups")
		if log.logFile.MaxBackups == 0 {
			log.logFile.MaxBackups = defaultLogFileMaxBackups
		}
		log.logFile.MaxAge = env.Env().Int("log.file.max.age")
		if log.logFile.MaxAge == 0 {
			log.logFile.MaxAge = defaultLogFileMaxAge
		}
		if env.Env().Exists("log.file.compress") {
			log.logFile.Compress = env.Env().Bool("log.file.compress")
		} else {
			log.logFile.Compress = defaultLogFileCompress
		}

		log.asyncWriter = newAsyncWriter(log.logFile)
		log.Logger = slog.New(
			slogFormatter.NewFormatterHandler(
				slogFormatter.TimezoneConverter(time.UTC),
				slogFormatter.TimeFormatter(time.RFC3339, nil),
			)(
				slog.NewJSONHandler(log.asyncWriter, opts),
			),
		)
	} else {
		log.Logger = slog.New(
			slogFormatter.NewFormatterHandler(
				slogFormatter.TimezoneConverter(time.UTC),
				slogFormatter.TimeFormatter(time.RFC3339, nil),
			)(
				slog.NewJSONHandler(os.Stdout, opts),
			),
		)
	}

	// Routes standard library log output through this structured handler.
	// After this call, stdlib log.Print / log.Printf calls are emitted as structured JSON entries.
	slog.SetDefault(log.Logger)

	return log
})

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

// CloseLogFile closes the underlying rotating log file used by the singleton logger.
// It should be called during graceful shutdown to flush and release file resources.
func CloseLogFile() {
	l := Instance()
	if l.asyncWriter != nil {
		l.asyncWriter.Close() // drain ring buffer before closing the file
	}
	l.logFile.Close()
}

// Debug logs a message at DEBUG level using the singleton logger.
func Debug(msg string) {
	Instance().Logger.Debug(msg)
}

// Info logs a message at INFO level using the singleton logger.
func Info(msg string) {
	Instance().Logger.Info(msg)
}

// Warn logs a message at WARN level using the singleton logger.
func Warn(msg string) {
	Instance().Logger.Warn(msg)
}

// Error logs a message at ERROR level using the singleton logger.
func Error(msg string) {
	Instance().Logger.Error(msg)
}

// Debugf logs a formatted message at DEBUG level using the singleton logger.
func Debugf(format string, args ...any) {
	Instance().Logger.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted message at INFO level using the singleton logger.
func Infof(format string, args ...any) {
	Instance().Logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted message at WARN level using the singleton logger.
func Warnf(format string, args ...any) {
	Instance().Logger.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted message at ERROR level using the singleton logger.
func Errorf(format string, args ...any) {
	Instance().Logger.Error(fmt.Sprintf(format, args...))
}

// log functions with structured key-value fields
func DebugWith(msg string, args ...any) { Instance().Logger.Debug(msg, args...) }
func InfoWith(msg string, args ...any)  { Instance().Logger.Info(msg, args...) }
func WarnWith(msg string, args ...any)  { Instance().Logger.Warn(msg, args...) }
func ErrorWith(msg string, args ...any) { Instance().Logger.Error(msg, args...) }
