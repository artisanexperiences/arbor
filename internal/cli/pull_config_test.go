package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/artisanexperiences/arbor/internal/git"
)

// setupPullConfigProject creates a minimal arbor project with:
// - a source repo cloned to .bare
// - a main worktree at <projectDir>/main
// - an optional arbor.yaml in the main worktree
// - an optional arbor.yaml at the project root
//
// Returns (projectDir, mainWorktreePath).
func setupPullConfigProject(t *testing.T, worktreeConfig, projectConfig string) (string, string) {
	t.Helper()

	sourceDir := t.TempDir()

	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	readmePath := filepath.Join(sourceDir, "README.md")
	requireNoError(t, os.WriteFile(readmePath, []byte("test"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	projectDir := t.TempDir()
	barePath := filepath.Join(projectDir, ".bare")

	cmd = exec.Command("git", "clone", "--bare", sourceDir, barePath)
	requireNoError(t, cmd.Run())

	mainPath := filepath.Join(projectDir, "main")
	requireNoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))

	if worktreeConfig != "" {
		requireNoError(t, os.WriteFile(filepath.Join(mainPath, "arbor.yaml"), []byte(worktreeConfig), 0644))
	}

	if projectConfig != "" {
		requireNoError(t, os.WriteFile(filepath.Join(projectDir, "arbor.yaml"), []byte(projectConfig), 0644))
	}

	return projectDir, mainPath
}

func ensurePullConfigTestFlags(t *testing.T) {
	t.Helper()

	if pullConfigCmd.Flags().Lookup("dry-run") == nil {
		pullConfigCmd.Flags().Bool("dry-run", false, "")
	}
	if pullConfigCmd.Flags().Lookup("force") == nil {
		pullConfigCmd.Flags().BoolP("force", "f", false, "")
	}
	if pullConfigCmd.Flags().Lookup("verbose") == nil {
		pullConfigCmd.Flags().BoolP("verbose", "v", false, "")
	}
	if pullConfigCmd.Flags().Lookup("quiet") == nil {
		pullConfigCmd.Flags().BoolP("quiet", "q", false, "")
	}
}

func TestPullConfig_Success(t *testing.T) {
	ensurePullConfigTestFlags(t)

	worktreeConfig := "site_name: my-site\ndefault_branch: main\npreset: laravel\n"
	projectConfig := "site_name: my-site\ndefault_branch: main\n"

	projectDir, _ := setupPullConfigProject(t, worktreeConfig, projectConfig)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	requireNoError(t, os.Chdir(projectDir))

	requireNoError(t, pullConfigCmd.Flags().Set("force", "true"))
	requireNoError(t, pullConfigCmd.Flags().Set("dry-run", "false"))
	defer func() {
		requireNoError(t, pullConfigCmd.Flags().Set("force", "false"))
	}()

	err := pullConfigCmd.RunE(pullConfigCmd, []string{})
	assert.NoError(t, err)

	// Verify project arbor.yaml now matches worktree arbor.yaml
	projectConfigBytes, err := os.ReadFile(filepath.Join(projectDir, "arbor.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, worktreeConfig, string(projectConfigBytes))
}

func TestPullConfig_NoWorktreeConfig(t *testing.T) {
	ensurePullConfigTestFlags(t)

	// No worktreeConfig — the main worktree has no arbor.yaml
	projectConfig := "site_name: my-site\ndefault_branch: main\n"
	projectDir, _ := setupPullConfigProject(t, "", projectConfig)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	requireNoError(t, os.Chdir(projectDir))

	requireNoError(t, pullConfigCmd.Flags().Set("force", "true"))
	requireNoError(t, pullConfigCmd.Flags().Set("dry-run", "false"))
	defer func() {
		requireNoError(t, pullConfigCmd.Flags().Set("force", "false"))
	}()

	err := pullConfigCmd.RunE(pullConfigCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no arbor.yaml found in")
}

func TestPullConfig_AlreadyUpToDate(t *testing.T) {
	ensurePullConfigTestFlags(t)

	// Both files have the same content
	config := "site_name: my-site\ndefault_branch: main\npreset: laravel\n"
	projectDir, _ := setupPullConfigProject(t, config, config)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	requireNoError(t, os.Chdir(projectDir))

	requireNoError(t, pullConfigCmd.Flags().Set("force", "true"))
	requireNoError(t, pullConfigCmd.Flags().Set("dry-run", "false"))
	defer func() {
		requireNoError(t, pullConfigCmd.Flags().Set("force", "false"))
	}()

	// Record the modification time before running
	projectConfigPath := filepath.Join(projectDir, "arbor.yaml")
	statBefore, err := os.Stat(projectConfigPath)
	assert.NoError(t, err)

	err = pullConfigCmd.RunE(pullConfigCmd, []string{})
	assert.NoError(t, err)

	// Verify the file was NOT rewritten (same modification time)
	statAfter, err := os.Stat(projectConfigPath)
	assert.NoError(t, err)
	assert.Equal(t, statBefore.ModTime(), statAfter.ModTime(), "file should not be modified when already up to date")
}

func TestPullConfig_DryRun(t *testing.T) {
	ensurePullConfigTestFlags(t)

	worktreeConfig := "site_name: my-site\ndefault_branch: main\npreset: laravel\n"
	projectConfig := "site_name: my-site\ndefault_branch: main\n"
	projectDir, _ := setupPullConfigProject(t, worktreeConfig, projectConfig)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	requireNoError(t, os.Chdir(projectDir))

	requireNoError(t, pullConfigCmd.Flags().Set("force", "true"))
	requireNoError(t, pullConfigCmd.Flags().Set("dry-run", "true"))
	defer func() {
		requireNoError(t, pullConfigCmd.Flags().Set("force", "false"))
		requireNoError(t, pullConfigCmd.Flags().Set("dry-run", "false"))
	}()

	err := pullConfigCmd.RunE(pullConfigCmd, []string{})
	assert.NoError(t, err)

	// Verify the project arbor.yaml was NOT changed
	projectConfigBytes, err := os.ReadFile(filepath.Join(projectDir, "arbor.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, projectConfig, string(projectConfigBytes), "dry-run should not modify the project config")
}

func TestPullConfig_NotInProject(t *testing.T) {
	ensurePullConfigTestFlags(t)

	// Run from a plain temp dir with no arbor project
	plainDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	requireNoError(t, os.Chdir(plainDir))

	err := pullConfigCmd.RunE(pullConfigCmd, []string{})
	assert.Error(t, err)
}
