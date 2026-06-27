package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/chernand/respec/internal/config"
	"github.com/chernand/respec/internal/formatter"
	"github.com/spf13/cobra"
)

func init() {
	var check bool
	cmd := &cobra.Command{
		Use:   "format <file>",
		Short: "Reflow prose to the configured width, leaving non-prose byte-identical",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			path := args[0]
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out := formatter.Format(src, cfg.ReflowWidth)

			if check {
				if !bytes.Equal(src, out) {
					return fmt.Errorf("%s is not formatted", path)
				}
				fmt.Fprintln(c.OutOrStdout(), "formatted")
				return nil
			}
			if bytes.Equal(src, out) {
				return nil
			}
			return os.WriteFile(path, out, 0o644)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report whether the file is already formatted without writing")
	rootCmd.AddCommand(cmd)
}
