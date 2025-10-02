package server

import (
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// ProgressTracker manages progress notifications for long-running operations
type ProgressTracker struct {
	mu     sync.RWMutex
	sender ProgressSender
}

// ProgressSender sends progress notifications to the client
type ProgressSender func(notification *mcp.ProgressNotification) error

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{}
}

// SetSender sets the function to send progress notifications
func (pt *ProgressTracker) SetSender(sender ProgressSender) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.sender = sender
}

// Notify sends a progress notification
func (pt *ProgressTracker) Notify(token mcp.ProgressToken, progress float64, total *float64) error {
	return pt.NotifyWithMessage(token, progress, total, "")
}

// NotifyWithMessage sends a progress notification with a descriptive message
func (pt *ProgressTracker) NotifyWithMessage(token mcp.ProgressToken, progress float64, total *float64, message string) error {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if pt.sender == nil {
		return nil // No sender configured
	}

	notification := &mcp.ProgressNotification{
		ProgressToken: token,
		Progress:      progress,
		Total:         total,
		Message:       message,
	}

	return pt.sender(notification)
}

// ProgressContext wraps a context with progress tracking
type ProgressContext struct {
	Token   mcp.ProgressToken
	Tracker *ProgressTracker
}

// NewProgressContext creates a new progress context
func NewProgressContext(token mcp.ProgressToken, tracker *ProgressTracker) *ProgressContext {
	return &ProgressContext{
		Token:   token,
		Tracker: tracker,
	}
}

// Update sends a progress update
func (pc *ProgressContext) Update(progress float64, total *float64) error {
	if pc.Tracker == nil {
		return nil
	}
	return pc.Tracker.Notify(pc.Token, progress, total)
}

// Server progress methods

// NotifyProgress sends a progress notification
func (s *Server) NotifyProgress(token mcp.ProgressToken, progress float64, total *float64) error {
	if s.progress == nil {
		return nil
	}
	return s.progress.Notify(token, progress, total)
}

// WithProgress configures progress tracking
func WithProgress() Option {
	return func(s *Server) {
		s.progress = NewProgressTracker()
	}
}
