package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"github.com/hurricanehrndz/respec/internal/store"
	"github.com/spf13/cobra"
)

type statusResult struct {
	State    string `json:"state"`
	Change   string `json:"change"`
	SpecSHA  string `json:"spec_sha256"`
	Recorded string `json:"recorded"`
}

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status <change-dir>",
		Short: "Report plan staleness (fresh / stale / unstamped)",
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

			if asJSON {
				enc := json.NewEncoder(c.OutOrStdout())
				return enc.Encode(res)
			}
			fmt.Fprintln(c.OutOrStdout(), res.State)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(cmd)
}
