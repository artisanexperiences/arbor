package steps

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
)

type BinaryStep struct {
	name     string
	binary   string
	args     []string
	priority int
}

func NewBinaryStep(name, binary string, args []string, priority int) *BinaryStep {
	return &BinaryStep{
		name:     name,
		binary:   binary,
		args:     args,
		priority: priority,
	}
}

func (s *BinaryStep) Name() string {
	return s.name
}

func (s *BinaryStep) Run(ctx types.ScaffoldContext, opts types.StepOptions) error {
	allArgs := append(s.args, opts.Args...)
	allArgs = s.replaceTemplate(allArgs, ctx)
	if opts.Verbose {
		binaryParts := strings.Fields(s.binary)
		fullCmd := append(binaryParts, allArgs...)
		fmt.Printf("  Running: %s\n", strings.Join(fullCmd, " "))
	}
	cmd := exec.Command(strings.Fields(s.binary)[0], append(strings.Fields(s.binary)[1:], allArgs...)...)
	cmd.Dir = ctx.WorktreePath
	return cmd.Run()
}

func (s *BinaryStep) replaceTemplate(args []string, ctx types.ScaffoldContext) []string {
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{{ .RepoName }}", ctx.RepoName)
		arg = strings.ReplaceAll(arg, "{{ .Branch }}", ctx.Branch)
		args[i] = arg
	}
	return args
}

func (s *BinaryStep) Priority() int {
	return s.priority
}

func (s *BinaryStep) Condition(ctx types.ScaffoldContext) bool {
	binaries := strings.Fields(s.binary)
	if len(binaries) == 0 {
		return false
	}
	_, err := exec.LookPath(binaries[0])
	return err == nil
}
