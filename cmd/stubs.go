package cmd

import "github.com/spf13/cobra"

// Phase 1 stubs. Each subcommand is fleshed out in a later phase; for now they
// print "not implemented" so the CLI surface (and --help) is complete.

func init() {
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "render",
			Short: "Build the central store as a browsable Hugo site",
			RunE:  notImplemented,
		},
		&cobra.Command{
			Use:   "serve",
			Short: "Serve the central store as a Hugo site with live reload",
			RunE:  notImplemented,
		},
	)
}
