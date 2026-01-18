package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func createTestRepo(t *testing.T) (string, string) {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	barePath := filepath.Join(tmpDir, ".bare")

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("initializing git repo: %v", err)
	}

	readmePath := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("test"), 0644); err != nil {
		t.Fatalf("writing README: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("staging files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("committing: %v", err)
	}

	cmd = exec.Command("git", "clone", "--bare", repoDir, barePath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("cloning to bare: %v", err)
	}

	return barePath, repoDir
}

func TestBranchExists(t *testing.T) {
	barePath, _ := createTestRepo(t)

	if !BranchExists(barePath, "main") {
		t.Error("main branch should exist after creating from repo with commit")
	}

	if BranchExists(barePath, "nonexistent") {
		t.Error("nonexistent branch should not exist")
	}
}

func TestListBranches(t *testing.T) {
	barePath, _ := createTestRepo(t)

	mainPath := filepath.Join(filepath.Dir(barePath), "main")
	if err := CreateWorktree(barePath, mainPath, "main", ""); err != nil {
		t.Fatalf("creating main worktree: %v", err)
	}

	branches, err := ListBranches(barePath)
	if err != nil {
		t.Fatalf("listing branches: %v", err)
	}

	for _, b := range branches {
		if b == "main" {
			t.Error("main branch (current) should not be in ListBranches output")
		}
	}

	featurePath := filepath.Join(filepath.Dir(barePath), "feature")
	if err := CreateWorktree(barePath, featurePath, "feature", "main"); err != nil {
		t.Fatalf("creating feature worktree: %v", err)
	}

	branches, err = ListBranches(barePath)
	if err != nil {
		t.Fatalf("listing branches: %v", err)
	}

	featureFound := false
	for _, b := range branches {
		if b == "feature" {
			featureFound = true
			break
		}
	}

	if !featureFound {
		t.Error("feature branch should be in list")
	}
}

func TestListAllBranches(t *testing.T) {
	barePath, _ := createTestRepo(t)

	branches, err := ListAllBranches(barePath)
	if err != nil {
		t.Fatalf("listing all branches: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "main" {
			found = true
			break
		}
	}

	if !found {
		t.Error("main branch should be in list")
	}
}

func TestListRemoteBranches(t *testing.T) {
	barePath, _ := createTestRepo(t)

	branches, err := ListRemoteBranches(barePath)
	if err != nil {
		t.Fatalf("listing remote branches: %v", err)
	}

	if len(branches) != 0 {
		t.Errorf("expected 0 remote branches, got %d", len(branches))
	}
}

func TestFindBarePath(t *testing.T) {
	barePath, _ := createTestRepo(t)

	mainPath := filepath.Join(filepath.Dir(barePath), "main")
	if err := CreateWorktree(barePath, mainPath, "main", ""); err != nil {
		t.Fatalf("creating main worktree: %v", err)
	}

	found, err := FindBarePath(mainPath)
	if err != nil {
		t.Fatalf("finding bare path: %v", err)
	}

	if found != barePath {
		t.Errorf("expected %s, got %s", barePath, found)
	}

	_, err = FindBarePath("/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestIsMerged(t *testing.T) {
	barePath, _ := createTestRepo(t)

	mainPath := filepath.Join(filepath.Dir(barePath), "main")
	if err := CreateWorktree(barePath, mainPath, "main", ""); err != nil {
		t.Fatalf("creating main worktree: %v", err)
	}

	featurePath := filepath.Join(filepath.Dir(barePath), "feature")
	if err := CreateWorktree(barePath, featurePath, "feature", "main"); err != nil {
		t.Fatalf("creating feature worktree: %v", err)
	}

	cmd := exec.Command("git", "checkout", "-b", "dev")
	cmd.Dir = featurePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("creating dev branch: %v", err)
	}

	readmePath := filepath.Join(featurePath, "README.md")
	if err := os.WriteFile(readmePath, []byte("test\nfeature"), 0644); err != nil {
		t.Fatalf("writing README: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = featurePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("staging files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Feature commit")
	cmd.Dir = featurePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("committing: %v", err)
	}

	merged, err := IsMerged(barePath, "main", "main")
	if err != nil {
		t.Fatalf("checking merge status: %v", err)
	}
	if !merged {
		t.Error("main should be merged into main")
	}

	merged, err = IsMerged(barePath, "dev", "main")
	if err != nil {
		t.Fatalf("checking merge status: %v", err)
	}
	if merged {
		t.Error("dev should not be merged into main yet (no commits on dev)")
	}

	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = mainPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("switching to main: %v", err)
	}

	cmd = exec.Command("git", "merge", "dev", "--no-edit")
	cmd.Dir = mainPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("merging dev into main: %v", err)
	}

	merged, err = IsMerged(barePath, "dev", "main")
	if err != nil {
		t.Fatalf("checking merge status after merge: %v", err)
	}
	if !merged {
		t.Error("dev should be merged into main after merge")
	}
}

func TestFindBarePathParentSearch(t *testing.T) {
	barePath, _ := createTestRepo(t)

	mainPath := filepath.Join(filepath.Dir(barePath), "main")
	if err := CreateWorktree(barePath, mainPath, "main", ""); err != nil {
		t.Fatalf("creating main worktree: %v", err)
	}

	subdirPath := filepath.Join(mainPath, "subdir")
	if err := os.MkdirAll(subdirPath, 0755); err != nil {
		t.Fatalf("creating subdir: %v", err)
	}

	found, err := FindBarePath(subdirPath)
	if err != nil {
		t.Fatalf("finding bare path from subdir: %v", err)
	}

	if found != barePath {
		t.Errorf("expected %s, got %s", barePath, found)
	}
}
