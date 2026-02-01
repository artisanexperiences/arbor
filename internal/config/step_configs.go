package config

import (
	"fmt"
)

// StepValidator is an interface for step-specific configuration validation.
// Each step type can implement this interface to validate its required fields.
type StepValidator interface {
	Validate() error
}

// BinaryStepConfig represents configuration for binary execution steps (php, npm, etc.)
type BinaryStepConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Args      []string               `mapstructure:"args"`
	StoreAs   string                 `mapstructure:"store_as"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that the binary step config is valid.
// Binary steps only require a name; all other fields are optional.
func (c BinaryStepConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("binary step: 'name' is required")
	}
	return nil
}

// FileCopyConfig represents configuration for file.copy step
type FileCopyConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	From      string                 `mapstructure:"from"`
	To        string                 `mapstructure:"to"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that required fields are present for file.copy step
func (c FileCopyConfig) Validate() error {
	if c.From == "" {
		return fmt.Errorf("file.copy: 'from' is required")
	}
	if c.To == "" {
		return fmt.Errorf("file.copy: 'to' is required")
	}
	return nil
}

// BashRunConfig represents configuration for bash.run step
type BashRunConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Command   string                 `mapstructure:"command"`
	StoreAs   string                 `mapstructure:"store_as"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that required fields are present for bash.run step
func (c BashRunConfig) Validate() error {
	if c.Command == "" {
		return fmt.Errorf("bash.run: 'command' is required")
	}
	return nil
}

// CommandRunConfig represents configuration for command.run step
type CommandRunConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Command   string                 `mapstructure:"command"`
	StoreAs   string                 `mapstructure:"store_as"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that required fields are present for command.run step
func (c CommandRunConfig) Validate() error {
	if c.Command == "" {
		return fmt.Errorf("command.run: 'command' is required")
	}
	return nil
}

// EnvReadConfig represents configuration for env.read step
type EnvReadConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Key       string                 `mapstructure:"key"`
	StoreAs   string                 `mapstructure:"store_as"`
	File      string                 `mapstructure:"file"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that required fields are present for env.read step
func (c EnvReadConfig) Validate() error {
	if c.Key == "" {
		return fmt.Errorf("env.read: 'key' is required")
	}
	return nil
}

// EnvWriteConfig represents configuration for env.write step
type EnvWriteConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Key       string                 `mapstructure:"key"`
	Value     string                 `mapstructure:"value"`
	File      string                 `mapstructure:"file"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that required fields are present for env.write step
func (c EnvWriteConfig) Validate() error {
	if c.Key == "" {
		return fmt.Errorf("env.write: 'key' is required")
	}
	return nil
}

// DbCreateConfig represents configuration for db.create step
type DbCreateConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Args      []string               `mapstructure:"args"`
	Type      string                 `mapstructure:"type"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that the db.create step config is valid.
// All fields are optional for db.create.
func (c DbCreateConfig) Validate() error {
	return nil
}

// DbDestroyConfig represents configuration for db.destroy step
type DbDestroyConfig struct {
	Name      string                 `mapstructure:"name"`
	Enabled   *bool                  `mapstructure:"enabled"`
	Args      []string               `mapstructure:"args"`
	Type      string                 `mapstructure:"type"`
	Condition map[string]interface{} `mapstructure:"condition"`
}

// Validate checks that the db.destroy step config is valid.
// All fields are optional for db.destroy.
func (c DbDestroyConfig) Validate() error {
	return nil
}

// ValidateStepConfig validates a StepConfig based on its step type.
// The stepName parameter is used to determine the step type for validation.
// This is the main entry point for step validation.
func ValidateStepConfig(stepName string, cfg StepConfig) error {
	switch stepName {
	case "file.copy":
		return FileCopyConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			From:      cfg.From,
			To:        cfg.To,
			Condition: cfg.Condition,
		}.Validate()
	case "bash.run":
		return BashRunConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Command:   cfg.Command,
			StoreAs:   cfg.StoreAs,
			Condition: cfg.Condition,
		}.Validate()
	case "command.run":
		return CommandRunConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Command:   cfg.Command,
			StoreAs:   cfg.StoreAs,
			Condition: cfg.Condition,
		}.Validate()
	case "env.read":
		return EnvReadConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Key:       cfg.Key,
			StoreAs:   cfg.StoreAs,
			File:      cfg.File,
			Condition: cfg.Condition,
		}.Validate()
	case "env.write":
		return EnvWriteConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Key:       cfg.Key,
			Value:     cfg.Value,
			File:      cfg.File,
			Condition: cfg.Condition,
		}.Validate()
	case "db.create":
		return DbCreateConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Args:      cfg.Args,
			Type:      cfg.Type,
			Condition: cfg.Condition,
		}.Validate()
	case "db.destroy":
		return DbDestroyConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Args:      cfg.Args,
			Type:      cfg.Type,
			Condition: cfg.Condition,
		}.Validate()
	default:
		// Binary steps (php, npm, composer, etc.) and unknown steps
		return BinaryStepConfig{
			Name:      stepName,
			Enabled:   cfg.Enabled,
			Args:      cfg.Args,
			StoreAs:   cfg.StoreAs,
			Condition: cfg.Condition,
		}.Validate()
	}
}
