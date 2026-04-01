package plugin

import (
	"context"
	"io"
)

// BusClient is the interface a BusAwarePlugin uses to interact with the event bus.
// It is satisfied by *busd.ConnectedClient or NoopBusClient.
// Defined here (not in busd) to avoid import cycles — the plugin package must
// not import busd.
type BusClient interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// BusAwarePlugin is an optional interface a Tier 1 plugin may implement to
// receive a BusClient on startup. The base Plugin interface is unchanged —
// existing plugins need no modification.
type BusAwarePlugin interface {
	Plugin
	SetBusClient(c BusClient)
}

// Plugin is the universal interface all glitch plugins implement,
// regardless of whether they are native go-plugins (Tier 1) or CLI wrappers (Tier 2).
//
// Execute parameters:
//   - input: the primary data payload / stdin for this invocation (prompt text or raw content)
//   - vars: string metadata passed as environment/template variables — not structured data;
//     for typed structured data use ExecuteRequest.Args (see proto/gl1tch/v1/plugin.proto)
type Plugin interface {
	Name() string
	Description() string
	Capabilities() []Capability
	Execute(ctx context.Context, input string, vars map[string]string, w io.Writer) error
	Close() error
}

// Capability describes one thing a plugin can do.
type Capability struct {
	Name         string
	InputSchema  string
	OutputSchema string
}

// StubPlugin is a test double that satisfies the Plugin interface.
type StubPlugin struct {
	PluginName string
	PluginDesc string
	PluginCaps []Capability
	ExecuteFn  func(ctx context.Context, input string, vars map[string]string, w io.Writer) error
}

func (s *StubPlugin) Name() string              { return s.PluginName }
func (s *StubPlugin) Description() string        { return s.PluginDesc }
func (s *StubPlugin) Capabilities() []Capability { return s.PluginCaps }
func (s *StubPlugin) Close() error               { return nil }
func (s *StubPlugin) Execute(ctx context.Context, input string, vars map[string]string, w io.Writer) error {
	if s.ExecuteFn != nil {
		return s.ExecuteFn(ctx, input, vars, w)
	}
	return nil
}
