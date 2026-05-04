package main

import (
	"errors"
	"testing"

	"github.com/hidetzu/ccpr/internal/app"
)

func TestTranslateAppErrorMissingPRRef(t *testing.T) {
	got := translateAppError(app.ErrMissingPRRef)
	want := "provide a PR URL or --repo and --pr-id flags"
	if got == nil || got.Error() != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestTranslateAppErrorMissingRegion(t *testing.T) {
	got := translateAppError(app.ErrMissingRegion)
	want := "region is required: use --region flag, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env"
	if got == nil || got.Error() != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestTranslateAppErrorPassesThrough(t *testing.T) {
	other := errors.New("something else")
	got := translateAppError(other)
	if got != other {
		t.Fatalf("got %v, want passthrough", got)
	}
}

func TestTranslateAppErrorPassesThroughNil(t *testing.T) {
	if got := translateAppError(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}