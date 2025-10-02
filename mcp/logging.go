package mcp

// LogLevel represents the severity of a log message (RFC 5424)
type LogLevel string

// Log level constants following RFC 5424 severity levels
const (
	LogLevelDebug     LogLevel = "debug"
	LogLevelInfo      LogLevel = "info"
	LogLevelNotice    LogLevel = "notice"
	LogLevelWarning   LogLevel = "warning"
	LogLevelError     LogLevel = "error"
	LogLevelCritical  LogLevel = "critical"
	LogLevelAlert     LogLevel = "alert"
	LogLevelEmergency LogLevel = "emergency"
)

// LoggingCapability indicates whether the server supports logging
type LoggingCapability struct{}

// SetLevelRequest represents a request to set the minimum log level
type SetLevelRequest struct {
	Level LogLevel `json:"level"`
}

// LogMessage represents a log message notification
type LogMessage struct {
	Level  LogLevel               `json:"level"`            // Severity level
	Logger string                 `json:"logger,omitempty"` // Optional logger name
	Data   map[string]interface{} `json:"data"`             // Structured log data
}

// Value returns a numeric value for comparison (higher = more severe)
func (l LogLevel) Value() int {
	switch l {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelNotice:
		return 2
	case LogLevelWarning:
		return 3
	case LogLevelError:
		return 4
	case LogLevelCritical:
		return 5
	case LogLevelAlert:
		return 6
	case LogLevelEmergency:
		return 7
	default:
		return 0
	}
}

// ShouldLog returns true if this level should be logged given the minimum level
func (l LogLevel) ShouldLog(minLevel LogLevel) bool {
	return l.Value() >= minLevel.Value()
}
