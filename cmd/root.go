package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the respec release version, managed by `go tool versionbump`
// (see versionbump.yaml).
const version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:     "respec",
	Version: version,
	Short:   "Single-operator, spec-driven research → plan → implement workflow tool",
	Long: `respec drives a research → plan → implement workflow backed by a single
central store. The CLI does deterministic plumbing only: hold config, render and
install the /rsx:* pi prompt-templates, stamp/compare staleness, reflow prose, and
render/serve the store as a Hugo site.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
