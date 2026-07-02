package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			changeDir := args[0]
			if err := store.ValidateChangeDir(changeDir); err != nil {
				return err
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

			now := time.Now().Format(time.RFC3339)

			// Provenance is read from the worked-on repo. A non-git dir is not
			// fatal: we still stamp the date and warn about the skipped fields.
			researchKVs := []frontmatter.KV{{Key: "date", Value: now}}
			if gm, gerr := gitmeta.Read(repoDir); gerr != nil {
				_, _ = fmt.Fprintf(c.ErrOrStderr(), "warning: skipping repo/git_commit provenance: %v\n", gerr)
			} else {
				repo := gm.Remote
				if repo == "" {
					repo = filepath.Base(gm.Toplevel)
				}
				researchKVs = append(researchKVs,
					frontmatter.KV{Key: "repo", Value: repo},
					frontmatter.KV{Key: "repo_path", Value: gm.Toplevel},
					frontmatter.KV{Key: "git_commit", Value: gm.Commit},
				)
			}

			// research.md: provenance, write-once (never clobber the snapshot).
			if _, err := stampFile(store.ResearchPath(changeDir), researchKVs, cfg.ReflowWidth, true); err != nil {
				return err
			}
			// spec.md: date, write-once. (Its hash is recorded into the plan below.)
			if _, err := stampFile(store.SpecPath(changeDir), []frontmatter.KV{{Key: "date", Value: now}}, cfg.ReflowWidth, true); err != nil {
				return err
			}

			// plan.md: refresh spec_sha256 from the now-finalized spec.md.
			sum, err := store.SHA256File(store.SpecPath(changeDir))
			if err != nil {
				return fmt.Errorf("hashing spec.md: %w", err)
			}
			if _, err := stampFile(store.PlanPath(changeDir), []frontmatter.KV{{Key: specSHAKey, Value: sum}}, cfg.ReflowWidth, false); err != nil {
				return err
			}

			_, err = fmt.Fprintf(c.OutOrStdout(), "stamped %s: %s=%s\n", store.PlanPath(changeDir), specSHAKey, sum)
			return err
		},
	}
	cmd.Flags().StringVar(&repoDir, "repo", "", "worked-on repo to read provenance from (default: current directory)")
	rootCmd.AddCommand(cmd)
}

// stampFile sets the given frontmatter keys on the file at path, then reflows it
// and writes it back only if changed. When writeOnce is true, keys already
// present are left untouched. A missing file is skipped (returns false, nil).
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
			_, present, gerr := frontmatter.GetString(data, kv.Key)
			if gerr != nil {
				return false, gerr
			}
			if !present {
				toSet = append(toSet, kv)
			}
		}
	}

	updated := data
	if len(toSet) > 0 {
		if updated, err = frontmatter.SetMany(data, toSet); err != nil {
			return false, err
		}
	}
	updated = formatter.Format(updated, width)

	if bytes.Equal(data, updated) {
		return false, nil
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		return false, err
	}
	return true, nil
}
