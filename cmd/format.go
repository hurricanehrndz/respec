package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/formatter"
	"github.com/spf13/cobra"
)

func init() {
	var check bool
	cmd := &cobra.Command{
		Use:   "format <path>...",
		Short: "Reflow prose to the configured width, leaving non-prose byte-identical",
		Long: `Reflow prose in Markdown files, or in every *.md file under a directory
(recursively), leaving tables, fenced code, headings, inline HTML, and bare
URLs byte-identical. Multiple paths are accepted (as passed by pre-commit).`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			var files []string
			for _, arg := range args {
				targets, err := markdownTargets(arg)
				if err != nil {
					return err
				}
				files = append(files, targets...)
			}

			var unformatted []string
			for _, path := range files {
				src, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				out := formatter.Format(src, cfg.ReflowWidth)
				if bytes.Equal(src, out) {
					continue
				}
				if check {
					unformatted = append(unformatted, path)
					continue
				}
				if err := os.WriteFile(path, out, 0o644); err != nil {
					return err
				}
			}

			if check {
				if len(unformatted) > 0 {
					c.SilenceUsage = true
					return fmt.Errorf("%d file(s) not formatted:\n  %s",
						len(unformatted), strings.Join(unformatted, "\n  "))
				}
				_, err = fmt.Fprintln(c.OutOrStdout(), "formatted")
				return err
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report whether files are already formatted without writing")
	rootCmd.AddCommand(cmd)
}

// markdownTargets returns the file(s) to format: path itself when it is a file,
// or every *.md file beneath it (recursively) when it is a directory.
func markdownTargets(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return nil
	})
	if err == nil && len(files) == 0 {
		// A dir with nothing to format would make --check pass vacuously.
		return nil, fmt.Errorf("no *.md files found under %q", path)
	}
	return files, err
}
