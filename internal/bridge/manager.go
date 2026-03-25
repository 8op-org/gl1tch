package bridge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/adam-stokes/orcai/internal/providers"
	bridgepb "github.com/adam-stokes/orcai/proto/bridgepb"
)

// Manager spawns and manages provider adapter subprocesses.
//
// NOTE: The gRPC bridge adapter model is being replaced by the plugin system v2
// direct-launch approach. Start() is a no-op stub until the new widget-based
// session launcher is wired in. Available() still reflects detected providers.
type Manager struct {
	cwd       string
	socketDir string
	adapters  []*adapterEntry
	cancel    context.CancelFunc
	reg       *providers.Registry
}

type adapterEntry struct {
	name     string
	proc     *exec.Cmd
	conn     *grpc.ClientConn
	client   bridgepb.ProviderBridgeClient
	descResp *bridgepb.DescribeResponse
}

// New creates a manager for the given working directory and config dir.
// configDir is used to load user-installed provider profiles.
func New(cwd, configDir string) *Manager {
	reg, _ := providers.NewRegistry(filepath.Join(configDir, "providers"))
	return &Manager{cwd: cwd, reg: reg}
}

// Start is a no-op stub. The gRPC bridge subprocess model is superseded by the
// plugin system v2 direct session launcher (see internal/widgets). This method
// exists for API compatibility while the migration completes.
func (m *Manager) Start(_ context.Context) error {
	return nil
}

// Stop is a no-op while Start is a stub.
func (m *Manager) Stop(_ context.Context) {
	if m.cancel != nil {
		m.cancel()
	}
	if m.socketDir != "" {
		os.RemoveAll(m.socketDir)
	}
}

// Client returns the gRPC client for the named adapter.
// Returns nil while the bridge is in stub mode.
func (m *Manager) Client(name string) bridgepb.ProviderBridgeClient {
	for _, a := range m.adapters {
		if a.name == name {
			return a.client
		}
	}
	return nil
}

// Capabilities returns all capabilities from all running adapters.
func (m *Manager) Capabilities() []*bridgepb.Capability {
	var all []*bridgepb.Capability
	for _, a := range m.adapters {
		all = append(all, a.descResp.Capabilities...)
	}
	return all
}

// AvailableProviders returns provider profiles whose binaries are in PATH.
func (m *Manager) AvailableProviders() []providers.Profile {
	if m.reg == nil {
		return nil
	}
	return m.reg.Available()
}

