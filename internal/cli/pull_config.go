package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/artisanexperiences/arbor/internal/ui"
)

var pullConfigCmd = &cobra.Command{
	Use:   "pull-config",
	Short: "Update project config from the default branch worktree",
	Long: `Copies arbor.yaml from the default branch worktree to the project root.

Use this command when the repository arbor.yaml (committed in the default branch)
has been updated and you want to pull those changes into the project-level config.

This replaces the project arbor.yaml entirely with the one from the default branch worktree.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pc, err := OpenProjectFromCWD()
		if err != nil {
			return err
		}

		dryRun := mustGetBool(cmd, "dry-run")
		force := mustGetBool(cmd, "force")
		verbose := mustGetBool(cmd, "verbose")
		quiet := mustGetBool(cmd, "quiet")

		sourcePath := filepath.Join(pc.ProjectPath, pc.DefaultBranch, "arbor.yaml")
		destPath := filepath.Join(pc.ProjectPath, "arbor.yaml")

		// Verify source exists
		sourceBytes, err := os.ReadFile(sourcePath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no arbor.yaml found in %s worktree", pc.DefaultBranch)
			}
			return fmt.Errorf("reading %s: %w", sourcePath, err)
		}

		// Check if destination exists and is already identical
		destBytes, err := os.ReadFile(destPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("reading project config: %w", err)
		}

		if bytes.Equal(sourceBytes, destBytes) {
			if !quiet {
				ui.PrintInfo("Already up to date")
			}
			return nil
		}

		if dryRun {
			ui.PrintStep(fmt.Sprintf("Would copy %s/arbor.yaml to project arbor.yaml", pc.DefaultBranch))
			return nil
		}

		if !force {
			if verbose && !quiet {
				ui.PrintStep(fmt.Sprintf("Pull config from %s/arbor.yaml to project arbor.yaml?", pc.DefaultBranch))
			}
			confirmed, err := ui.Confirm(fmt.Sprintf("Pull config from %s/arbor.yaml to project arbor.yaml?", pc.DefaultBranch))
			if err != nil {
				return fmt.Errorf("confirmation prompt: %w", err)
			}
			if !confirmed {
				if !quiet {
					ui.PrintInfo("Aborted")
				}
				return nil
			}
		}

		if verbose && !quiet {
			ui.PrintStep(fmt.Sprintf("Copying config from %s worktree to project root", pc.DefaultBranch))
		}

		if err := os.WriteFile(destPath, sourceBytes, 0644); err != nil {
			return fmt.Errorf("writing project config: %w", err)
		}

		if !quiet {
			ui.PrintSuccess("Project config updated")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullConfigCmd)

	pullConfigCmd.Flags().Bool("dry-run", false, "Show what would be copied without making changes")
	pullConfigCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	pullConfigCmd.Flags().BoolP("verbose", "v", false, "Show detailed output")
	pullConfigCmd.Flags().BoolP("quiet", "q", false, "Suppress non-essential output")
}
