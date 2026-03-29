package busd

import "context"

// NoopBusClient satisfies plugin.BusClient and silently discards all operations.
// Use it when the bus is unavailable so BusAwarePlugins don't need to handle nil.
type NoopBusClient struct{}

func (NoopBusClient) Publish(_ context.Context, _ string, _ []byte) error { return nil }
