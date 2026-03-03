package steps

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/artisanexperiences/arbor/internal/config"
	"github.com/artisanexperiences/arbor/internal/git"
	"github.com/artisanexperiences/arbor/internal/scaffold/prompts"
	"github.com/artisanexperiences/arbor/internal/scaffold/types"
)

// mockDbPrompter records calls to ConfirmMigrations for assertion in tests.
type mockDbPrompter struct {
	confirmMigrationsCall string // the databaseName passed to ConfirmMigrations
	confirmResult         bool
}

func (m *mockDbPrompter) SelectDatabase(options []prompts.DatabaseOption) (string, error) {
	return "", nil
}

func (m *mockDbPrompter) ConfirmMigrations(databaseName string) (bool, error) {
	m.confirmMigrationsCall = databaseName
	return m.confirmResult, nil
}

func (m *mockDbPrompter) ConfirmDatabaseDrop(suffix string, databases []string) (bool, error) {
	return true, nil
}

func TestDbCreateStep(t *testing.T) {
	t.Run("name returns db.create", func(t *testing.T) {
		step := NewDbCreateStep(config.StepConfig{})
		assert.Equal(t, "db.create", step.Name())
	})

	t.Run("condition always returns true - controlled by preset", func(t *testing.T) {
		step := NewDbCreateStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: t.TempDir(),
		}
		assert.True(t, step.Condition(ctx))
	})

	t.Run("skips when no DB_CONNECTION in env file and no type config", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("APP_NAME=test\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		step := NewDbCreateStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
	})

	t.Run("creates database with mock client", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set after db.create")
		assert.Equal(t, 1, mockClient.DatabaseCount(), "Should have created one database")
	})

	t.Run("auto-detects mysql engine from DB_CONNECTION env", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set after db.create")
	})

	t.Run("auto-detects pgsql engine from DB_CONNECTION env", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=pgsql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set after db.create")
	})

	t.Run("uses explicit type config over env detection", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{Type: "pgsql"}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set after db.create")
	})

	t.Run("generates database name with site name and suffix", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "my-app",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)

		suffix := ctx.GetDbSuffix()
		assert.NotEmpty(t, suffix, "DbSuffix should be set")

		parts := strings.Split(suffix, "_")
		assert.Len(t, parts, 2, "Suffix should be in format {adjective}_{noun}")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1, "Should have one create call")
		assert.True(t, strings.HasPrefix(createCalls[0], "my_app_"), "Database name should start with sanitized site name")
	})

	t.Run("writes DbSuffix to local state", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		require.NoError(t, err)

		suffix := ctx.GetDbSuffix()
		assert.NotEmpty(t, suffix, "DbSuffix should be set in context")

		localState, err := config.ReadLocalState(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, suffix, localState.DbSuffix, "DbSuffix should be persisted to .arbor.local")
	})

	t.Run("reads APP_NAME from .env if SiteName is empty", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\nAPP_NAME=myapp\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set even with empty SiteName")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1)
		assert.True(t, strings.HasPrefix(createCalls[0], "myapp_"), "Should use APP_NAME from .env")
	})

	t.Run("sanitizes site name for database generation", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "My Test-App!",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix(), "DbSuffix should be set")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1)
		assert.True(t, strings.HasPrefix(createCalls[0], "my_test_app_"), "Site name should be sanitized")
	})

	t.Run("creates SQLite database file", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=sqlite\nDB_DATABASE=database/test.sqlite\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		step := NewDbCreateStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: true})
		assert.NoError(t, err)

		dbFile := filepath.Join(tmpDir, "database", "test.sqlite")
		assert.FileExists(t, dbFile)
	})

	t.Run("SQLite does not set DbSuffix", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=sqlite\nDB_DATABASE=database/test.sqlite\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		step := NewDbCreateStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.Empty(t, ctx.GetDbSuffix(), "DbSuffix should not be set for SQLite")
	})

	t.Run("creates database with custom prefix", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{
			Args: []string{"--prefix", "mycustom"},
		}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)

		suffix := ctx.GetDbSuffix()
		assert.NotEmpty(t, suffix, "DbSuffix should be set")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1)
		assert.True(t, strings.HasPrefix(createCalls[0], "mycustom_"), "Should use custom prefix")

		localState, err := config.ReadLocalState(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, suffix, localState.DbSuffix, "Suffix should be persisted to local state")
	})

	t.Run("creates database without prefix uses siteName", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.NotEmpty(t, ctx.GetDbSuffix())

		createCalls := mockClient.GetCreateCalls()
		assert.True(t, strings.HasPrefix(createCalls[0], "myapp_"))
	})

	t.Run("db.create uses existing suffix from context", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}
		ctx.SetDbSuffix("preexisting_suffix")

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.Equal(t, "preexisting_suffix", ctx.GetDbSuffix(), "Should use preexisting suffix from context")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1)
		assert.Equal(t, "testapp_preexisting_suffix", createCalls[0], "Should use preexisting suffix")

		localState, err := config.ReadLocalState(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "preexisting_suffix", localState.DbSuffix, "Should persist preexisting suffix to local state")
	})

	t.Run("db.create with prefix uses existing suffix", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}
		ctx.SetDbSuffix("shared_suffix")

		step := NewDbCreateStepWithFactory(config.StepConfig{
			Args: []string{"--prefix", "app"},
		}, MockClientFactory(mockClient))

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.Equal(t, "shared_suffix", ctx.GetDbSuffix(), "Should use shared suffix from context")

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 1)
		assert.Equal(t, "app_shared_suffix", createCalls[0], "Should use prefix with shared suffix")
	})

	t.Run("retries on database exists error", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.SetExistsOnFirstNCalls(2)

		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)

		createCalls := mockClient.GetCreateCalls()
		assert.Len(t, createCalls, 3, "Should have retried 3 times (2 failures + 1 success)")
		assert.Equal(t, 1, mockClient.DatabaseCount(), "Should have created one database")
	})

	t.Run("fails after max retries", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.SetExistsOnFirstNCalls(10)

		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create database after 5 attempts")
	})

	t.Run("skips when database ping fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.SetPingError(errors.New("connection refused"))

		step := NewDbCreateStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "testapp",
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err, "Should not error when ping fails, just skip")
		assert.Empty(t, ctx.GetDbSuffix(), "DbSuffix should not be set when skipped")
	})
}

