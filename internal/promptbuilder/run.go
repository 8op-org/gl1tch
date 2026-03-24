package promptbuilder

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adam-stokes/orcai/internal/pipeline"
	"github.com/adam-stokes/orcai/internal/plugin"
)

// Run launches the prompt builder as a standalone BubbleTea program.
func Run() {
	mgr := plugin.NewManager()
	for _, name := range []string{"claude", "gemini", "openspec", "openclaw"} {
		mgr.Register(plugin.NewCliAdapter(name, name+" CLI adapter", name))
	}

	m := New(mgr)
	m.SetName("new-pipeline")
	m.AddStep(pipeline.Step{ID: "input", Type: "input", Prompt: "Enter your prompt:"})
	m.AddStep(pipeline.Step{ID: "step1", Plugin: "claude", Model: "claude-sonnet-4-6"})
	m.AddStep(pipeline.Step{ID: "output", Type: "output"})

	bubble := NewBubble(m)
	p := tea.NewProgram(bubble, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("prompt builder error: %v\n", err)
	}
}
