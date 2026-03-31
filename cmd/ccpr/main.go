package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
)

// systemError represents a system-level error (AWS, Git) that should exit with code 2.
type systemError struct {
	err error
}

func (e *systemError) Error() string { return e.err.Error() }
func (e *systemError) Unwrap() error { return e.err }

func newSystemError(format string, a ...any) error {
	return &systemError{err: fmt.Errorf(format, a...)}
}

// exitCode returns the appropriate exit code for the given error.
// System errors (AWS, Git) return 2; user errors return 1.
func exitCode(err error) int {
	var sysErr *systemError
	if errors.As(err, &sysErr) {
		return 2
	}
	return 1
}

// version is set via ldflags at build time.
// Falls back to module version from BuildInfo (go install).
var version = ""

func getVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(exitCode(err))
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ccpr <command> [flags]\n\nCommands:\n  review     Review a CodeCommit pull request\n  list       List pull requests for a repository\n  open       Open a pull request in the browser\n  init       Initialize configuration file\n  doctor     Validate environment and config\n  comment    Post a comment to a pull request\n  --version  Print version")
	}

	switch args[0] {
	case "review":
		return runReview(args[1:])
	case "list":
		return runList(args[1:])
	case "open":
		return runOpen(args[1:])
	case "init":
		return runInit(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "comment":
		return runComment(args[1:])
	case "--version", "version":
		v := getVersion()
		fmt.Printf("ccpr version %s\nhttps://github.com/hidetzu/ccpr/releases/tag/%s\n", v, v)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
