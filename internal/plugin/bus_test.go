package plugin_test

import (
	"context"
	"testing"

	"github.com/adam-stokes/orcai/internal/plugin"
)

// noopBusClientForTest satisfies plugin.BusClient without importing busd,
// avoiding a test-only import cycle.
type noopBusClientForTest struct{}

func (noopBusClientForTest) Publish(_ context.Context, _ string, _ []byte) error { return nil }

// busAwareStub is a StubPlugin that also implements BusAwarePlugin.
type busAwareStub struct {
	plugin.StubPlugin
	client plugin.BusClient
}

func (b *busAwareStub) SetBusClient(c plugin.BusClient) { b.client = c }

func TestBusAwarePlugin_ReceivesClientOnSetBusClient(t *testing.T) {
	mgr := plugin.NewManager()
	p := &busAwareStub{StubPlugin: plugin.StubPlugin{PluginName: "test"}}
	if err := mgr.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	noop := noopBusClientForTest{}
	mgr.SetBusClient(noop)

	if p.client == nil {
		t.Fatal("expected client to be set after SetBusClient")
	}

	// Publish should not panic and should return nil.
	err := p.client.Publish(context.Background(), "test.topic", []byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error from noop Publish: %v", err)
	}
}

func TestBusAwarePlugin_ReceivesClientOnRegister(t *testing.T) {
	mgr := plugin.NewManager()

	// Set the bus client before registering any plugin.
	noop := noopBusClientForTest{}
	mgr.SetBusClient(noop)

	// Plugin registered after SetBusClient should also receive the client.
	p := &busAwareStub{StubPlugin: plugin.StubPlugin{PluginName: "late"}}
	if err := mgr.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if p.client == nil {
		t.Fatal("expected client to be injected during Register")
	}
}

func TestNonBusAwarePlugin_UnaffectedBySetBusClient(t *testing.T) {
	mgr := plugin.NewManager()
	plain := &plugin.StubPlugin{PluginName: "plain"}
	if err := mgr.Register(plain); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Should not panic — plain plugin doesn't implement BusAwarePlugin.
	mgr.SetBusClient(noopBusClientForTest{})

	if _, ok := mgr.Get("plain"); !ok {
		t.Error("plain plugin should still be registered")
	}
}
