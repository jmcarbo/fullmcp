package server

import (
	"context"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// CancellationManager manages request cancellations
type CancellationManager struct {
	mu             sync.RWMutex
	cancelFuncs    map[interface{}]context.CancelFunc
	cancellationCh chan *mcp.CancelledNotification
}

// NewCancellationManager creates a new cancellation manager
func NewCancellationManager() *CancellationManager {
	return &CancellationManager{
		cancelFuncs:    make(map[interface{}]context.CancelFunc),
		cancellationCh: make(chan *mcp.CancelledNotification, 10),
	}
}

// Register registers a cancellable context for a request
func (cm *CancellationManager) Register(requestID interface{}, cancel context.CancelFunc) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cancelFuncs[requestID] = cancel
}

// Unregister removes a request from cancellation tracking
func (cm *CancellationManager) Unregister(requestID interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cancelFuncs, requestID)
}

// Cancel cancels a request by ID
func (cm *CancellationManager) Cancel(requestID interface{}, _ string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cancel, exists := cm.cancelFuncs[requestID]; exists {
		cancel()
		delete(cm.cancelFuncs, requestID)
		return true
	}

	return false
}

// HandleCancellation handles a cancellation notification
func (cm *CancellationManager) HandleCancellation(notification *mcp.CancelledNotification) {
	cm.Cancel(notification.RequestID, notification.Reason)
}

// WithCancellation enables cancellation support
func WithCancellation() Option {
	return func(s *Server) {
		s.cancellation = NewCancellationManager()
	}
}

// Server cancellation methods

// CancelRequest cancels a request by ID
func (s *Server) CancelRequest(requestID interface{}, reason string) bool {
	if s.cancellation == nil {
		return false
	}
	return s.cancellation.Cancel(requestID, reason)
}

// RegisterCancellable registers a cancellable context
func (s *Server) RegisterCancellable(requestID interface{}, cancel context.CancelFunc) {
	if s.cancellation != nil {
		s.cancellation.Register(requestID, cancel)
	}
}

// UnregisterCancellable removes a request from cancellation tracking
func (s *Server) UnregisterCancellable(requestID interface{}) {
	if s.cancellation != nil {
		s.cancellation.Unregister(requestID)
	}
}
