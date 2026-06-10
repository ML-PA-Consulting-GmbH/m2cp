package tools

import "fmt"

func ErrorRewrite(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(msg, args...)
}
