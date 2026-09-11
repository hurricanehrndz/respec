package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/hurricanehrndz/respec/internal/agents"
	"github.com/spf13/cobra"
)

// listCap bounds how many models are printed per harness by default. A full
// catalogue runs to hundreds of rows; dumping that into the planning session
// it exists to inform would cost more context than the decision is worth.
const listCap = 8

func init() {
	var asJSON bool
	var filter string
	var all bool
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Probe external agent CLIs, models, and effort levels",
		Long: "Report which external agent CLIs are installed here, which models each can reach,\n" +
			"and which reasoning levels it accepts. Use this after choosing external CLI\n" +
			"delegation. Native subagents are not probed; their availability is managed by\n" +
			"the current app or agent environment. The CLI roster is probed, never stored.",
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			found := agents.Detect(c.Context())

			if filter != "" {
				for i := range found {
					found[i].Models = matching(found[i].Models, filter)
				}
			}

			if asJSON {
				enc := json.NewEncoder(c.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(found)
			}

			out := c.OutOrStdout()
			any := false
			for _, a := range found {
				if !a.Installed() {
					continue
				}
				any = true
				_, _ = fmt.Fprintf(out, "%s (%s)\n", a.Harness, a.Path)
				_, _ = fmt.Fprintf(out, "  efforts: %s\n", strings.Join(a.Efforts, ", "))
				switch {
				case len(a.Models) > 0:
					shown := a.Models
					if !all && len(shown) > listCap {
						shown = shown[:listCap]
					}
					_, _ = fmt.Fprintf(out, "  models:  %d available, showing %d\n", len(a.Models), len(shown))
					tw := tabwriter.NewWriter(out, 4, 4, 2, ' ', 0)
					for _, m := range shown {
						_, _ = fmt.Fprintf(tw, "    %s\t%s:%s@%s\n", m, a.Harness, m, defaultEffort(a))
					}
					_ = tw.Flush()
					if len(shown) < len(a.Models) {
						_, _ = fmt.Fprintf(out, "    … %d more: re-run with --filter <substring> or --all\n", len(a.Models)-len(shown))
					}
				case filter != "":
					_, _ = fmt.Fprintf(out, "  models:  none matching %q\n", filter)
				default:
					_, _ = fmt.Fprintf(out, "  models:  not enumerable\n")
				}
				if a.ModelsNote != "" {
					_, _ = fmt.Fprintf(out, "  note:    %s\n", a.ModelsNote)
				}
				_, _ = fmt.Fprintln(out)
			}

			for _, a := range found {
				if !a.Installed() {
					_, _ = fmt.Fprintf(out, "%s: not installed\n", a.Harness)
				}
			}
			if !any {
				_, _ = fmt.Fprintln(out, "\nNo external agent CLI found on PATH. Native subagents are not probed.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&filter, "filter", "", "only show models whose id contains this substring")
	cmd.Flags().BoolVar(&all, "all", false, "show every model instead of the first few")
	rootCmd.AddCommand(cmd)
}

// matching returns the models containing sub, compared case-insensitively.
func matching(models []string, sub string) []string {
	sub = strings.ToLower(sub)
	var out []string
	for _, m := range models {
		if strings.Contains(strings.ToLower(m), sub) {
			out = append(out, m)
		}
	}
	return out
}

// defaultEffort picks the middle level a harness offers, used only to show a
// copy-pasteable example spec next to each model.
func defaultEffort(a agents.Availability) string {
	if len(a.Efforts) == 0 {
		return "medium"
	}
	return a.Efforts[len(a.Efforts)/2]
}
