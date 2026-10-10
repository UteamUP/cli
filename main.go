package main

import (
	"fmt"
	"os"

	"github.com/uteamup/cli/cmd"
	"github.com/uteamup/cli/internal/logging"
)

// Build-time variables set by goreleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetBuildInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", logging.SafeDiagnostic(err.Error()))
		os.Exit(1)
	}
}
