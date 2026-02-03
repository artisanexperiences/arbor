package main

import (
	"os"

	"github.com/michaeldyrynda/arbor/internal/cli"
)

// These variables are set at build time via -ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cli.Version = version
	cli.Commit = commit
	cli.BuildDate = date
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
