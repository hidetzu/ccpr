package main

import (
	"fmt"
	"os"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ccpr <command> [flags]\n\nCommands:\n  review     Review a CodeCommit pull request\n  --version  Print version")
	}

	switch args[0] {
	case "review":
		return runReview(args[1:])
	case "--version", "version":
		fmt.Printf("ccpr version %s\nhttps://github.com/hidetzu/ccpr/releases/tag/%s\n", version, version)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