func TestHandleMigrationPrompt(t *testing.T) {
	t.Run("passes full database name to prompter when suffix is set", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockClient := NewMockDatabaseClient()
		mockPrompter := &mockDbPrompter{confirmResult: true}

		step := NewDbCreateStepWithPrompter(config.StepConfig{}, MockClientFactory(mockClient), mockPrompter)
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}
		ctx.SetDbSuffix("swift_runner")

		err := step.handleMigrationPrompt(ctx, types.StepOptions{PromptMode: types.PromptMode{Interactive: true}})
		assert.NoError(t, err)
		assert.Equal(t, "myapp_swift_runner", mockPrompter.confirmMigrationsCall,
			"Should pass the full database name to the prompter")
	})

	t.Run("passes full database name with custom prefix to prompter", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockClient := NewMockDatabaseClient()
		mockPrompter := &mockDbPrompter{confirmResult: true}

		step := NewDbCreateStepWithPrompter(config.StepConfig{
			Args: []string{"--prefix", "quotes"},
		}, MockClientFactory(mockClient), mockPrompter)
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}
		ctx.SetDbSuffix("swift_runner")

		err := step.handleMigrationPrompt(ctx, types.StepOptions{PromptMode: types.PromptMode{Interactive: true}})
		assert.NoError(t, err)
		assert.Equal(t, "quotes_swift_runner", mockPrompter.confirmMigrationsCall,
			"Should use the custom prefix, not the site name")
	})

	t.Run("passes empty database name to prompter when no suffix is set", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockClient := NewMockDatabaseClient()
		mockPrompter := &mockDbPrompter{confirmResult: true}

		step := NewDbCreateStepWithPrompter(config.StepConfig{}, MockClientFactory(mockClient), mockPrompter)
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}
		// No suffix set

		err := step.handleMigrationPrompt(ctx, types.StepOptions{PromptMode: types.PromptMode{Interactive: true}})
		assert.NoError(t, err)
		assert.Equal(t, "", mockPrompter.confirmMigrationsCall,
			"Should pass empty string when no suffix is available")
	})

	t.Run("sets skip_migrations when user declines", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockClient := NewMockDatabaseClient()
		mockPrompter := &mockDbPrompter{confirmResult: false}

		step := NewDbCreateStepWithPrompter(config.StepConfig{}, MockClientFactory(mockClient), mockPrompter)
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}
		ctx.SetDbSuffix("swift_runner")

		err := step.handleMigrationPrompt(ctx, types.StepOptions{PromptMode: types.PromptMode{Interactive: true}})
		assert.NoError(t, err)
		assert.Equal(t, "true", ctx.GetVar("skip_migrations"),
			"Should set skip_migrations when user declines")
	})

	t.Run("does not prompt when PromptMode does not allow", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockClient := NewMockDatabaseClient()
		mockPrompter := &mockDbPrompter{confirmResult: true}

		step := NewDbCreateStepWithPrompter(config.StepConfig{}, MockClientFactory(mockClient), mockPrompter)
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
			SiteName:     "myapp",
		}
		ctx.SetDbSuffix("swift_runner")

		// Default StepOptions has PromptMode that does not allow prompts
		err := step.handleMigrationPrompt(ctx, types.StepOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "", mockPrompter.confirmMigrationsCall,
			"Should not call prompter when prompts are not allowed")
	})
}

