package logger

import (
	"context"
	"log/slog"
)

// FanoutHandler broadcasts log records to multiple handlers.
type FanoutHandler struct {
	handlers []slog.Handler
}

// NewFanoutHandler creates a new FanoutHandler with the given handlers.
func NewFanoutHandler(handlers ...slog.Handler) *FanoutHandler {
	return &FanoutHandler{handlers: handlers}
}

// Enabled returns true if at least one handler is enabled for the level.
func (h *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle handles the record by passing it to all registered handlers.
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// WithAttrs returns a new FanoutHandler with the attributes added to each handler.
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return NewFanoutHandler(newHandlers...)
}

// WithGroup returns a new FanoutHandler with the group added to each handler.
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return NewFanoutHandler(newHandlers...)
}
