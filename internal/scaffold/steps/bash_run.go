package steps

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/michaeldyrynda/arbor/internal/scaffold/template"
	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
)

type BashRunStep struct {
	command string
	storeAs string
}

func NewBashRunStep(command string, storeAs string) *BashRunStep {
	return &BashRunStep{command: command, storeAs: storeAs}
}

func (s *BashRunStep) Name() string {
	return "bash.run"
}

func (s *BashRunStep) Run(ctx *types.ScaffoldContext, opts types.StepOptions) error {
	command, err := template.ReplaceTemplateVars(s.command, ctx)
	if err != nil {
		return fmt.Errorf("template replacement failed: %w", err)
	}

	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = ctx.WorktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bash.run failed: %w\n%s", err, string(output))
	}

	if s.storeAs != "" {
		ctx.SetVar(s.storeAs, strings.TrimSpace(string(output)))
		if opts.Verbose {
			fmt.Printf("  Stored output as %s\n", s.storeAs)
		}
	}

	return nil
}

func (s *BashRunStep) Condition(ctx *types.ScaffoldContext) bool {
	return true
}
