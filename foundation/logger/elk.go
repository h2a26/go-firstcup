// foundation/logger/elk.go
package logger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	maxConnRetries = 3
	retryDelay     = 5 * time.Second
)

// elkWriter is a custom io.Writer that manages the TCP connection to Logstash.
type elkWriter struct {
	mu     sync.Mutex
	conn   net.Conn
	config ELKConfig
}

// newELKWriter creates a new writer for publishing logs to Logstash.
func newELKWriter(config ELKConfig) *elkWriter {
	w := &elkWriter{
		config: config,
	}
	// Attempt initial connection asynchronously to not block startup
	go w.connect(context.Background())
	return w
}

// connect attempts to establish a connection to the Logstash instance.
func (w *elkWriter) connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		w.conn.Close()
	}

	address := fmt.Sprintf("%s:%s", w.config.Host, w.config.Port)
	var conn net.Conn
	var err error

	for i := 0; i < maxConnRetries; i++ {
		conn, err = net.DialTimeout(w.config.Protocol, address, 5*time.Second)
		if err == nil {
			w.conn = conn
			slog.Info("logger: connected to ELK", "address", address)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			// retry
		}
	}

	slog.Error("logger: failed to connect to ELK after retries", "address", address, "error", err)
	return err
}

// Write implements the io.Writer interface. It assumes the p slice contains
// the full JSON log record (as written by slog.JSONHandler).
func (w *elkWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Reconnect if necessary
	if w.conn == nil {
		if err := w.connect(context.Background()); err != nil {
			// Fail silently if reconnection fails
			return 0, nil
		}
	}

	// 2. Add the required newline for Logstash json_lines codec.
	// Since p is an immutable slice, we must create a copy/buffer to add the newline.
	buf := bytes.NewBuffer(p)

	// Add the newline. bytes.Buffer has WriteByte
	if err := buf.WriteByte('\n'); err != nil {
		return 0, err
	}

	finalBytes := buf.Bytes()

	// 3. Write the JSON bytes to the network connection.
	if _, err := w.conn.Write(finalBytes); err != nil {
		// Connection failed during write; close and set conn to nil for reconnect on next log
		w.conn.Close()
		w.conn = nil
		// Return 0, nil to tell slog the Write was non-fatal for the application
		return 0, nil
	}

	// Return the original length of the log record (excluding the newline)
	// to satisfy the slog handler's expected behavior.
	return len(p), nil
}

// elkHandler wraps the underlying slog.Handler.
// It is structurally identical to the console handler, but uses our custom elkWriter.
type elkHandler struct {
	slog.Handler
}

// newELKHandler constructs a fully compliant slog.Handler for network output.
func newELKHandler(config ELKConfig, minLevel Level, replaceAttr func(groups []string, a slog.Attr) slog.Attr) *elkHandler {
	// 1. Create the custom writer
	writer := newELKWriter(config)

	// 2. Create the standard slog.JSONHandler, delegating the writing to our custom writer.
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.Level(minLevel),
		ReplaceAttr: replaceAttr,
	})

	// 3. Wrap the handler.
	return &elkHandler{
		Handler: handler,
	}
}
