package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
)

// dbExecutor implements StepExecutor for the built-in "db" step type.
// It executes SQL against the result store attached to the ExecutionContext.
//
// Step YAML:
//
//	- id: my_query
//	  type: db
//	  args:
//	    op: query              # "query" | "exec"
//	    sql: "SELECT * FROM runs WHERE kind = '{{param.kind}}' LIMIT 5"
//
// Results:
//   - op=query: ec.Set("<step-id>.out", JSON array); ec.Set("<step-id>.rows", count)
//   - op=exec:  ec.Set("<step-id>.rows_affected", n)
type dbExecutor struct {
	ec *ExecutionContext
}

func (d *dbExecutor) Init(_ context.Context) error { return nil }

// Execute runs the SQL op against the store database.
func (d *dbExecutor) Execute(_ context.Context, args map[string]any) (map[string]any, error) {
	s := d.ec.DB()
	if s == nil {
		return nil, fmt.Errorf("db step: no result store configured (pass WithRunStore to pipeline.Run)")
	}

	op, _ := args["op"].(string)
	query, _ := args["sql"].(string)
	if query == "" {
		return nil, fmt.Errorf("db step: missing required 'sql' arg")
	}

	// SQL interpolation is already handled by interpolateArgs before Execute is called,
	// but apply once more with the current snapshot for any late-binding placeholders.
	query = Interpolate(query, d.ec.Snapshot())

	db := s.DB()

	switch op {
	case "query", "":
		rows, err := db.Query(query) //nolint:rowserrcheck
		if err != nil {
			return nil, fmt.Errorf("db step query: %w", err)
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("db step columns: %w", err)
		}

		var results []map[string]any
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return nil, fmt.Errorf("db step scan: %w", err)
			}
			row := make(map[string]any, len(cols))
			for i, col := range cols {
				row[col] = vals[i]
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("db step rows: %w", err)
		}

		jsonBytes, err := json.Marshal(results)
		if err != nil {
			return nil, fmt.Errorf("db step marshal: %w", err)
		}

		out := string(jsonBytes)
		d.ec.Set(d.stepID(args)+".out", out)
		d.ec.Set(d.stepID(args)+".rows", len(results))
		return map[string]any{"value": out, "rows": len(results)}, nil

	case "exec":
		res, err := db.Exec(query)
		if err != nil {
			return nil, fmt.Errorf("db step exec: %w", err)
		}
		n, _ := res.RowsAffected()
		d.ec.Set(d.stepID(args)+".rows_affected", n)
		return map[string]any{"rows_affected": n}, nil

	default:
		return nil, fmt.Errorf("db step: unknown op %q (must be 'query' or 'exec')", op)
	}
}

func (d *dbExecutor) Cleanup(_ context.Context) error { return nil }

// stepID extracts the step id from args (injected by the runner as "_step_id").
func (d *dbExecutor) stepID(args map[string]any) string {
	if v, ok := args["_step_id"].(string); ok {
		return v
	}
	return "db"
}

