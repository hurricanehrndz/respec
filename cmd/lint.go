package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hurricanehrndz/respec/internal/artifact"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

type lintArtifact struct {
	Artifact string   `json:"artifact"`
	Problems []string `json:"problems"`
}

type lintResult struct {
	Change    string         `json:"change"`
	OK        bool           `json:"ok"`
	Artifacts []lintArtifact `json:"artifacts"`
}

// lintTarget pairs an artifact kind with its filename and path in a change dir.
type lintTarget struct {
	name string
	kind artifact.Kind
	path string
}

func lintTargets(changeDir string) []lintTarget {
	return []lintTarget{
		{"research.md", artifact.KindResearch, store.ResearchPath(changeDir)},
		{"spec.md", artifact.KindSpec, store.SpecPath(changeDir)},
		{"plan.md", artifact.KindPlan, store.PlanPath(changeDir)},
	}
}

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "lint <change-dir>",
		Short: "Validate frontmatter fields, status, and required sections of the change's artifacts",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			changeDir := args[0]
			info, err := os.Stat(changeDir)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("change dir %q is not a directory", changeDir)
			}

			res := lintResult{Change: changeDir, OK: true}
			present := 0
			for _, t := range lintTargets(changeDir) {
				data, err := os.ReadFile(t.path)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return err
				}
				present++
				var problems []string
				art, perr := artifact.Parse(t.kind, data)
				if perr != nil {
					problems = []string{perr.Error()}
				} else {
					problems = art.Validate()
				}
				if len(problems) > 0 {
					res.OK = false
				}
				res.Artifacts = append(res.Artifacts, lintArtifact{Artifact: t.name, Problems: problems})
			}
			if present == 0 {
				return fmt.Errorf("no research.md, spec.md, or plan.md found in %q", changeDir)
			}

			if asJSON {
				enc := json.NewEncoder(c.OutOrStdout())
				if err := enc.Encode(res); err != nil {
					return err
				}
			} else {
				var b strings.Builder
				for _, a := range res.Artifacts {
					if len(a.Problems) == 0 {
						fmt.Fprintf(&b, "%s: ok\n", a.Artifact)
						continue
					}
					fmt.Fprintf(&b, "%s:\n", a.Artifact)
					for _, p := range a.Problems {
						fmt.Fprintf(&b, "  - %s\n", p)
					}
				}
				if _, err := fmt.Fprint(c.OutOrStdout(), b.String()); err != nil {
					return err
				}
			}

			if !res.OK {
				c.SilenceUsage = true
				return errors.New("lint failed")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(cmd)
}
