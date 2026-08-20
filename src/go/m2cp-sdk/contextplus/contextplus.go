package contextplus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"m2cp"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

var (
	logger                *zerolog.Logger
	output                = os.Stdout
	counterBranch         = 0
	counterBranchLock     = &sync.Mutex{}
	debugMode             = false
	customWriterSingleton *CustomWriter
)

type Context struct {
	context.Context
	cancel                  context.CancelFunc
	cancelInterruptListener context.CancelFunc
	logger                  zerolog.Logger
	name                    string
	parent                  *Context
	// callerSkip is the number of stack frames to skip when reporting the file and line number of the log message.
	// Default is 1, which means the log message will report the file and line number of the caller of the logging function.
	// If the logging function is wrapped in another function it might be configured differently, that is to be left for the future.
	callerSkip int
}

func NewContextPlus() m2cp.ContextPlus {
	return FromContext(context.Background())
}

func FromContext(from context.Context) m2cp.ContextPlus {
	ctx, cancelInt := signal.NotifyContext(from, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(ctx)

	// cancel the interrupt listener when the child context is cancelled
	go func() {
		countBranch(1)
		<-ctx.Done()
		cancelInt()
		countBranch(-1)
	}()

	skip := zerolog.CallerSkipFrameCount + 1
	c := Context{
		Context:                 ctx,
		cancel:                  cancel,
		cancelInterruptListener: cancelInt,
		logger:                  LoggerConsole().With().CallerWithSkipFrameCount(skip).Logger(),
		callerSkip:              skip,
	}
	return &c
}

var (
	logToCloudLevel m2cp.LogLevel
	logToCloudNode  m2cp.Node
	logToCloudLock  sync.Mutex
)

func ConfigureLogToCloud(node m2cp.Node, level m2cp.LogLevel) {
	logToCloudLock.Lock()
	defer logToCloudLock.Unlock()
	logToCloudLevel = level
	logToCloudNode = node
}

func logToCloud(module, msg string, level m2cp.LogLevel) {
	// catch panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("logToCloud panic: %v\n", r)
		}
	}()

	// sending messages is logged, logs get send as messages... so we need to avoid a loop here
	if module == "networks.amqp.Sender" {
		return
	}

	node := logToCloudNode
	if node == nil {
		return
	}
	if level < logToCloudLevel {
		return
	}
	var signalLevel string
	switch level {
	case m2cp.LogLevelDebug:
		signalLevel = m2cp.SignalLevelDebug
	case m2cp.LogLevelInfo:
		signalLevel = m2cp.SignalLevelInfo
	case m2cp.LogLevelWarn:
		signalLevel = m2cp.SignalLevelWarning
	case m2cp.LogLevelError:
		signalLevel = m2cp.SignalLevelError
	case m2cp.LogLevelFatal:
		signalLevel = m2cp.SignalLevelError
	default:
		signalLevel = m2cp.SignalLevelInfo
	}
	_ = node.EmitSignal(module, msg, signalLevel)
}

func (c *Context) Cancel() {
	c.cancelInterruptListener()
	c.cancel()
}

func (c *Context) IsCancelled() bool {
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

func (c *Context) SetModule(name string) m2cp.ContextPlus {
	c.logger = c.logger.With().Str("module", name).Logger()
	c.name = name
	return c
}

func (c *Context) SetTaskId(id string) m2cp.ContextPlus {
	c.logger = c.logger.With().Str("task", id).Logger()
	return c
}

func (c *Context) SetDeviceId(id string) m2cp.ContextPlus {
	c.logger = c.logger.With().Str("device", id).Logger()
	return c
}

func (c *Context) SetUserId(id string) m2cp.ContextPlus {
	c.logger = c.logger.With().Str("user", id).Logger()
	return c
}

func (c *Context) Branch() m2cp.ContextPlus {
	ctx, cancel := context.WithCancel(c.Context)
	ctx, cancelInt := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// cancel the interrupt listener when the child context is cancelled
	go func() {
		countBranch(1)
		<-ctx.Done()
		cancelInt()
		countBranch(-1)
	}()

	return &Context{
		Context:                 ctx,
		cancel:                  cancel,
		cancelInterruptListener: cancelInt,
		logger:                  c.logger,
		name:                    c.name,
		parent:                  c,
		callerSkip:              c.callerSkip,
	}
}

func (c *Context) BranchWithName(name string) m2cp.ContextPlus {
	ctp := c.Branch()
	ctp = ctp.SetModule(name)
	return ctp
}

func (c *Context) BranchWithTimeout(duration time.Duration) m2cp.ContextPlus {
	ctx, cancel := context.WithTimeout(c.Context, duration)
	ctx, cancelInt := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// cancel the interrupt listener when the child context is cancelled
	go func() {
		countBranch(1)
		<-ctx.Done()
		cancelInt()
		countBranch(-1)
	}()

	return &Context{
		Context:                 ctx,
		cancel:                  cancel,
		cancelInterruptListener: cancelInt,
		logger:                  c.logger,
		name:                    c.name,
		parent:                  c,
		callerSkip:              c.callerSkip,
	}
}

func (c *Context) Parent() m2cp.ContextPlus {
	return c.parent
}

func (c *Context) LogDebug(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToCloud(c.name, msg, m2cp.LogLevelDebug)
	c.logger.Debug().Msg(msg)
}

func (c *Context) LogInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToCloud(c.name, msg, m2cp.LogLevelInfo)
	c.logger.Info().Msg(msg)
}

