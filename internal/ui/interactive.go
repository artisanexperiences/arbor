package ui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/michaeldyrynda/arbor/internal/git"
)

func SelectBranchInteractive(barePath string, localBranches, remoteBranches []string) (string, error) {
	var selected string

	options := []huh.Option[string]{
		huh.NewOption("Create new branch...", "__new__"),
	}

	for _, b := range localBranches {
		options = append(options, huh.NewOption(b, b))
	}

	for _, b := range remoteBranches {
		options = append(options, huh.NewOption("↓ "+b, b))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a branch").
				Description("Choose an existing branch or create a new one").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return "", err
	}

	if selected == "__new__" {
		return PromptNewBranch()
	}

	return selected, nil
}

func PromptNewBranch() (string, error) {
	var name string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New branch name").
				Placeholder("feature/my-feature").
				Value(&name).
				Validate(validateBranchName),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return "", err
	}

	return name, nil
}

func validateBranchName(s string) error {
	if s == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if len(s) < 2 {
		return fmt.Errorf("branch name must be at least 2 characters")
	}
	return nil
}

func SelectWorktreesToPrune(removable []git.Worktree) ([]git.Worktree, error) {
	if len(removable) == 0 {
		return nil, nil
	}

	options := make([]huh.Option[string], len(removable))
	for i, wt := range removable {
		label := fmt.Sprintf("%s (%s)", wt.Branch, filepath.Base(wt.Path))
		options[i] = huh.NewOption(label, wt.Branch)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select worktrees to remove").
				Description("Space to toggle, Enter to confirm").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		return nil, nil
	}

	var result []git.Worktree
	for _, branch := range selected {
		for _, wt := range removable {
			if wt.Branch == branch {
				result = append(result, wt)
				break
			}
		}
	}

	return result, nil
}

func ConfirmRemoval(count int) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Remove worktrees").
				Description(fmt.Sprintf("Remove %d selected worktree(s)?", count)).
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}
