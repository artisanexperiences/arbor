package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/michaeldyrynda/arbor/internal/git"
	"github.com/michaeldyrynda/arbor/internal/ui"
	"github.com/michaeldyrynda/arbor/internal/utils"
)

var workCmd = &cobra.Command{
	Use:   "work [BRANCH] [PATH]",
	Short: "Create or checkout a feature worktree",
	Long: `Creates or checks out a new worktree for a feature branch.

Arguments:
  BRANCH  Name of the feature branch
  PATH    Optional custom path (defaults to sanitised branch name)

If no branch is provided, interactive mode allows selection from
available branches or entering a new branch name.`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pc, err := OpenProjectFromCWD()
		if err != nil {
			return err
		}

		baseBranch := mustGetString(cmd, "base")
		interactive := mustGetBool(cmd, "interactive")
		dryRun := mustGetBool(cmd, "dry-run")
		verbose := mustGetBool(cmd, "verbose")

		var branch string
		if len(args) > 0 {
			branch = args[0]
		} else if interactive {
			localBranches, err := git.ListAllBranches(pc.BarePath)
			if err != nil {
				return fmt.Errorf("listing local branches: %w", err)
			}

			remoteBranches, _ := git.ListRemoteBranches(pc.BarePath)

			selected, err := ui.SelectBranchInteractive(pc.BarePath, localBranches, remoteBranches)
			if err != nil {
				return fmt.Errorf("selecting branch: %w", err)
			}
			branch = selected
		}

		if branch == "" {
			return fmt.Errorf("branch name is required (or use --interactive)")
		}

		if baseBranch == "" {
			baseBranch = pc.DefaultBranch
		}

		worktreePath := ""
		if len(args) > 1 {
			worktreePath = args[1]
		} else {
			worktreePath = filepath.Join(pc.ProjectPath, utils.SanitisePath(branch))
		}

		absWorktreePath, err := filepath.Abs(worktreePath)
		if err != nil {
			return fmt.Errorf("getting absolute path: %w", err)
		}

		exists := git.BranchExists(pc.BarePath, branch)
		if exists {
			worktrees, err := git.ListWorktrees(pc.BarePath)
			if err != nil {
				return fmt.Errorf("listing worktrees: %w", err)
			}
			for _, wt := range worktrees {
				if wt.Branch == branch {
					fmt.Printf("Worktree already exists at %s\n", wt.Path)
					return nil
				}
			}
		}

		fmt.Printf("Creating worktree for branch '%s' from '%s'\n", branch, baseBranch)
		fmt.Printf("Path: %s\n", absWorktreePath)

		if !dryRun {
			if err := git.CreateWorktree(pc.BarePath, absWorktreePath, branch, baseBranch); err != nil {
				return fmt.Errorf("creating worktree: %w", err)
			}
		} else {
			fmt.Println("[DRY RUN] Would create worktree")
		}

		if !dryRun {
			preset := pc.Config.Preset
			if preset == "" {
				preset = pc.PresetManager().Detect(absWorktreePath)
			}

			if verbose {
				fmt.Printf("Running scaffold for preset: %s\n", preset)
			}

			repoName := filepath.Base(filepath.Dir(absWorktreePath))
			if err := pc.ScaffoldManager().RunScaffold(absWorktreePath, branch, repoName, preset, pc.Config, false, verbose); err != nil {
				fmt.Printf("Warning: scaffold steps failed: %v\n", err)
			}
		} else {
			fmt.Println("[DRY RUN] Would run scaffold steps")
		}

		fmt.Printf("\nWorktree ready at %s\n", absWorktreePath)
		return nil
	},
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func init() {
	rootCmd.AddCommand(workCmd)

	workCmd.Flags().StringP("base", "b", "", "Base branch for new worktree")
	workCmd.Flags().Bool("interactive", false, "Interactive branch selection")
}
