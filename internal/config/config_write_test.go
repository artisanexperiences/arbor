package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteWorktreeConfig(t *testing.T) {
	t.Run("creates new config file", func(t *testing.T) {
		tmpDir := t.TempDir()

		data := map[string]string{
			"db_suffix": "test123",
		}

		err := WriteWorktreeConfig(tmpDir, data)
		if err != nil {
			t.Fatalf("WriteWorktreeConfig failed: %v", err)
		}

		// Verify file was created
		configPath := filepath.Join(tmpDir, "arbor.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Verify content by reading it back
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if string(content) == "" {
			t.Error("config file is empty")
		}
	})

	t.Run("merges with existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "arbor.yaml")

		// Create initial config
		initialContent := `site_name: mysite
db_suffix: old123
`
		if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
			t.Fatalf("failed to create initial config: %v", err)
		}

		// Write new data
		data := map[string]string{
			"db_suffix": "new456",
		}

		err := WriteWorktreeConfig(tmpDir, data)
		if err != nil {
			t.Fatalf("WriteWorktreeConfig failed: %v", err)
		}

		// Read back and verify
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !contains(contentStr, "db_suffix: new456") {
			t.Errorf("expected new db_suffix value not found in:\n%s", contentStr)
		}
		if !contains(contentStr, "site_name: mysite") {
			t.Errorf("expected existing site_name to be preserved in:\n%s", contentStr)
		}
	})

	t.Run("multiple writes accumulate data", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First write
		err := WriteWorktreeConfig(tmpDir, map[string]string{"key1": "value1"})
		if err != nil {
			t.Fatalf("first write failed: %v", err)
		}

		// Second write
		err = WriteWorktreeConfig(tmpDir, map[string]string{"key2": "value2"})
		if err != nil {
			t.Fatalf("second write failed: %v", err)
		}

		// Read back and verify both keys exist
		configPath := filepath.Join(tmpDir, "arbor.yaml")
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !contains(contentStr, "key1: value1") {
			t.Errorf("expected key1 not found in:\n%s", contentStr)
		}
		if !contains(contentStr, "key2: value2") {
			t.Errorf("expected key2 not found in:\n%s", contentStr)
		}
	})
}

func TestSaveProject(t *testing.T) {
	t.Run("creates new project config", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &Config{
			SiteName:      "MyProject",
			Preset:        "laravel",
			DefaultBranch: "main",
		}

		err := SaveProject(tmpDir, cfg)
		if err != nil {
			t.Fatalf("SaveProject failed: %v", err)
		}

		// Verify file was created
		configPath := filepath.Join(tmpDir, "arbor.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Load it back and verify
		loaded, err := LoadProject(tmpDir)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		if loaded.SiteName != "MyProject" {
			t.Errorf("expected SiteName 'MyProject', got '%s'", loaded.SiteName)
		}
		if loaded.Preset != "laravel" {
			t.Errorf("expected Preset 'laravel', got '%s'", loaded.Preset)
		}
		if loaded.DefaultBranch != "main" {
			t.Errorf("expected DefaultBranch 'main', got '%s'", loaded.DefaultBranch)
		}
	})

	t.Run("preserves existing config data", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "arbor.yaml")

		// Create initial config with extra fields
		initialContent := `site_name: OldSite
preset: old_preset
default_branch: old_branch
custom_field: custom_value
`
		if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
			t.Fatalf("failed to create initial config: %v", err)
		}

		// Save with only some fields updated
		cfg := &Config{
			SiteName: "NewSite",
			// Preset and DefaultBranch left empty - should preserve existing
		}

		err := SaveProject(tmpDir, cfg)
		if err != nil {
			t.Fatalf("SaveProject failed: %v", err)
		}

		// Read back the raw content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !contains(contentStr, "site_name: NewSite") {
			t.Errorf("expected updated site_name not found in:\n%s", contentStr)
		}
		if !contains(contentStr, "preset: old_preset") {
			t.Errorf("expected preserved preset not found in:\n%s", contentStr)
		}
		if !contains(contentStr, "default_branch: old_branch") {
			t.Errorf("expected preserved default_branch not found in:\n%s", contentStr)
		}
	})

	t.Run("round-trip preserves data", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Save initial config
		cfg := &Config{
			SiteName:      "TestProject",
			Preset:        "php",
			DefaultBranch: "develop",
		}

		err := SaveProject(tmpDir, cfg)
		if err != nil {
			t.Fatalf("SaveProject failed: %v", err)
		}

		// Load it back
		loaded, err := LoadProject(tmpDir)
		if err != nil {
			t.Fatalf("LoadProject failed: %v", err)
		}

		// Verify all fields
		if loaded.SiteName != "TestProject" {
			t.Errorf("SiteName mismatch: expected 'TestProject', got '%s'", loaded.SiteName)
		}
		if loaded.Preset != "php" {
			t.Errorf("Preset mismatch: expected 'php', got '%s'", loaded.Preset)
		}
		if loaded.DefaultBranch != "develop" {
			t.Errorf("DefaultBranch mismatch: expected 'develop', got '%s'", loaded.DefaultBranch)
		}
	})
}

func TestConfigRoundTrip(t *testing.T) {
	t.Run("complex config round-trip", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Save project config
		projectCfg := &Config{
			SiteName:      "ComplexProject",
			Preset:        "laravel",
			DefaultBranch: "main",
		}
		err := SaveProject(tmpDir, projectCfg)
		if err != nil {
			t.Fatalf("SaveProject failed: %v", err)
		}

		// Write worktree config
		err = WriteWorktreeConfig(tmpDir, map[string]string{
			"db_suffix": "abc123",
		})
		if err != nil {
			t.Fatalf("WriteWorktreeConfig failed: %v", err)
		}

		// Load project config - should still work
		loaded, err := LoadProject(tmpDir)
		if err != nil {
			t.Fatalf("LoadProject failed: %v", err)
		}

		if loaded.SiteName != "ComplexProject" {
			t.Errorf("SiteName mismatch: expected 'ComplexProject', got '%s'", loaded.SiteName)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
