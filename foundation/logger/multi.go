//foundation/logger/multi.go

package logger

import (
	"context"
	"log/slog"
)

// multiHandler is a Handler that multiplexes its calls to a set of handlers.
type multiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler constructs a new multiHandler.
func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{
		handlers: handlers,
	}
}

// Enabled reports whether any of the underlying handlers are enabled for the given level.
func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// WithAttrs returns a new multiHandler whose attributes consist of
// the receiver's attributes followed by attrs, applied to all underlying handlers.
func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return newMultiHandler(newHandlers...)
}

// WithGroup returns a new multiHandler with the given group appended
// to the receiver's existing groups, applied to all underlying handlers.
func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return newMultiHandler(newHandlers...)
}

// Handle processes the Record for all underlying handlers.
// Calls are concurrent and non-blocking to the main application thread.
func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	// The original record MUST be cloned before sending to concurrent handlers.
	rCopy := r.Clone()

	for _, handler := range h.handlers {
		if handler.Enabled(ctx, rCopy.Level) {

			// We DO NOT use a WaitGroup. We simply fire a goroutine and let it complete.
			// This makes log writes non-blocking.
			go func(h slog.Handler, r slog.Record) {
				// Errors are logged internally (in elkWriter) and not returned.
				// This ensures the main application is never blocked by log I/O.
				h.Handle(ctx, r)
			}(handler, rCopy)
		}
	}

	// Important: We return immediately, allowing the main application to continue.
	return nil
}