func (c *Context) LogWarn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToCloud(c.name, msg, m2cp.LogLevelWarn)
	c.logger.Warn().Msg(msg)
}

func (c *Context) LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToCloud(c.name, msg, m2cp.LogLevelError)
	c.logger.Error().Msg(msg)
}

func (c *Context) LogFatal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToCloud(c.name, msg, m2cp.LogLevelFatal)
	c.LogWarn("fatal error - waiting 5 seconds before exiting")
	c.Sleep(5 * time.Second)
	c.logger.Fatal().Msg(msg)
}

func (c *Context) SetLogLevel(level m2cp.LogLevel) {
	c.logger = c.logger.Level(zerolog.Level(level))
}

func (c *Context) LoosenLogLevel(level m2cp.LogLevel) {
	if c.GetLogLevel() > level {
		c.SetLogLevel(level)
	}
}

func (c *Context) GetLogLevel() m2cp.LogLevel {
	return m2cp.LogLevel(c.logger.GetLevel())
}

func (c *Context) SetLogOutput(w io.Writer) {
	c.logger = c.logger.Output(w)
}

func (c *Context) Sleep(duration time.Duration) {
	select {
	case <-c.Done():
		return
	case <-time.After(duration):
	}
}

func isDeveloperMachine() bool {
	if runtime.GOARCH == "amd64" {
		return true
	}
	if os.Getenv("HOSTTYPE") == "x86_64" {
		return true
	}
	return false
}

func LoggerConsole() zerolog.Logger {
	if logger != nil {
		return *logger
	}
	// Output to console
	var l zerolog.Logger
	if isDeveloperMachine() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		l = zerolog.New(getCustomWriter(os.Stdout)).With().Timestamp().Logger()
	} else {
		l = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	logger = &l
	return *logger
}

// CustomWriter is an io.Writer that processes JSON log entries.
type CustomWriter struct {
	output io.Writer
}

func getCustomWriter(output io.Writer) *CustomWriter {
	if customWriterSingleton == nil {
		customWriterSingleton = &CustomWriter{output: output}
	}
	return customWriterSingleton
}

// Write processes the JSON log entry.
func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	// Unmarshal the JSON log entry
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return 0, err
	}

	// Format the log entry (custom formatting logic here)
	formattedLog := cw.formatLogEntry(logEntry)

	// Write the formatted log to the output
	return cw.output.Write([]byte(formattedLog))
}

// formatLogEntry formats the log entry as desired
func (cw *CustomWriter) formatLogEntry(logEntry map[string]interface{}) string {
	var buf bytes.Buffer
	if timestamp, ok := logEntry["time"]; ok {
		if tsFloat, ok := timestamp.(float64); ok {
			t := time.Unix(0, int64(tsFloat*float64(time.Second))).UTC()
			buf.WriteString(t.Format("15:04:05.000|"))
		} else {
			buf.WriteString(fmt.Sprintf("%v|", timestamp))
		}
	}
	if level, ok := logEntry["level"]; ok {
		levelStr := strings.ToUpper(fmt.Sprintf("%v", level))
		buf.WriteString(fmt.Sprintf("%v|", levelStr))
	}
	if module, ok := logEntry["module"]; ok {
		buf.WriteString(fmt.Sprintf("%v|", module))
	}

	if message, ok := logEntry["message"]; ok {
		buf.WriteString(fmt.Sprintf(" %s", message))
	}

	if caller, ok := logEntry["caller"]; ok {
		buf.WriteString(fmt.Sprintf(" (%v)", caller))
	}

	buf.WriteString("\n")
	return buf.String()
}

func SetDebugMode(debug bool) {
	debugMode = debug
}

func countBranch(delta int) {
	counterBranchLock.Lock()
	counterBranch += delta
	if debugMode {
		fmt.Printf("countBranch %d -> %d\n", delta, counterBranch)
	}
	counterBranchLock.Unlock()
}
