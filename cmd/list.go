package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/hurricanehrndz/respec/internal/artifact"
	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

type listRow struct {
	Repo       string   `json:"repo"`
	Slug       string   `json:"slug"`
	Dir        string   `json:"dir"`
	Date       string   `json:"date,omitempty"`
	PlanStatus string   `json:"plan_status"`
	Spec       string   `json:"spec"` // fresh | stale | unstamped | absent | error
	DependsOn  []string `json:"depends_on,omitempty"`
	Blocked    bool     `json:"blocked"`
}

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List every effort in the store, grouped by repo, with status and staleness",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			changes, err := store.Changes(cfg.StorePath())
			if err != nil {
				return err
			}

			rows := make([]listRow, 0, len(changes))
			doneByRef := map[string]bool{} // "repo/slug" -> plan is done
			var warnings []string
			for _, ch := range changes {
				row, done, warns := effortRow(ch)
				rows = append(rows, row)
				warnings = append(warnings, warns...)
				doneByRef[ch.Repo+"/"+ch.Slug] = done
			}

			var unresolved []string
			for i := range rows {
				rows[i].Blocked, unresolved = resolveDeps(rows[i], doneByRef, unresolved)
			}

			sort.Slice(rows, func(i, j int) bool {
				if rows[i].Repo != rows[j].Repo {
					return rows[i].Repo < rows[j].Repo
				}
				if rows[i].Date != rows[j].Date {
					return rows[i].Date > rows[j].Date // most recent first
				}
				return rows[i].Slug < rows[j].Slug
			})

			for _, w := range warnings {
				_, _ = fmt.Fprintf(c.ErrOrStderr(), "warning: %s\n", w)
			}
			for _, u := range unresolved {
				_, _ = fmt.Fprintf(c.ErrOrStderr(), "warning: unresolved depends_on reference %q\n", u)
			}

			if asJSON {
				return json.NewEncoder(c.OutOrStdout()).Encode(rows)
			}
			if len(rows) == 0 {
				_, err = fmt.Fprintf(c.OutOrStdout(), "no changes found in %s\n", cfg.StorePath())
				return err
			}
			_, err = fmt.Fprint(c.OutOrStdout(), renderTable(rows))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(cmd)
}

// effortRow reads a change's plan/spec/date into a display row (blocked unset)
// plus whether its plan is done, returning warnings for anything unreadable or
// unparseable rather than silently misreporting it (fail loud).
func effortRow(ch store.Change) (listRow, bool, []string) {
	row := listRow{Repo: ch.Repo, Slug: ch.Slug, Dir: ch.Dir}
	var warns []string

	date, derr := effortDate(ch.Dir)
	if derr != nil {
		warns = append(warns, fmt.Sprintf("%s: %v", ch.Dir, derr))
	}
	row.Date = date

	planData, err := os.ReadFile(store.PlanPath(ch.Dir))
	if err != nil {
		row.Spec = "error"
		return row, false, append(warns, fmt.Sprintf("%s: %v", store.PlanPath(ch.Dir), err))
	}
	art, perr := artifact.Parse(artifact.KindPlan, planData)
	if perr != nil {
		row.Spec = "error"
		return row, false, append(warns, fmt.Sprintf("%s: %v", store.PlanPath(ch.Dir), perr))
	}
	var recorded string
	var done bool
	if pl, ok := art.(artifact.Plan); ok {
		row.PlanStatus = pl.Status
		row.DependsOn = pl.DependsOn
		recorded = pl.SpecSHA
		done = pl.IsDone()
	}
	row.Spec = specStateFor(ch.Dir, recorded)
	return row, done, warns
}

// resolveDeps computes whether a row is blocked (any resolved dependency is not
// done) and accumulates unresolved references. Same-repo deps are bare slugs;
// cross-repo deps are "repo/slug".
func resolveDeps(row listRow, doneByRef map[string]bool, unresolved []string) (bool, []string) {
	blocked := false
	for _, d := range row.DependsOn {
		ref := d
		if !strings.Contains(d, "/") {
			ref = row.Repo + "/" + d
		}
		done, known := doneByRef[ref]
		if !known {
			unresolved = append(unresolved, d)
			continue
		}
		if !done {
			blocked = true
		}
	}
	return blocked, unresolved
}

// specStateFor reports spec.md staleness relative to the recorded hash.
func specStateFor(changeDir, recorded string) string {
	specPath := store.SpecPath(changeDir)
	if _, err := os.Stat(specPath); err != nil {
		return "absent"
	}
	sum, err := store.SHA256File(specPath)
	if err != nil {
		return "error"
	}
	return store.Freshness(recorded, sum)
}

// effortDate returns the date frontmatter of research.md (else spec.md), the
// origin date of the effort. Empty when neither carries one; a frontmatter
// parse error is returned so the caller can surface it.
func effortDate(changeDir string) (string, error) {
	for _, path := range []string{store.ResearchPath(changeDir), store.SpecPath(changeDir)} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // absent artifacts are fine at list level
		}
		d, _, gerr := frontmatter.GetString(data, "date")
		if gerr != nil {
			return "", fmt.Errorf("%s: %w", path, gerr)
		}
		if d != "" {
			return d, nil
		}
	}
	return "", nil
}

func renderTable(rows []listRow) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "REPO\tSLUG\tDATE\tPLAN\tSPEC\tDEPENDS ON")
	for _, r := range rows {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			dash(r.Repo), dash(r.Slug), shortDate(r.Date), dash(r.PlanStatus), dash(r.Spec), depsCell(r))
	}
	_ = w.Flush()
	return b.String()
}

func depsCell(r listRow) string {
	if len(r.DependsOn) == 0 {
		return "-"
	}
	s := strings.Join(r.DependsOn, ", ")
	if r.Blocked {
		s += " (blocked)"
	}
	return s
}

func shortDate(d string) string {
	if len(d) >= 10 {
		return d[:10]
	}
	return dash(d)
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
