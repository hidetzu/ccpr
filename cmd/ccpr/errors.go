package main

import (
	"errors"
	"fmt"

	"github.com/hidetzu/ccpr/internal/app"
)

// translateAppError converts surface-neutral validation errors from
// internal/app into CLI flag-aware messages. AWS/system failures and other
// errors pass through unchanged.
func translateAppError(err error) error {
	switch {
	case errors.Is(err, app.ErrMissingPRRef):
		return fmt.Errorf("provide a PR URL or --repo and --pr-id flags")
	case errors.Is(err, app.ErrMissingRegion):
		return fmt.Errorf("region is required: use --region flag, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
	}
	return err
}
