package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Value is the base interface for all evaluator values.
type Value interface {
	String() string
}

// StringVal is a string value.
type StringVal string

func (s StringVal) String() string { return string(s) }

// ListVal is an ordered list of values.
type ListVal []Value

func (l ListVal) String() string {
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = v.String()
	}
	return strings.Join(parts, "\n")
}

// MapVal is an ordered associative map (keyword → Value).
type MapVal struct {
	Keys   []string
	Values []Value
}

func (m *MapVal) String() string {
	parts := make([]string, len(m.Keys))
	for i, k := range m.Keys {
		parts[i] = ":" + k + " " + m.Values[i].String()
	}
	return "{" + strings.Join(parts, " ") + "}"
}

// Get looks up a key in the map. Returns (value, true) or (nil, false).
func (m *MapVal) Get(key string) (Value, bool) {
	for i, k := range m.Keys {
		if k == key {
			return m.Values[i], true
		}
	}
	return nil, false
}

// NilVal represents the absence of a value.
type NilVal struct{}

func (NilVal) String() string { return "" }

// BoolVal is a boolean value.
type BoolVal bool

func (b BoolVal) String() string {
	if b {
		return "true"
	}
	return "false"
}

// FnVal is a user-defined closure.
type FnVal struct {
	Params []string
	Body   []*sexpr.Node
	Env    *Env
}

func (f *FnVal) String() string {
	return fmt.Sprintf("<fn(%s)>", strings.Join(f.Params, " "))
}

// BuiltinFunc is the signature for all builtin functions.
type BuiltinFunc func(ev *Evaluator, env *Env, args []*sexpr.Node) (Value, error)

// BuiltinVal wraps a builtin function as a Value.
type BuiltinVal struct {
	Name string
	Fn   BuiltinFunc
}

func (b *BuiltinVal) String() string {
	return fmt.Sprintf("<builtin:%s>", b.Name)
}

// isTruthy determines the truthiness of a value.
func isTruthy(v Value) bool {
	switch val := v.(type) {
	case NilVal:
		return false
	case *NilVal:
		return false
	case BoolVal:
		return bool(val)
	case StringVal:
		return string(val) != ""
	case ListVal:
		return len(val) > 0
	case *MapVal:
		return len(val.Keys) > 0
	default:
		return true
	}
}
