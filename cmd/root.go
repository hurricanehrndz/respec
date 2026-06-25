package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "respec",
	Short: "Single-operator, spec-driven research → plan → implement workflow tool",
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

// notImplemented is a stub RunE for subcommands not yet built.
func notImplemented(c *cobra.Command, _ []string) error {
	fmt.Printf("%s: not implemented\n", c.Name())
	return nil
}
