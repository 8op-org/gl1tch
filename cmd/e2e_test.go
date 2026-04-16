//go:build integration

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/store"
	"github.com/8op-org/gl1tch/internal/workspace"
)

// TestE2EWorkspaceResources drives the full workspace + local-resource + pipeline
// render path: init a workspace, add a local-path resource, verify the
// workspace.glitch entry and the resources/<name> symlink, then execute a
// workflow whose step references {{.resource.notes.path}} and confirm the
// rendered output matches the symlink location.
func TestE2EWorkspaceResources(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ws := filepath.Join(home, "demo-ws")
	if err := runWorkspaceInit(ws, "demo"); err != nil {
		t.Fatalf("init: %v", err)
	}

	notesDir := filepath.Join(home, "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatalf("mkdir notes: %v", err)
	}
	if err := runWorkspaceAdd(ws, notesDir, "notes", "", ""); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Verify workspace.glitch contains the resource.
	data, err := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	if err != nil {
		t.Fatalf("read workspace.glitch: %v", err)
	}
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		t.Fatalf("parse workspace.glitch: %v", err)
	}
	if len(wsp.Resources) != 1 || wsp.Resources[0].Name != "notes" {
		t.Fatalf("resource not recorded: %+v", wsp.Resources)
	}

	// Verify the symlink is wired to the target directory.
	link, err := os.Readlink(filepath.Join(ws, "resources", "notes"))
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if link != notesDir {
		t.Fatalf("symlink target mismatch: got %q want %q", link, notesDir)
	}

	// Scaffold a workflow that references the resource path.
	flowDir := filepath.Join(ws, "workflows")
	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	flowPath := filepath.Join(flowDir, "flow.glitch")
	if err := os.WriteFile(flowPath,
		[]byte(`(workflow "flow" (step "s" (run "echo path={{.resource.notes.path}}")))`),
		0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	// Run the workflow via pipeline.Run with the resource bindings populated.
	wf, err := pipeline.ParseSexprWorkflowFromFile(flowPath)
	if err != nil {
		t.Fatalf("parse workflow: %v", err)
	}
	resources := ResourceBindings(wsp, ws)
	res, err := pipeline.Run(wf, "", "", map[string]string{}, nil, pipeline.RunOpts{
		Workspace:    "demo",
		Resources:    resources,
		WorkflowsDir: flowDir,
	})
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}
	want := filepath.Join(ws, "resources", "notes")
	if !strings.Contains(res.Steps["s"], "path="+want) {
		t.Fatalf("expected path=%q in step output, got %q", want, res.Steps["s"])
	}
}

// TestE2ECallWorkflowFromAsk verifies that the ask path threads WorkflowsDir
// into pipeline.Run so (call-workflow ...) resolves inside a workspace. This
// mirrors the code in cmd/ask.go's router-routed branch: workflowsDir is
// derived the same way, and pipeline.Run is invoked with the same option set.
// Without the WorkflowsDir thread-through, call-workflow errors with
// "call-workflow requires WorkflowsDir in RunOpts".
func TestE2ECallWorkflowFromAsk(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ws := filepath.Join(home, "demo-ws")
	if err := runWorkspaceInit(ws, "demo"); err != nil {
		t.Fatalf("init: %v", err)
	}

	flowDir := filepath.Join(ws, "workflows")
	parentPath := filepath.Join(flowDir, "parent.glitch")
	childPath := filepath.Join(flowDir, "child.glitch")
	if err := os.WriteFile(parentPath,
		[]byte(`(workflow "parent" (step "o" (call-workflow "child" :input "hello")))`),
		0o644); err != nil {
		t.Fatalf("write parent: %v", err)
	}
	if err := os.WriteFile(childPath,
		[]byte(`(workflow "child" (step "e" (run "echo got:{{.input}}")))`),
		0o644); err != nil {
		t.Fatalf("write child: %v", err)
	}

	// Simulate ask's path: resolve workspace, load workflows, pick the one
	// the router would match, run via pipeline.Run with WorkflowsDir set the
	// way ask sets it.
	prevWsPath := workspacePath
	workspacePath = ws
	t.Cleanup(func() { workspacePath = prevWsPath })

	workflowsDir := filepath.Join(ws, "workflows")
	wf, err := pipeline.ParseSexprWorkflowFromFile(parentPath)
	if err != nil {
		t.Fatalf("parse parent: %v", err)
	}
	result, err := pipeline.Run(wf, "", "", map[string]string{}, nil, pipeline.RunOpts{
		Workspace:    "demo",
		WorkflowsDir: workflowsDir,
	})
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}
	// The child's step "e" runs `echo got:hello`; its output should appear in
	// the combined result via the parent's "o" step output.
	got := result.Steps["o"]
	if !strings.Contains(got, "got:hello") {
		t.Fatalf("expected child output 'got:hello' in parent step, got %q", got)
	}
}

