package executor

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	executorTracer = otel.Tracer("gl1tch/executor")
	executorCalls  metric.Int64Counter
)

func init() {
	meter := otel.Meter("gl1tch/executor")
	executorCalls, _ = meter.Int64Counter("gl1tch.executor.calls",
		metric.WithDescription("Total executor dispatch calls"),
	)
}

// actionExecutor wraps an Executor and injects _action into the vars map on Execute.
// It is returned by Manager.Get for category.action lookups.
type actionExecutor struct {
	Executor
	action string
}

func (ap *actionExecutor) Name() string { return ap.Executor.Name() + "." + ap.action }

func (ap *actionExecutor) Execute(ctx context.Context, input string, vars map[string]string, w io.Writer) error {
	merged := make(map[string]string, len(vars)+1)
	for k, v := range vars {
		merged[k] = v
	}
	merged["_action"] = ap.action
	return ap.Executor.Execute(ctx, input, merged, w)
}

// TODO(translations): When an executor execution context interface is added,
// inject a translations.Provider via a Translations() method so executors can
// use the same operator-configured string overrides as the TUI. Wire it from
// translations.GlobalProvider() in the executor host/executor setup.

// Manager holds the registry of all active executors.
type Manager struct {
	mu            sync.RWMutex
	executors     map[string]Executor
	categoryIndex map[string]Executor // keyed by category prefix (e.g. "providers.claude")
	busClient     BusClient
}

// NewManager returns an empty executor manager.
func NewManager() *Manager {
	return &Manager{
		executors:     make(map[string]Executor),
		categoryIndex: make(map[string]Executor),
	}
}

// Register adds an executor. Returns an error if an executor with the same name is already registered.
// If a BusClient has already been set on the Manager, it is injected into any
// BusAwareExecutor immediately on registration.
func (m *Manager) Register(e Executor) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.executors[e.Name()]; exists {
		return fmt.Errorf("executor %q already registered", e.Name())
	}
	m.executors[e.Name()] = e
	if m.busClient != nil {
		if ba, ok := e.(BusAwareExecutor); ok {
			ba.SetBusClient(m.busClient)
		}
	}
	return nil
}

// SetBusClient sets the BusClient on the Manager and injects it into all
// already-registered BusAwareExecutors. Executors registered after this call also
// receive the client automatically via Register.
func (m *Manager) SetBusClient(c BusClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.busClient = c
	for _, e := range m.executors {
		if ba, ok := e.(BusAwareExecutor); ok {
			ba.SetBusClient(c)
		}
	}
}

// RegisterCategory registers an executor under a category key and tracks the valid action.
// This enables hierarchical "category.action" lookups via Get.
// category is the prefix (e.g. "providers.claude"), action is the specific capability.
func (m *Manager) RegisterCategory(category, action string, e Executor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Store under the category name. If multiple actions are registered, the last one wins —
	// the actual dispatch is done by injecting _action into the vars map.
	m.categoryIndex[category] = e
	// Also register under the full "category.action" name in the flat executors map for direct lookup.
	fullName := category + "." + action
	if _, exists := m.executors[fullName]; !exists {
		m.executors[fullName] = &actionExecutor{Executor: e, action: action}
	}
}

// Get returns an executor by name.
// Resolution order:
//  1. Direct lookup in the flat executors map.
//  2. Category lookup: split name on the last dot, look up category in categoryIndex,
//     return a wrapper that injects _action into the vars map.
func (m *Manager) Get(name string) (Executor, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 1. Direct lookup.
	if e, ok := m.executors[name]; ok {
		return e, true
	}

	// 2. Category lookup: split on last dot.
	idx := strings.LastIndex(name, ".")
	if idx <= 0 {
		return nil, false
	}
	category := name[:idx]
	action := name[idx+1:]
	if e, ok := m.categoryIndex[category]; ok {
		return &actionExecutor{Executor: e, action: action}, true
	}

	return nil, false
}

// LoadWrappersFromDir scans dir for sidecar YAML files and registers all resulting executors.
// Per-file errors (parse failures and duplicate names) are returned; they do not prevent other
// wrappers from being registered. If a sidecar has a category field, the adapter is also
// registered under that category via RegisterCategory.
func (m *Manager) LoadWrappersFromDir(dir string) []error {
	executors, errs := LoadWrappers(dir)
	for _, e := range executors {
		if err := m.Register(e); err != nil {
			errs = append(errs, err)
			continue
		}
		// If the adapter has a category, register it in the category index.
		if ca, ok := e.(*CliAdapter); ok && ca.Category() != "" {
			m.RegisterCategory(ca.Category(), ca.Name(), e)
		}
	}
	return errs
}

// Execute looks up executor by name and runs it, recording a trace span and
// incrementing the gl1tch.executor.calls counter.
func (m *Manager) Execute(ctx context.Context, name string, input string, vars map[string]string, w io.Writer) error {
	e, ok := m.Get(name)
	if !ok {
		return fmt.Errorf("executor %q not found", name)
	}

	category := name
	if idx := strings.LastIndex(name, "."); idx > 0 {
		category = name[:idx]
	}

	ctx, span := executorTracer.Start(ctx, "executor.execute",
		trace.WithAttributes(
			attribute.String("executor.name", name),
			attribute.String("executor.category", category),
		),
	)
	defer span.End()

	executorCalls.Add(ctx, 1, metric.WithAttributes(attribute.String("executor", name)))

	if err := e.Execute(ctx, input, vars, w); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// List returns all registered executors in no guaranteed order.
func (m *Manager) List() []Executor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Executor, 0, len(m.executors))
	for _, e := range m.executors {
		out = append(out, e)
	}
	return out
}