func TestDbDestroyStep(t *testing.T) {
	t.Run("name returns db.destroy", func(t *testing.T) {
		step := NewDbDestroyStep(config.StepConfig{})
		assert.Equal(t, "db.destroy", step.Name())
	})

	t.Run("condition always returns true - controlled by preset", func(t *testing.T) {
		step := NewDbDestroyStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: t.TempDir(),
		}
		assert.True(t, step.Condition(ctx))
	})

	t.Run("returns nil when no DbSuffix in context or local state", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err, "Should return nil when no DbSuffix found")
	})

	t.Run("reads DbSuffix from local state", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		if err := config.WriteLocalState(tmpDir, config.LocalState{DbSuffix: "swift_runner"}); err != nil {
			t.Fatalf("writing local state: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.AddDatabase("myapp_swift_runner")

		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.Equal(t, "swift_runner", ctx.GetDbSuffix(), "DbSuffix should be read from local state")

		listCalls := mockClient.listCalls
		assert.Len(t, listCalls, 1)
		assert.Equal(t, "%_swift_runner", listCalls[0])
	})

	t.Run("drops databases matching suffix", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.AddDatabase("app1_test_suffix")
		mockClient.AddDatabase("app2_test_suffix")

		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}
		ctx.SetDbSuffix("test_suffix")

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)

		dropCalls := mockClient.GetDropCalls()
		assert.Len(t, dropCalls, 2, "Should have dropped 2 databases")
		assert.Equal(t, 0, mockClient.DatabaseCount(), "All databases should be dropped")
	})

	t.Run("auto-detects mysql engine from DB_CONNECTION env", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		if err := config.WriteLocalState(tmpDir, config.LocalState{DbSuffix: "test_suffix"}); err != nil {
			t.Fatalf("writing local state: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
	})

	t.Run("auto-detects pgsql engine from DB_CONNECTION env", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=pgsql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		if err := config.WriteLocalState(tmpDir, config.LocalState{DbSuffix: "test_suffix"}); err != nil {
			t.Fatalf("writing local state: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
	})

	t.Run("uses explicit type config over env detection", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		if err := config.WriteLocalState(tmpDir, config.LocalState{DbSuffix: "test_suffix"}); err != nil {
			t.Fatalf("writing local state: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
	})

	t.Run("uses DbSuffix from context if set", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.AddDatabase("app_context_suffix")

		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}
		ctx.SetDbSuffix("context_suffix")

		if err := config.WriteLocalState(tmpDir, config.LocalState{DbSuffix: "config_suffix"}); err != nil {
			t.Fatalf("writing local state: %v", err)
		}

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
		assert.Equal(t, "context_suffix", ctx.GetDbSuffix(), "Should use DbSuffix from context, not local state")

		listCalls := mockClient.listCalls
		assert.Len(t, listCalls, 1)
		assert.Equal(t, "%_context_suffix", listCalls[0], "Should search with context suffix")
	})

	t.Run("skips when database ping fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.SetPingError(errors.New("connection refused"))

		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}
		ctx.SetDbSuffix("test_suffix")

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err, "Should not error when ping fails, just skip")
	})

	t.Run("skips sqlite engine", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=sqlite\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		step := NewDbDestroyStep(config.StepConfig{})
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}
		ctx.SetDbSuffix("test_suffix")

		err := step.Run(ctx, types.StepOptions{Verbose: false})
		assert.NoError(t, err)
	})

	t.Run("dry run does not drop databases", func(t *testing.T) {
		tmpDir := t.TempDir()

		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("DB_CONNECTION=mysql\n"), 0644); err != nil {
			t.Fatalf("writing env file: %v", err)
		}

		mockClient := NewMockDatabaseClient()
		mockClient.AddDatabase("app_test_suffix")

		step := NewDbDestroyStepWithFactory(config.StepConfig{}, MockClientFactory(mockClient))
		ctx := &types.ScaffoldContext{
			WorktreePath: tmpDir,
		}
		ctx.SetDbSuffix("test_suffix")

		err := step.Run(ctx, types.StepOptions{Verbose: false, DryRun: true})
		assert.NoError(t, err)

		dropCalls := mockClient.GetDropCalls()
		assert.Len(t, dropCalls, 0, "Should not drop databases in dry run")
		assert.Equal(t, 1, mockClient.DatabaseCount(), "Database should still exist")
	})
}

