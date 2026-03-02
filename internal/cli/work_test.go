package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/artisanexperiences/arbor/internal/git"
)

// TestWorkCommand_RemoteBranchResolution verifies that when a remote branch like
// "origin/feature/foo" is passed to the work command it results in a local branch
// checked out in the worktree rather than a detached HEAD.
func TestWorkCommand_RemoteBranchResolution(t *testing.T) {
	// Create a source repo with a feature branch
	sourceDir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = sourceDir
		requireNoError(t, cmd.Run())
	}

	readmePath := filepath.Join(sourceDir, "README.md")
	requireNoError(t, os.WriteFile(readmePath, []byte("test"), 0644))
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "Initial commit"},
		{"checkout", "-b", "feature/remote-branch"},
		{"checkout", "main"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = sourceDir
		requireNoError(t, cmd.Run())
	}

	// Clone to bare repo (simulating arbor init)
	projectDir := t.TempDir()
	barePath := filepath.Join(projectDir, ".bare")
	cmd := exec.Command("git", "clone", "--bare", sourceDir, barePath)
	requireNoError(t, cmd.Run())
	requireNoError(t, git.ConfigureFetchRefspec(barePath, sourceDir))

	// Fetch so remote tracking branches are available
	cmd = exec.Command("git", "-C", barePath, "fetch", "origin")
	requireNoError(t, cmd.Run())

	// Simulate the remote branch resolution logic from work.go:
	// user selects "origin/feature/remote-branch" from the interactive list.
	selectedBranch := "origin/feature/remote-branch"

	remotes, err := git.ListRemotes(barePath)
	assert.NoError(t, err)

	branch := selectedBranch
	baseBranch := ""
	if idx := strings.IndexByte(branch, '/'); idx != -1 {
		remote := branch[:idx]
		localBranch := branch[idx+1:]
		for _, r := range remotes {
			if r == remote {
				baseBranch = branch
				branch = localBranch
				break
			}
		}
	}

	assert.Equal(t, "feature/remote-branch", branch, "branch should have remote prefix stripped")
	assert.Equal(t, "origin/feature/remote-branch", baseBranch, "baseBranch should be the original remote ref")

	// Create the worktree using the resolved branch name and remote base
	worktreePath := filepath.Join(projectDir, "feature-remote-branch")
	requireNoError(t, git.CreateWorktree(barePath, worktreePath, branch, baseBranch))

	// Verify the worktree is on a named branch (not detached HEAD)
	cmd = exec.Command("git", "-C", worktreePath, "branch", "--show-current")
	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, "feature/remote-branch", strings.TrimSpace(string(output)),
		"worktree should be on a local branch, not in detached HEAD state")

	// Verify it appears in ListWorktrees (which excludes detached HEAD)
	worktrees, err := git.ListWorktrees(barePath)
	assert.NoError(t, err)
	found := false
	for _, wt := range worktrees {
		if wt.Branch == "feature/remote-branch" {
			found = true
			break
		}
	}
	assert.True(t, found, "worktree should be listed by arbor list after creation from remote branch")
}

func TestWorkCommand_SetsUpBranchTracking(t *testing.T) {
	// Create a source repo
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

	// Create initial commit
	readmePath := filepath.Join(sourceDir, "README.md")
	requireNoError(t, os.WriteFile(readmePath, []byte("test"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = sourceDir
	requireNoError(t, cmd.Run())

	// Clone to bare repo and set up like arbor init does
	projectDir := t.TempDir()
	barePath := filepath.Join(projectDir, ".bare")
	cmd = exec.Command("git", "clone", "--bare", sourceDir, barePath)
	requireNoError(t, cmd.Run())

	// Configure fetch refspec
	requireNoError(t, git.ConfigureFetchRefspec(barePath, sourceDir))

	// Create main worktree
	mainPath := filepath.Join(projectDir, "main")
	requireNoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))

	// Set up tracking for main (this is what would happen in work command)
	requireNoError(t, git.SetBranchUpstream(barePath, "main", "origin"))

	// Verify main has tracking
	hasTracking, err := git.HasBranchTracking(barePath, "main")
	assert.NoError(t, err)
	assert.True(t, hasTracking)

	// Create feature branch worktree
	featurePath := filepath.Join(projectDir, "feature")
	requireNoError(t, git.CreateWorktree(barePath, featurePath, "feature", "main"))

	// Set up tracking for feature (this is what would happen in work command)
	requireNoError(t, git.SetBranchUpstream(barePath, "feature", "origin"))

	// Verify feature has tracking
	hasTracking, err = git.HasBranchTracking(barePath, "feature")
	assert.NoError(t, err)
	assert.True(t, hasTracking)

	// Check the tracking config
	cmd = exec.Command("git", "-C", barePath, "config", "--get", "branch.feature.remote")
	output, err := cmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, "origin", strings.TrimSpace(string(output)))

	cmd = exec.Command("git", "-C", barePath, "config", "--get", "branch.feature.merge")
	output, err = cmd.Output()
	assert.NoError(t, err)
	assert.Equal(t, "refs/heads/feature", strings.TrimSpace(string(output)))
}
