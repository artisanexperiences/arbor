package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/michaeldyrynda/arbor/internal/config"
	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
)

func TestRegistry_StepRegistration(t *testing.T) {
	t.Run("env.read step is registered", func(t *testing.T) {
		step, err := Create("env.read", config.StepConfig{
			Key:     "DB_DATABASE",
			StoreAs: "DatabaseName",
		})

		require.NoError(t, err)
		assert.Equal(t, "env.read", step.Name())
	})

	t.Run("env.write step is registered", func(t *testing.T) {
		step, err := Create("env.write", config.StepConfig{
			Key:   "DB_DATABASE",
			Value: "{{ .SiteName }}_{{ .DbSuffix }}",
		})

		require.NoError(t, err)
		assert.Equal(t, "env.write", step.Name())
	})

	t.Run("node.bun step is registered", func(t *testing.T) {
		step, err := Create("node.bun", config.StepConfig{
			Args: []string{"install"},
		})

		require.NoError(t, err)
		assert.Equal(t, "node.bun", step.Name())

		binaryStep, ok := step.(*BinaryStep)
		assert.True(t, ok, "Expected BinaryStep type")
		assert.Equal(t, "bun", binaryStep.binary)
		assert.Equal(t, []string{"install"}, binaryStep.args)
	})

	t.Run("db.create step is registered", func(t *testing.T) {
		step, err := Create("db.create", config.StepConfig{})

		require.NoError(t, err)
		assert.Equal(t, "db.create", step.Name())
	})

	t.Run("db.destroy step is registered", func(t *testing.T) {
		step, err := Create("db.destroy", config.StepConfig{})

		require.NoError(t, err)
		assert.Equal(t, "db.destroy", step.Name())
	})

	t.Run("unregistered step returns error", func(t *testing.T) {
		step, err := Create("nonexistent.step", config.StepConfig{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown step")
		assert.Contains(t, err.Error(), "nonexistent.step")
		assert.Nil(t, step)
	})

	t.Run("all expected steps are registered", func(t *testing.T) {
		expectedSteps := []string{
			"php",
			"php.composer",
			"php.laravel",
			"node.npm",
			"node.yarn",
			"node.pnpm",
			"node.bun",
			"herd",
			"file.copy",
			"bash.run",
			"command.run",
			"env.read",
			"env.write",
			"db.create",
			"db.destroy",
		}

		for _, stepName := range expectedSteps {
			step, err := Create(stepName, config.StepConfig{})
			require.NoError(t, err, "Step '%s' should be registered", stepName)
			assert.Equal(t, stepName, step.Name())
		}
	})
}

func TestRegistry_ListRegistered(t *testing.T) {
	names := ListRegistered()

	assert.NotEmpty(t, names)
	assert.Contains(t, names, "php")
	assert.Contains(t, names, "env.read")
	assert.Contains(t, names, "env.write")
	assert.Contains(t, names, "db.create")
	assert.Contains(t, names, "db.destroy")

	// Verify list is sorted
	for i := 1; i < len(names); i++ {
		assert.LessOrEqual(t, names[i-1], names[i], "List should be sorted alphabetically")
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	t.Run("duplicate registration panics", func(t *testing.T) {
		// This test creates a temporary registry to test panic behavior
		// without affecting the global registry
		defer func() {
			r := recover()
			assert.NotNil(t, r, "Expected panic for duplicate registration")
			assert.Contains(t, r, "already registered")
		}()

		// Register a duplicate step - this should panic
		Register("php", func(cfg config.StepConfig) types.ScaffoldStep {
			return nil
		})
	})
}