func TestIsDatabaseExistsError(t *testing.T) {
	t.Run("returns true for DatabaseExistsError", func(t *testing.T) {
		err := &DatabaseExistsError{Name: "test_db"}
		assert.True(t, IsDatabaseExistsError(err))
	})

	t.Run("returns true for error containing 'already exists'", func(t *testing.T) {
		err := errors.New("database 'test_db' already exists")
		assert.True(t, IsDatabaseExistsError(err))
	})

	t.Run("returns true for error containing '1007'", func(t *testing.T) {
		err := errors.New("Error 1007: Can't create database")
		assert.True(t, IsDatabaseExistsError(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsDatabaseExistsError(nil))
	})

	t.Run("returns false for unrelated error", func(t *testing.T) {
		err := errors.New("some other error")
		assert.False(t, IsDatabaseExistsError(err))
	})
}

func createTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	barePath := filepath.Join(tmpDir, ".bare")

	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	readmePath := filepath.Join(repoDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("test"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "clone", "--bare", repoDir, barePath)
	require.NoError(t, cmd.Run())

	return barePath
}

func TestDiscoverWorktreeDatabases(t *testing.T) {
	t.Run("returns nil when barePath is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		results, err := discoverWorktreeDatabases("", tmpDir)
		assert.NoError(t, err)
		assert.Nil(t, results)
	})

	t.Run("returns empty list when no other worktrees exist", func(t *testing.T) {
		barePath := createTestRepo(t)
		projectDir := filepath.Dir(barePath)
		mainPath := filepath.Join(projectDir, "main")
		require.NoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))
		require.NoError(t, config.WriteLocalState(mainPath, config.LocalState{DbSuffix: "main_suffix"}))

		results, err := discoverWorktreeDatabases(barePath, mainPath)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("excludes current worktree from results", func(t *testing.T) {
		barePath := createTestRepo(t)
		projectDir := filepath.Dir(barePath)
		mainPath := filepath.Join(projectDir, "main")
		featurePath := filepath.Join(projectDir, "feature")

		require.NoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))
		require.NoError(t, git.CreateWorktree(barePath, featurePath, "feature", "main"))
		require.NoError(t, config.WriteLocalState(mainPath, config.LocalState{DbSuffix: "main_suffix"}))
		require.NoError(t, config.WriteLocalState(featurePath, config.LocalState{DbSuffix: "feature_suffix"}))

		results, err := discoverWorktreeDatabases(barePath, mainPath)
		assert.NoError(t, err)
		if assert.Len(t, results, 1) {
			assert.Equal(t, "feature", results[0].Branch)
			assert.Equal(t, "feature_suffix", results[0].DbSuffix)
		}
	})

	t.Run("only includes worktrees with DbSuffix", func(t *testing.T) {
		barePath := createTestRepo(t)
		projectDir := filepath.Dir(barePath)
		mainPath := filepath.Join(projectDir, "main")
		alphaPath := filepath.Join(projectDir, "alpha")
		featurePath := filepath.Join(projectDir, "feature")

		require.NoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))
		require.NoError(t, git.CreateWorktree(barePath, alphaPath, "alpha", "main"))
		require.NoError(t, git.CreateWorktree(barePath, featurePath, "feature", "main"))
		require.NoError(t, config.WriteLocalState(featurePath, config.LocalState{DbSuffix: "feature_suffix"}))

		results, err := discoverWorktreeDatabases(barePath, mainPath)
		assert.NoError(t, err)
		if assert.Len(t, results, 1) {
			assert.Equal(t, "feature", results[0].Branch)
			assert.Equal(t, "feature_suffix", results[0].DbSuffix)
		}
	})

	t.Run("sorts results by branch name", func(t *testing.T) {
		barePath := createTestRepo(t)
		projectDir := filepath.Dir(barePath)
		mainPath := filepath.Join(projectDir, "main")
		zuluPath := filepath.Join(projectDir, "zulu")
		alphaPath := filepath.Join(projectDir, "alpha")

		require.NoError(t, git.CreateWorktree(barePath, mainPath, "main", ""))
		require.NoError(t, git.CreateWorktree(barePath, zuluPath, "zulu", "main"))
		require.NoError(t, git.CreateWorktree(barePath, alphaPath, "alpha", "main"))
		require.NoError(t, config.WriteLocalState(zuluPath, config.LocalState{DbSuffix: "zulu_suffix"}))
		require.NoError(t, config.WriteLocalState(alphaPath, config.LocalState{DbSuffix: "alpha_suffix"}))

		results, err := discoverWorktreeDatabases(barePath, mainPath)
		assert.NoError(t, err)
		if assert.Len(t, results, 2) {
			assert.Equal(t, "alpha", results[0].Branch)
			assert.Equal(t, "zulu", results[1].Branch)
		}
	})
}
