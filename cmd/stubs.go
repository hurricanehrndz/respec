package cmd

import "github.com/spf13/cobra"

// Phase 1 stubs. Each subcommand is fleshed out in a later phase; for now they
// print "not implemented" so the CLI surface (and --help) is complete.

func init() {
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Render and install the /rsx:* prompt-templates and skill at user scope",
			RunE:  notImplemented,
		},
		&cobra.Command{
			Use:   "stamp",
			Short: "Record spec.md's sha256 in plan.md frontmatter",
			RunE:  notImplemented,
		},
		&cobra.Command{
			Use:   "status",
			Short: "Report plan staleness (fresh / stale / unstamped)",
			RunE:  notImplemented,
		},
		&cobra.Command{
			Use:   "format",
			Short: "Reflow prose to the configured width, leaving non-prose byte-identical",
			RunE:  notImplemented,
		},
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
