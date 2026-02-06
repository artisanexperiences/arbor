package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/artisanexperiences/arbor/internal/scaffold/prompts"
)

// UIDbPrompter implements the DbPrompter interface using huh for terminal UI.
type UIDbPrompter struct{}

// SelectDatabase prompts the user to choose between creating a new database
// or reusing an existing one from another worktree.
func (p UIDbPrompter) SelectDatabase(options []prompts.DatabaseOption) (string, error) {
	if len(options) == 0 {
		return "", nil
	}

	// Build options for the select prompt
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt.Label, opt.DbSuffix)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select database").
				Description("Choose an existing database or create a new one").
				Options(huhOptions...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return "", NormalizeAbort(err)
	}

	return selected, nil
}

// ConfirmMigrations prompts the user to confirm running database migrations.
func (p UIDbPrompter) ConfirmMigrations() (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Run migrations?").
				Description("php artisan migrate:fresh --seed").
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return false, NormalizeAbort(err)
	}

	return confirmed, nil
}

// ConfirmDatabaseDrop prompts the user to confirm dropping databases that match
// the given suffix. Shows the list of databases that will be dropped.
func (p UIDbPrompter) ConfirmDatabaseDrop(suffix string, databases []string) (bool, error) {
	var confirmed bool

	description := fmt.Sprintf("Databases to drop:\n%s", strings.Join(databases, "\n"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Drop databases matching suffix '%s'?", suffix)).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return false, NormalizeAbort(err)
	}

	return confirmed, nil
}
