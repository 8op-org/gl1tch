package pipeline

// Env is a lexical scope with parent chain for variable lookup.
type Env struct {
	bindings map[string]Value
	parent   *Env
}

// NewEnv creates a new environment with the given parent scope.
func NewEnv(parent *Env) *Env {
	return &Env{
		bindings: make(map[string]Value),
		parent:   parent,
	}
}

// Get looks up a binding by name, walking the parent chain.
func (e *Env) Get(name string) (Value, bool) {
	if v, ok := e.bindings[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set binds a name to a value in this environment.
func (e *Env) Set(name string, val Value) {
	e.bindings[name] = val
}
