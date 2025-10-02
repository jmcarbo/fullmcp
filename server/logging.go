package server

import (
	"context"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// LoggingManager handles log message notifications
type LoggingManager struct {
	mu       sync.RWMutex
	minLevel mcp.LogLevel
	enabled  bool
	sender   LogSender
}

// LogSender sends log notifications to the client
type LogSender func(msg *mcp.LogMessage) error

// NewLoggingManager creates a new logging manager
func NewLoggingManager() *LoggingManager {
	return &LoggingManager{
		minLevel: mcp.LogLevelInfo, // Default to info level
		enabled:  false,            // Disabled until client sets level
	}
}

// EnableLogging returns an option that enables logging capability
func EnableLogging() Option {
	return func(s *Server) {
		s.logging = NewLoggingManager()
	}
}

// SetLevel sets the minimum log level
func (lm *LoggingManager) SetLevel(level mcp.LogLevel) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.minLevel = level
	lm.enabled = true
}

// SetSender sets the function to send log notifications
func (lm *LoggingManager) SetSender(sender LogSender) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.sender = sender
}

// Log sends a log message if the level is sufficient
func (lm *LoggingManager) Log(level mcp.LogLevel, logger string, data map[string]interface{}) error {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if !lm.enabled {
		return nil // Logging not enabled yet
	}

	if !level.ShouldLog(lm.minLevel) {
		return nil // Level below threshold
	}

	if lm.sender == nil {
		return nil // No sender configured
	}

	msg := &mcp.LogMessage{
		Level:  level,
		Logger: logger,
		Data:   data,
	}

	return lm.sender(msg)
}

// Server logging methods

// Log sends a log message
func (s *Server) Log(level mcp.LogLevel, logger string, data map[string]interface{}) error {
	if s.logging == nil {
		return nil
	}
	return s.logging.Log(level, logger, data)
}

// LogDebug logs a debug message
func (s *Server) LogDebug(logger string, data map[string]interface{}) error {
	return s.Log(mcp.LogLevelDebug, logger, data)
}

// LogInfo logs an info message
func (s *Server) LogInfo(logger string, data map[string]interface{}) error {
	return s.Log(mcp.LogLevelInfo, logger, data)
}

// LogWarning logs a warning message
func (s *Server) LogWarning(logger string, data map[string]interface{}) error {
	return s.Log(mcp.LogLevelWarning, logger, data)
}

// LogError logs an error message
func (s *Server) LogError(logger string, data map[string]interface{}) error {
	return s.Log(mcp.LogLevelError, logger, data)
}

// SetLogLevel handles the logging/setLevel request
func (s *Server) SetLogLevel(_ context.Context, level mcp.LogLevel) error {
	if s.logging == nil {
		return &mcp.Error{
			Code:    mcp.MethodNotFound,
			Message: "logging not enabled on this server",
		}
	}
	s.logging.SetLevel(level)
	return nil
}
