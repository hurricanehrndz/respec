package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/spf13/cobra"
)

// hookMarker identifies a respec-managed hook so re-installs are safe but
// hand-written hooks are never silently overwritten.
const hookMarker = "# respec-managed pre-commit hook"

const preCommitHook = hookMarker + `
# Verifies staged Markdown is reflowed before commit. Regenerate with: respec install-hook
files=$(git diff --cached --name-only --diff-filter=ACM -- '*.md')
[ -z "$files" ] && exit 0
fail=0
for f in $files; do
  respec format --check "$f" >/dev/null 2>&1 || { echo "not formatted: $f" >&2; fail=1; }
done
if [ "$fail" -ne 0 ]; then
  echo "run: respec format <file> (or 'respec format .') then re-stage" >&2
  exit 1
fi
exit 0
`

func init() {
	var force bool
	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install a pre-commit hook in the store repo that checks Markdown formatting",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			store := cfg.StorePath()
			gitDir := filepath.Join(store, ".git")
			info, err := os.Stat(gitDir)
			if err != nil || !info.IsDir() {
				return fmt.Errorf("store %q is not a git repository (run: git -C %q init)", store, store)
			}

			hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
			if existing, err := os.ReadFile(hookPath); err == nil {
				if !force && !strings.Contains(string(existing), hookMarker) {
					c.SilenceUsage = true
					return fmt.Errorf("%s already exists and is not respec-managed; re-run with --force to overwrite", hookPath)
				}
			}

			if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(hookPath, []byte("#!/bin/sh\n"+preCommitHook), 0o755); err != nil {
				return err
			}
			_, err = fmt.Fprintf(c.OutOrStdout(), "installed pre-commit hook: %s\n", hookPath)
			return err
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing non-respec pre-commit hook")
	rootCmd.AddCommand(cmd)
}
