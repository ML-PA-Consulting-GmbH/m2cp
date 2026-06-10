package f

import (
	"fmt"
	"os"
)

var verbose bool

// SetVerbose controls the global verbosity level.
// When false, Debug output is suppressed. Fatal and Warn always print.
func SetVerbose(v bool) {
	verbose = v
}

// IsVerbose returns the current verbosity setting.
func IsVerbose() bool {
	return verbose
}

// Fatal prints an error message to stderr and exits with code 1.
// Always prints regardless of verbosity.
func Fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}

// Warn prints a warning to stderr. Always prints regardless of verbosity.
func Warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARN: "+format+"\n", args...)
}

// Info prints an informational message to stderr. Always prints.
func Info(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Debug prints a diagnostic message to stderr. Only prints when --verbose is set.
func Debug(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
