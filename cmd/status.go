package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

type artifactStatus struct {
	Artifact string `json:"artifact"`
	Present  bool   `json:"present"`
	Status   string `json:"status,omitempty"`
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
			plan, err := os.ReadFile(store.PlanPath(changeDir))
			if err != nil {
				return err
			}
			recorded, present, err := frontmatter.GetString(plan, specSHAKey)
			if err != nil {
				return err
			}

			res := statusResult{Change: changeDir, SpecSHA: specSHA, Recorded: recorded}
			switch {
			case !present || recorded == "":
				res.State = "unstamped"
			case recorded == specSHA:
				res.State = "fresh"
			default:
				res.State = "stale"
			}
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
	targets := []struct {
		name string
		path string
	}{
		{"research.md", store.ResearchPath(changeDir)},
		{"spec.md", store.SpecPath(changeDir)},
		{"plan.md", store.PlanPath(changeDir)},
	}
	out := make([]artifactStatus, 0, len(targets))
	for _, t := range targets {
		as := artifactStatus{Artifact: t.name}
		if data, err := os.ReadFile(t.path); err == nil {
			as.Present = true
			if s, ok, gerr := frontmatter.GetString(data, "status"); gerr == nil && ok {
				as.Status = s
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			// Unreadable-for-another-reason: leave Present=false; status/lint surface details.
			continue
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
		if !a.Present {
			out += fmt.Sprintf("  %-12s absent\n", a.Artifact)
			continue
		}
		st := a.Status
		if st == "" {
			st = "-"
		}
		out += fmt.Sprintf("  %-12s present (%s)\n", a.Artifact, st)
	}
	return out
}
