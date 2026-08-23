package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/internal/update"
)

func updateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update codebahn to the latest version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version == "dev" {
				return fmt.Errorf("cannot update a development build; install a release build first")
			}

			if checkOnly {
				rel, err := update.CheckLatest(version)
				if err != nil {
					return fmt.Errorf("checking for updates: %w", err)
				}
				if rel.Newer {
					fmt.Printf("Update available: v%s (current: %s).\n", rel.Version, version)
				} else {
					fmt.Println("Already up to date.")
				}
				return nil
			}

			rel, err := update.CheckLatest(version)
			if err != nil {
				return fmt.Errorf("checking for updates: %w", err)
			}
			if !rel.Newer {
				fmt.Println("Already up to date.")
				return nil
			}

			fmt.Printf("Updating to v%s...\n", rel.Version)
			if err := update.Update(version, ""); err != nil {
				return err
			}
			fmt.Printf("Updated to v%s.\n", rel.Version)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without installing")
	return cmd
}
