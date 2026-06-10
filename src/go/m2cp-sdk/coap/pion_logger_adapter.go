package coap

import (
	"m2cp"

	"github.com/pion/logging"
)

// This adapter allows pion/dtls to log through the ContextPlus logging system,
// Log levels:
//   - logging.LogLevelTrace: Most verbose, includes handshake details
//   - logging.LogLevelDebug: Internal operations
//   - logging.LogLevelInfo:  Normal state transitions
//   - logging.LogLevelWarn:  Abnormal events
//   - logging.LogLevelError: Fatal errors
//   - logging.LogLevelDisabled: No logging
type PionLoggerAdapter struct {
	ctp   m2cp.ContextPlus
	level logging.LogLevel
}

// NewPionLoggerFactory creates a LoggerFactory that uses ContextPlus for logging
func NewPionLoggerFactory(ctp m2cp.ContextPlus, level logging.LogLevel) logging.LoggerFactory {
	return &pionLoggerFactory{
		ctp:   ctp,
		level: level,
	}
}

type pionLoggerFactory struct {
	ctp   m2cp.ContextPlus
	level logging.LogLevel
}

func (f *pionLoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	// Create a branched context for this scope
	scopedCtp := f.ctp.BranchWithName(scope)
	return &PionLoggerAdapter{
		ctp:   scopedCtp,
		level: f.level,
	}
}

// Trace logs at trace level
func (l *PionLoggerAdapter) Trace(msg string) {
	if l.level < logging.LogLevelTrace {
		return
	}
	l.ctp.LogDebug("[TRACE] %s", msg)
}

// Tracef logs at trace level with formatting
func (l *PionLoggerAdapter) Tracef(format string, args ...interface{}) {
	if l.level < logging.LogLevelTrace {
		return
	}
	l.ctp.LogDebug("[TRACE] "+format, args...)
}

// Debug logs at debug level
func (l *PionLoggerAdapter) Debug(msg string) {
	if l.level < logging.LogLevelDebug {
		return
	}
	l.ctp.LogDebug(msg)
}

// Debugf logs at debug level with formatting
func (l *PionLoggerAdapter) Debugf(format string, args ...interface{}) {
	if l.level < logging.LogLevelDebug {
		return
	}
	l.ctp.LogDebug(format, args...)
}

// Info logs at info level
func (l *PionLoggerAdapter) Info(msg string) {
	if l.level < logging.LogLevelInfo {
		return
	}
	l.ctp.LogInfo(msg)
}

// Infof logs at info level with formatting
func (l *PionLoggerAdapter) Infof(format string, args ...interface{}) {
	if l.level < logging.LogLevelInfo {
		return
	}
	l.ctp.LogInfo(format, args...)
}

// Warn logs at warn level
func (l *PionLoggerAdapter) Warn(msg string) {
	if l.level < logging.LogLevelWarn {
		return
	}
	l.ctp.LogWarn(msg)
}

// Warnf logs at warn level with formatting
func (l *PionLoggerAdapter) Warnf(format string, args ...interface{}) {
	if l.level < logging.LogLevelWarn {
		return
	}
	l.ctp.LogWarn(format, args...)
}

// Error logs at error level
func (l *PionLoggerAdapter) Error(msg string) {
	if l.level < logging.LogLevelError {
		return
	}
	l.ctp.LogError(msg)
}

// Errorf logs at error level with formatting
func (l *PionLoggerAdapter) Errorf(format string, args ...interface{}) {
	if l.level < logging.LogLevelError {
		return
	}
	l.ctp.LogError(format, args...)
}
