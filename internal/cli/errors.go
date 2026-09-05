// Package cli defines the cobra commands for the nscardprice CLI.
package cli

import "errors"

// ExitError carries an explicit process exit code and an optional human
// message shown on stderr.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

func exitError(code int, msg string) error {
	return &ExitError{Code: code, Msg: msg}
}

// AsExitError unwraps an ExitError if present.
func AsExitError(err error) (*ExitError, bool) {
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee, true
	}
	return nil, false
}

// ExitCode maps an error to a process exit code: 0 for nil, the carried code
// for an ExitError, and 1 for any other error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := AsExitError(err); ok {
		return ee.Code
	}
	return 1
}
