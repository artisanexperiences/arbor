package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// StashAll creates a stash including untracked and ignored files
// This captures tracked modifications, untracked files, and ignored files
func StashAll(worktreePath string, message string) error {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "push", "--include-untracked", "--all", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// Check if the error is because there's nothing to stash
		if strings.Contains(outputStr, "No local changes to save") {
			return nil // Not an error, just nothing to stash
		}
		return fmt.Errorf("git stash failed: %w\n%s", err, outputStr)
	}
	return nil
}

// PopStash pops the most recent stash
// Returns an error if there are conflicts or if the pop fails
func PopStash(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "pop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// Check if it's a conflict error
		if strings.Contains(outputStr, "CONFLICT") || strings.Contains(outputStr, "conflict") {
			return &StashConflictError{Output: outputStr}
		}
		return fmt.Errorf("git stash pop failed: %w\n%s", err, outputStr)
	}
	return nil
}

// HasStash checks if there are any stashes in the repository
func HasStash(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("checking stash list: %w", err)
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// HasChanges checks if there are any changes that would be captured by stash
// This includes tracked modifications, untracked files, and ignored files
func HasChanges(worktreePath string) (bool, error) {
	// Check for any files in working tree (tracked, untracked, ignored)
	// Using --ignored to also check ignored files
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain", "--ignored")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("checking for changes: %w", err)
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// StashConflictError represents a stash pop that failed due to conflicts
type StashConflictError struct {
	Output string
}

func (e *StashConflictError) Error() string {
	return fmt.Sprintf("stash pop has conflicts:\n%s\n\nResolve the conflicts, stage the changes with 'git add', then run 'git stash drop' to remove the stash, or run 'git reset --hard && git stash pop' to try again", e.Output)
}
