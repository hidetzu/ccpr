package app

import (
	"errors"
	"fmt"
)

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

// Sentinel validation errors. CLI callers translate these to flag-aware
// wording via errors.Is; the MCP server lets them surface as-is so MCP clients
// see surface-neutral messages.
var (
	// ErrMissingPRRef is returned when a use case requires either a PR URL
	// or a (repo + prId) pair and neither is provided.
	ErrMissingPRRef = errors.New("provide a PR URL or repo and prId")

	// ErrMissingRegion is returned when no AWS region can be resolved from
	// the use-case input, the config file, or the standard AWS env variables.
	ErrMissingRegion = errors.New("region is required: use region option, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
)
