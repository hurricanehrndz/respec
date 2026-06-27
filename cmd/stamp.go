package cmd

import (
	"fmt"
	"os"

	"github.com/chernand/respec/internal/frontmatter"
	"github.com/chernand/respec/internal/store"
	"github.com/spf13/cobra"
)

// specSHAKey is the plan.md frontmatter key holding the stamped spec.md hash.
const specSHAKey = "spec_sha256"

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stamp <change-dir>",
		Short: "Record spec.md's sha256 in plan.md frontmatter",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			changeDir := args[0]
			if err := store.ValidateChangeDir(changeDir); err != nil {
				return err
			}
			sum, err := store.SHA256File(store.SpecPath(changeDir))
			if err != nil {
				return fmt.Errorf("hashing spec.md: %w", err)
			}
			planPath := store.PlanPath(changeDir)
			data, err := os.ReadFile(planPath)
			if err != nil {
				return err
			}
			updated, err := frontmatter.SetString(data, specSHAKey, sum)
			if err != nil {
				return err
			}
			if err := os.WriteFile(planPath, updated, 0o644); err != nil {
				return err
			}
			fmt.Printf("stamped %s: %s=%s\n", planPath, specSHAKey, sum)
			return nil
		},
	})
}
