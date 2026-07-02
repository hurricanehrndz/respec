package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hurricanehrndz/respec/internal/artifact"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

type artifactStatus struct {
	Artifact string `json:"artifact"`
	Present  bool   `json:"present"`
	Status   string `json:"status,omitempty"`
	Error    string `json:"error,omitempty"`
}

type statusResult struct {
	State     string           `json:"state"`
	Change    string           `json:"change"`
	SpecSHA   string           `json:"spec_sha256"`
	Recorded  string           `json:"recorded"`
	Artifacts []artifactStatus `json:"artifacts"`
}

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status <change-dir>",
		Short: "Report plan staleness (fresh / stale / unstamped) and per-artifact status",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			changeDir := args[0]
			if err := store.ValidateChangeDir(changeDir); err != nil {
				return err
			}
			specSHA, err := store.SHA256File(store.SpecPath(changeDir))
			if err != nil {
				return fmt.Errorf("hashing spec.md: %w", err)
			}
			planData, err := os.ReadFile(store.PlanPath(changeDir))
			if err != nil {
				return err
			}
			art, err := artifact.Parse(artifact.KindPlan, planData)
			if err != nil {
				return err
			}
			var recorded string
			if pl, ok := art.(artifact.Plan); ok {
				recorded = pl.SpecSHA
			}

			res := statusResult{Change: changeDir, SpecSHA: specSHA, Recorded: recorded}
			res.State = store.Freshness(recorded, specSHA)
			res.Artifacts = artifactStatuses(changeDir)

			if asJSON {
				return json.NewEncoder(c.OutOrStdout()).Encode(res)
			}
			_, err = fmt.Fprint(c.OutOrStdout(), plainStatus(res))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(cmd)
}

// artifactStatuses reports the presence and frontmatter status of each artifact.
func artifactStatuses(changeDir string) []artifactStatus {
	targets := artifactTargets(changeDir)
	out := make([]artifactStatus, 0, len(targets))
	for _, t := range targets {
		as := artifactStatus{Artifact: t.name}
		switch data, err := os.ReadFile(t.path); {
		case err == nil:
			as.Present = true
			if art, perr := artifact.Parse(t.kind, data); perr != nil {
				as.Error = perr.Error()
			} else {
				as.Status = artifact.StatusOf(art)
			}
		case !errors.Is(err, os.ErrNotExist):
			// Unreadable-for-another-reason: report it, never drop the entry.
			as.Error = err.Error()
		}
		out = append(out, as)
	}
	return out
}

// plainStatus renders the human-readable status: the state word (first line, for
// scripting), then one line per artifact.
func plainStatus(res statusResult) string {
	out := res.State + "\n"
	for _, a := range res.Artifacts {
		switch {
		case a.Error != "":
			out += fmt.Sprintf("  %-12s error: %s\n", a.Artifact, a.Error)
		case !a.Present:
			out += fmt.Sprintf("  %-12s absent\n", a.Artifact)
		default:
			st := a.Status
			if st == "" {
				st = "-"
			}
			out += fmt.Sprintf("  %-12s present (%s)\n", a.Artifact, st)
		}
	}
	return out
}
