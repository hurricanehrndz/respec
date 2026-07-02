package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/formatter"
	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"github.com/hurricanehrndz/respec/internal/gitmeta"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

// specSHAKey is the plan.md frontmatter key holding the stamped spec.md hash.
const specSHAKey = "spec_sha256"

func init() {
	var repoDir string
	cmd := &cobra.Command{
		Use:   "stamp <change-dir>",
		Short: "Write deterministic frontmatter (provenance + spec hash) and reflow the change's artifacts",
		Long: `Stamp whichever artifacts exist in the change directory: provenance
(date, repo, repo_path, git_commit) write-once into research.md, date
write-once into spec.md, and spec.md's sha256 into plan.md (refreshed).
Each stamped file is reflowed. Works at any workflow phase — a change dir
holding only research.md is fine.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			changeDir := args[0]
			info, err := os.Stat(changeDir)
			if err != nil {
				return fmt.Errorf("change dir %q: %w", changeDir, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("change dir %q is not a directory", changeDir)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if repoDir == "" {
				if repoDir, err = os.Getwd(); err != nil {
					return err
				}
			}

			// Provenance comes from the worked-on repo; respec's model is
			// git-based, so a non-git dir is an error, not a silent skip —
			// lint requires the fields only git can provide.
			gm, err := gitmeta.Read(repoDir)
			if err != nil {
				c.SilenceUsage = true
				return fmt.Errorf("reading provenance from %q: %w (the worked-on repo must be a git repository; use --repo to point at it)", repoDir, err)
			}
			repo := gm.Remote
			if repo == "" {
				repo = filepath.Base(gm.Toplevel)
			}
			now := time.Now().Format(time.RFC3339)

			var stamped []string

			// research.md: provenance, write-once (never clobber the snapshot).
			researchKVs := []frontmatter.KV{
				{Key: "date", Value: now},
				{Key: "repo", Value: repo},
				{Key: "repo_path", Value: gm.Toplevel},
				{Key: "git_commit", Value: gm.Commit},
			}
			resPresent, err := stampFile(store.ResearchPath(changeDir), researchKVs, cfg.ReflowWidth, true)
			if err != nil {
				return err
			}
			if resPresent {
				stamped = append(stamped, "research.md: provenance")
			}

			// spec.md: date, write-once. Stamped before hashing so the plan
			// records the finalized spec bytes.
			specPresent, err := stampFile(store.SpecPath(changeDir), []frontmatter.KV{{Key: "date", Value: now}}, cfg.ReflowWidth, true)
			if err != nil {
				return err
			}
			if specPresent {
				stamped = append(stamped, "spec.md: date")
			}

			// plan.md: refresh spec_sha256 (needs both plan and spec).
			planPresent := false
			if _, err := os.Stat(store.PlanPath(changeDir)); err == nil {
				planPresent = true
			}
			switch {
			case planPresent && specPresent:
				sum, err := store.SHA256File(store.SpecPath(changeDir))
				if err != nil {
					return fmt.Errorf("hashing spec.md: %w", err)
				}
				if _, err := stampFile(store.PlanPath(changeDir), []frontmatter.KV{{Key: specSHAKey, Value: sum}}, cfg.ReflowWidth, false); err != nil {
					return err
				}
				stamped = append(stamped, fmt.Sprintf("plan.md: %s=%s", specSHAKey, sum))
			case planPresent:
				stamped = append(stamped, "plan.md: skipped "+specSHAKey+" (no spec.md)")
			}

			if !resPresent && !specPresent && !planPresent {
				return fmt.Errorf("no research.md, spec.md, or plan.md found in %q", changeDir)
			}

			_, err = fmt.Fprintf(c.OutOrStdout(), "stamped %s\n  %s\n", changeDir, strings.Join(stamped, "\n  "))
			return err
		},
	}
	cmd.Flags().StringVar(&repoDir, "repo", "", "worked-on repo to read provenance from (default: current directory)")
	rootCmd.AddCommand(cmd)
}

// stampFile sets the given frontmatter keys on the file at path, then reflows it
// and writes it back only if changed. When writeOnce is true, keys that already
// hold a non-empty value are left untouched (an empty or null key is filled).
// It reports whether the file was present.
func stampFile(path string, kvs []frontmatter.KV, width int, writeOnce bool) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	toSet := kvs
	if writeOnce {
		toSet = nil
		for _, kv := range kvs {
			val, _, gerr := frontmatter.GetString(data, kv.Key)
			if gerr != nil {
				return true, gerr
			}
			if val == "" {
				toSet = append(toSet, kv)
			}
		}
	}

	updated := data
	if len(toSet) > 0 {
		if updated, err = frontmatter.SetMany(data, toSet); err != nil {
			return true, err
		}
	}
	updated = formatter.Format(updated, width)

	if bytes.Equal(data, updated) {
		return true, nil
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		return true, err
	}
	return true, nil
}
