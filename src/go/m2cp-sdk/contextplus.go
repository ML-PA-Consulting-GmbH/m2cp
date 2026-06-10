package m2cp

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

type LogLevel zerolog.Level

const (
	LogLevelDebug    = LogLevel(zerolog.DebugLevel)
	LogLevelInfo     = LogLevel(zerolog.InfoLevel)
	LogLevelWarn     = LogLevel(zerolog.WarnLevel)
	LogLevelError    = LogLevel(zerolog.ErrorLevel)
	LogLevelFatal    = LogLevel(zerolog.FatalLevel)
	LogLevelPanic    = LogLevel(zerolog.PanicLevel)
	LogLevelDisabled = LogLevel(zerolog.Disabled)
)

type ContextPlus interface {
	Cancel()
	IsCancelled() bool
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
	SetModule(name string) ContextPlus
	SetTaskId(id string) ContextPlus
	SetDeviceId(id string) ContextPlus
	SetUserId(id string) ContextPlus
	Branch() ContextPlus
	BranchWithName(name string) ContextPlus
	BranchWithTimeout(duration time.Duration) ContextPlus
	LogDebug(format string, a ...interface{})
	LogInfo(format string, a ...interface{})
	LogWarn(format string, a ...interface{})
	LogError(format string, a ...interface{})
	LogFatal(format string, a ...interface{})
	SetLogLevel(level LogLevel)
	LoosenLogLevel(level LogLevel)
	GetLogLevel() LogLevel
	SetLogOutput(w io.Writer)
	Sleep(duration time.Duration)
	Parent() ContextPlus
}
