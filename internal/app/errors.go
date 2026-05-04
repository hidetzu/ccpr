package app

import "fmt"

// SystemError marks an error as system-level (AWS, Git) so CLI callers can
// route it to a non-user exit code. Detect via errors.As.
type SystemError struct {
	err error
}

func (e *SystemError) Error() string { return e.err.Error() }
func (e *SystemError) Unwrap() error { return e.err }

func newSystemError(format string, a ...any) error {
	return &SystemError{err: fmt.Errorf(format, a...)}
}
