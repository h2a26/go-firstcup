package logger

import (
	"context"
	"log/slog"
	"time"
)

// Level represents different logging levels.
type Level slog.Level

// UnmarshalText implements encoding.TextUnmarshaler to parse a log level from text.
func (l *Level) UnmarshalText(text []byte) error {
	switch string(text) {
	case "debug", "DEBUG":
		*l = LevelDebug
	case "info", "INFO", "": // make the zero value useful
		*l = LevelInfo
	case "warn", "WARN":
		*l = LevelWarn
	case "error", "ERROR":
		*l = LevelError
	default:
		*l = LevelInfo // Default to info for unknown levels
	}
	return nil
}

// A set of possible logging levels.
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// ELKConfig contains the configuration settings for publishing logs to Logstash/ELK.
type ELKConfig struct {
	Host     string
	Port     string
	Protocol string
}

// Config contains the configuration settings for the logger.
type Config struct {
	MinLevel  Level
	Service   string
	TraceIDFn TraceIDFn
	Outputs   []string // e.g., "console", "elk"
	ELKConfig ELKConfig
	Events    Events
}

// Record represents the data that is being logged.
type Record struct {
	Time       time.Time
	Message    string
	Level      Level
	Attributes map[string]any
}

func toRecord(r slog.Record) Record {
	atts := make(map[string]any, r.NumAttrs())

	f := func(attr slog.Attr) bool {
		atts[attr.Key] = attr.Value.Any()
		return true
	}
	r.Attrs(f)

	return Record{
		Time:       r.Time,
		Message:    r.Message,
		Level:      Level(r.Level),
		Attributes: atts,
	}
}

// EventFn is a function to be executed when configured against a log level.
type EventFn func(ctx context.Context, r Record)

// Events contains an assignment of an event function to a log level.
type Events struct {
	Debug EventFn
	Info  EventFn
	Warn  EventFn
	Error EventFn
}
