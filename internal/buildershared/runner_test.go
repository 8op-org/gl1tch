package buildershared

import (
	"context"
	"testing"
)

func TestRunnerPanelClear(t *testing.T) {
	r := NewRunnerPanel()
	r.lines = []string{"line1", "line2"}
	r.running = true
	r.statusMsg = "oops"
	r.statusErr = true

	r2 := r.Clear()
	if r2.IsRunning() {
		t.Error("Clear: expected IsRunning=false")
	}
	if len(r2.Lines()) != 0 {
		t.Errorf("Clear: expected no lines, got %d", len(r2.Lines()))
	}
	if r2.statusMsg != "" {
		t.Errorf("Clear: expected empty statusMsg, got %q", r2.statusMsg)
	}
	if r2.statusErr {
		t.Error("Clear: expected statusErr=false")
	}
}

func TestRunnerPanelStartRun(t *testing.T) {
	r := NewRunnerPanel()
	ch := make(chan string, 5)
	ch <- "hello"
	ch <- "world"
	close(ch)

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	r2, cmd := r.StartRun(ch, cancel)

	if !r2.IsRunning() {
		t.Error("StartRun: expected IsRunning=true")
	}
	if len(r2.Lines()) != 0 {
		t.Error("StartRun: expected no lines yet (streaming hasn't happened)")
	}
	if cmd == nil {
		t.Fatal("StartRun: expected a wait command, got nil")
	}

	// Simulate receiving the first line.
	msg := cmd()
	rl, ok := msg.(RunLineMsg)
	if !ok {
		t.Fatalf("expected RunLineMsg, got %T", msg)
	}
	if string(rl) != "hello" {
		t.Errorf("expected hello, got %q", string(rl))
	}

	// Process line through Update.
	r3, cmd2 := r2.Update(rl)
	if len(r3.Lines()) != 1 || r3.Lines()[0] != "hello" {
		t.Errorf("after Update: expected [hello], got %v", r3.Lines())
	}
	if cmd2 == nil {
		t.Fatal("expected next wait cmd, got nil")
	}

	// Receive "world".
	msg2 := cmd2()
	rl2, ok := msg2.(RunLineMsg)
	if !ok {
		t.Fatalf("expected RunLineMsg, got %T", msg2)
	}
	r4, cmd3 := r3.Update(rl2)
	if len(r4.Lines()) != 2 {
		t.Errorf("expected 2 lines, got %d", len(r4.Lines()))
	}

	// Channel closed — RunDoneMsg.
	if cmd3 == nil {
		t.Fatal("expected RunDoneMsg cmd")
	}
	msg3 := cmd3()
	done, ok := msg3.(RunDoneMsg)
	if !ok {
		t.Fatalf("expected RunDoneMsg, got %T", msg3)
	}
	r5, _ := r4.Update(done)
	if r5.IsRunning() {
		t.Error("after RunDoneMsg: expected IsRunning=false")
	}
	if r5.statusMsg != "run complete" {
		t.Errorf("expected 'run complete', got %q", r5.statusMsg)
	}
}

func TestRunnerPanelLineCapAt200(t *testing.T) {
	r := NewRunnerPanel()
	for i := range 250 {
		r.lines = append(r.lines, string(rune('a'+i%26)))
	}
	// Simulate an Update that receives one more line when already at 200.
	r.lines = r.lines[:200] // cap to exactly 200
	r2, _ := r.Update(RunLineMsg("new"))
	if len(r2.Lines()) > 200 {
		t.Errorf("expected ≤200 lines, got %d", len(r2.Lines()))
	}
}

func TestRunnerPanelSetFocused(t *testing.T) {
	r := NewRunnerPanel()
	if r.focused {
		t.Error("new runner should not be focused")
	}
	r = r.SetFocused(true)
	if !r.focused {
		t.Error("SetFocused(true) should set focused")
	}
}
