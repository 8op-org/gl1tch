package pipeline

import (
	"fmt"
	"strings"
)

// buildDAG constructs an adjacency list mapping each step ID to the IDs of
// steps that depend on it (i.e. reverse edges). It also validates that all
// IDs referenced in Needs exist, and detects cycles via DFS.
//
// Returns:
//
//	dependents map[stepID][]stepID — downstream steps that become ready when stepID completes
//	error      — non-nil if a cycle is detected or a Needs reference is invalid
func buildDAG(steps []Step) (map[string][]string, error) {
	// Build set of known IDs.
	known := make(map[string]struct{}, len(steps))
	for _, s := range steps {
		known[s.ID] = struct{}{}
	}

	// Build reverse adjacency list: for each step, record which steps depend on it.
	// Also build forward list (step → its needs) for cycle detection.
	dependents := make(map[string][]string, len(steps))
	forward := make(map[string][]string, len(steps))
	for _, s := range steps {
		if _, ok := dependents[s.ID]; !ok {
			dependents[s.ID] = nil
		}
		for _, need := range s.Needs {
			if _, ok := known[need]; !ok {
				return nil, fmt.Errorf("step %q needs unknown step %q", s.ID, need)
			}
			dependents[need] = append(dependents[need], s.ID)
			forward[s.ID] = append(forward[s.ID], need)
		}
	}

	if err := detectCycle(forward); err != nil {
		return nil, err
	}
	return dependents, nil
}

// detectCycle runs a DFS on the forward adjacency list (step → its needs).
// It uses grey/black colouring: grey = currently on stack, black = fully visited.
// Returns an error describing the cycle path if one is found.
func detectCycle(graph map[string][]string) error {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	colour := make(map[string]int, len(graph))
	path := make([]string, 0, len(graph))

	var dfs func(node string) error
	dfs = func(node string) error {
		colour[node] = grey
		path = append(path, node)

		for _, dep := range graph[node] {
			switch colour[dep] {
			case grey:
				// Found back-edge → cycle.
				// Reconstruct cycle segment.
				idx := -1
				for i, n := range path {
					if n == dep {
						idx = i
						break
					}
				}
				cycle := append(path[idx:], dep)
				return fmt.Errorf("DAG cycle detected: %s", strings.Join(cycle, " → "))
			case white:
				if err := dfs(dep); err != nil {
					return err
				}
			}
		}

		path = path[:len(path)-1]
		colour[node] = black
		return nil
	}

	// Visit every node in the graph (handles disconnected sub-graphs).
	for node := range graph {
		if colour[node] == white {
			if err := dfs(node); err != nil {
				return err
			}
		}
	}
	return nil
}
