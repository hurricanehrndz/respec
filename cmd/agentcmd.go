package cmd

import (
	"fmt"

	"github.com/hurricanehrndz/respec/internal/agents"
	"github.com/spf13/cobra"
)

func init() {
	var promptFile string
	cmd := &cobra.Command{
		Use:   "agent-cmd <spec>",
		Short: "Print the exact command line for one delegation spec",
		Long: "Build the invocation for an agent spec of the form <harness>:<model>@<effort>,\n" +
			"e.g. pi:openai-codex/gpt-5.6-sol@medium or claude:opus@high.\n\n" +
			"Prompts are fed on stdin, never as an argument: phase tasks contain quotes,\n" +
			"backticks, and newlines, and hand-assembling them into a command line is where\n" +
			"delegation loops break. Pass --prompt-file to get a ready-to-run redirection.",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			spec, err := agents.ParseSpec(args[0])
			if err != nil {
				return err
			}
			if promptFile == "" {
				promptFile = "/dev/stdin"
			}
			line, err := spec.Shell(promptFile)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(c.OutOrStdout(), line)
			return nil
		},
	}
	cmd.Flags().StringVar(&promptFile, "prompt-file", "", "file the delegated prompt is read from (default /dev/stdin)")
	rootCmd.AddCommand(cmd)
}
