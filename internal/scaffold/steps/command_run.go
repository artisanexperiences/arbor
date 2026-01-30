package steps

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
)

type CommandRunStep struct {
	command string
	storeAs string
}

func NewCommandRunStep(command string, storeAs string) *CommandRunStep {
	return &CommandRunStep{command: command, storeAs: storeAs}
}

func (s *CommandRunStep) Name() string {
	return "command.run"
}

func (s *CommandRunStep) Run(ctx *types.ScaffoldContext, opts types.StepOptions) error {
	cmd := exec.Command("sh", "-c", s.command)
	cmd.Dir = ctx.WorktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command.run failed: %w\n%s", err, string(output))
	}

	if s.storeAs != "" {
		ctx.SetVar(s.storeAs, strings.TrimSpace(string(output)))
		if opts.Verbose {
			fmt.Printf("  Stored output as %s\n", s.storeAs)
		}
	}

	return nil
}

func (s *CommandRunStep) Condition(ctx *types.ScaffoldContext) bool {
	return true
}
