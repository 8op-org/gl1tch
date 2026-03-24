package plugin

import (
	"context"
	"io"
	"os/exec"
	"strings"
)

// CliAdapter wraps an arbitrary CLI tool as a Tier 2 Plugin.
// Input is written to the subprocess stdin; stdout/stderr is streamed to the writer.
// args are fixed command-line arguments prepended to every Execute call.
type CliAdapter struct {
	name string
	desc string
	cmd  string
	args []string
}

// NewCliAdapter creates a Tier 2 plugin that wraps cmd.
func NewCliAdapter(name, description, cmd string, args ...string) *CliAdapter {
	return &CliAdapter{name: name, desc: description, cmd: cmd, args: args}
}

func (c *CliAdapter) Name() string              { return c.name }
func (c *CliAdapter) Description() string        { return c.desc }
func (c *CliAdapter) Capabilities() []Capability { return nil }
func (c *CliAdapter) Close() error               { return nil }

// Execute spawns the subprocess, writes input to stdin, and streams stdout to w.
func (c *CliAdapter) Execute(ctx context.Context, input string, _ map[string]string, w io.Writer) error {
	cmd := exec.CommandContext(ctx, c.cmd, c.args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}
