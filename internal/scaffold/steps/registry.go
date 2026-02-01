package steps

import (
	"fmt"
	"sort"

	"github.com/michaeldyrynda/arbor/internal/config"
	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
)

type StepFactory func(cfg config.StepConfig) types.ScaffoldStep

var registry = make(map[string]StepFactory)

func Register(name string, factory StepFactory) {
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("step %q already registered", name))
	}
	registry[name] = factory
}

func Create(name string, cfg config.StepConfig) (types.ScaffoldStep, error) {
	if factory, ok := registry[name]; ok {
		return factory(cfg), nil
	}
	return nil, fmt.Errorf("unknown step %q (available: %v)", name, ListRegistered())
}

func ListRegistered() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type binaryDefinition struct {
	name   string
	binary string
}

var binaries = []binaryDefinition{
	{"php", "php"},
	{"php.composer", "composer"},
	{"php.laravel", "php artisan"},
	{"node.npm", "npm"},
	{"node.yarn", "yarn"},
	{"node.pnpm", "pnpm"},
	{"node.bun", "bun"},
	{"herd", "herd"},
}

func init() {
	for _, b := range binaries {
		name := b.name
		binary := b.binary
		Register(name, func(cfg config.StepConfig) types.ScaffoldStep {
			return NewBinaryStepWithCondition(name, cfg, binary)
		})
	}

	Register("file.copy", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewFileCopyStep(cfg.From, cfg.To)
	})
	Register("bash.run", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewBashRunStep(cfg.Command, cfg.StoreAs)
	})
	Register("command.run", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewCommandRunStep(cfg.Command, cfg.StoreAs)
	})
	Register("env.read", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewEnvReadStep(cfg)
	})
	Register("env.write", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewEnvWriteStep(cfg)
	})
	Register("db.create", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewDbCreateStep(cfg)
	})
	Register("db.destroy", func(cfg config.StepConfig) types.ScaffoldStep {
		return NewDbDestroyStep(cfg)
	})
}
