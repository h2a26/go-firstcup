// foundation/logger/logger.go
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TraceIDFn represents a function that can return the trace id from
// the specified context.
type TraceIDFn func(ctx context.Context) string

// Logger represents a logger for logging information.
type Logger struct {
	discard   bool
	handler   slog.Handler
	traceIDFn TraceIDFn
}

// NewWithConfig constructs a new log for application use, supporting multiple outputs.
func NewWithConfig(cfg Config) *Logger {

	// Function to clean up the source code location when logged.
	replaceAttrFn := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// -------------------------------------------------------------------------
	// 1. Build the list of core output handlers.

	var coreHandlers []slog.Handler

	for _, output := range cfg.Outputs {
		switch strings.ToLower(output) {
		case "console":
			// Console Handler: JSON output to os.Stdout
			handler := slog.NewJSONHandler(io.Writer(os.Stdout), &slog.HandlerOptions{
				AddSource:   true,
				Level:       slog.Level(cfg.MinLevel),
				ReplaceAttr: replaceAttrFn,
			})
			coreHandlers = append(coreHandlers, handler)

		case "elk":
			// ELK Handler: JSON output over TCP to Logstash
			handler := newELKHandler(cfg.ELKConfig, cfg.MinLevel, replaceAttrFn)
			coreHandlers = append(coreHandlers, handler)
		}
	}

	// If no outputs are configured, default to a discard handler.
	if len(coreHandlers) == 0 {
		coreHandlers = append(coreHandlers, slog.NewJSONHandler(io.Discard, nil))
	}

	// -------------------------------------------------------------------------
	// 2. Assemble the final handler chain.

	// The base handler is the MultiHandler, sending the record to all configured outputs.
	handler := slog.Handler(newMultiHandler(coreHandlers...))

	// If events are to be processed, wrap the base handler around the custom log handler.
	// This ensures events are triggered BEFORE logs are written to any output.
	if cfg.Events.Debug != nil || cfg.Events.Info != nil || cfg.Events.Warn != nil || cfg.Events.Error != nil {
		handler = newLogHandler(handler, cfg.Events)
	}

	// Attributes to add to every log.
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(cfg.Service)},
	}

	// Add those attributes and capture the final handler.
	handler = handler.WithAttrs(attrs)

	// -------------------------------------------------------------------------
	// 3. Return the final Logger.

	return &Logger{
		discard:   false, // Discard check is now handled by coreHandlers logic
		handler:   handler,
		traceIDFn: cfg.TraceIDFn,
	}
}

// NewStdLogger returns a standard library Logger that wraps the slog Logger.
func NewStdLogger(logger *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(logger.handler, slog.Level(level))
}

// Debug logs at LevelDebug with the given context.
func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelDebug, 3, msg, args...)
}

// Debugc logs the information at the specified call stack position.
func (log *Logger) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelDebug, caller, msg, args...)
}

// Info logs at LevelInfo with the given context.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelInfo, 3, msg, args...)
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelInfo, caller, msg, args...)
}

// Warn logs at LevelWarn with the given context.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelWarn, 3, msg, args...)
}

// Warnc logs the information at the specified call stack position.
func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelWarn, caller, msg, args...)
}

// Error logs at LevelError with the given context.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelError, 3, msg, args...)
}

// Errorc logs the information at the specified call stack position.
func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelError, caller, msg, args...)
}

func (log *Logger) write(ctx context.Context, level Level, caller int, msg string, args ...any) {
	slogLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	if log.traceIDFn != nil {
		args = append(args, "trace_id", log.traceIDFn(ctx))
	}
	r.Add(args...)

	log.handler.Handle(ctx, r)
}