// TestE2ECallWorkflowParentLinkage drives parent + child workflow execution
// via call-workflow, wiring ChildRunCreator to the real store so child rows
// carry parent_run_id. Asserts the resulting run tree via store.GetRunTree.
func TestE2ECallWorkflowParentLinkage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ws := filepath.Join(home, "demo-ws")
	if err := runWorkspaceInit(ws, "demo"); err != nil {
		t.Fatalf("init: %v", err)
	}

	flowDir := filepath.Join(ws, "workflows")
	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	parentPath := filepath.Join(flowDir, "parent.glitch")
	childPath := filepath.Join(flowDir, "child.glitch")
	if err := os.WriteFile(parentPath,
		[]byte(`(workflow "parent" (step "o" (call-workflow "child" :input "hi")))`),
		0o644); err != nil {
		t.Fatalf("write parent: %v", err)
	}
	if err := os.WriteFile(childPath,
		[]byte(`(workflow "child" (step "e" (run "echo got:{{.input}}")))`),
		0o644); err != nil {
		t.Fatalf("write child: %v", err)
	}

	s, err := store.OpenForWorkspace(ws)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	// Pre-create the parent's run row — the id we pass in as ParentRunID.
	parentID, err := s.RecordRun(store.RunRecord{
		Kind:         "workflow",
		Name:         "parent",
		Input:        "hi",
		WorkflowName: "parent",
		Workspace:    "demo",
	})
	if err != nil {
		t.Fatalf("record parent: %v", err)
	}

	// The creator is invoked by call-workflow before each nested run starts;
	// it records the child row and hands the new id back to be used as the
	// child's ParentRunID — giving proper per-level parent linkage.
	creator := func(parent int64, name string) (int64, error) {
		return s.RecordRun(store.RunRecord{
			Kind:         "workflow",
			Name:         name,
			WorkflowName: name,
			ParentRunID:  parent,
			Workspace:    "demo",
		})
	}

	wf, err := pipeline.ParseSexprWorkflowFromFile(parentPath)
	if err != nil {
		t.Fatalf("parse parent: %v", err)
	}
	if _, err := pipeline.Run(wf, "hi", "", nil, nil, pipeline.RunOpts{
		Workspace:       "demo",
		WorkflowsDir:    flowDir,
		ParentRunID:     parentID,
		ChildRunCreator: creator,
	}); err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}

	// Verify the store's run tree: parent with one child whose parent_run_id
	// links back to the parent's row id.
	tree, err := s.GetRunTree(parentID)
	if err != nil {
		t.Fatalf("GetRunTree: %v", err)
	}
	if tree.ID != parentID {
		t.Fatalf("root mismatch: got %d want %d", tree.ID, parentID)
	}
	if len(tree.Children) != 1 {
		t.Fatalf("expected exactly one child, got %d: %+v", len(tree.Children), tree.Children)
	}
	if tree.Children[0].WorkflowName != "child" {
		t.Fatalf("expected child WorkflowName=%q, got %q", "child", tree.Children[0].WorkflowName)
	}
	if tree.Children[0].ParentRunID != parentID {
		t.Fatalf("child parent_run_id mismatch: got %d want %d",
			tree.Children[0].ParentRunID, parentID)
	}
}
