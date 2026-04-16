package pipeline

import "testing"

func TestRenderResourceBinding(t *testing.T) {
	data := map[string]any{
		"input": "",
		"param": map[string]string{},
		"resource": map[string]map[string]string{
			"ensemble": {"path": "/tmp/ensemble", "url": "https://x", "ref": "main", "pin": "sha123"},
		},
	}
	out, err := render("~resource.ensemble.path:~resource.ensemble.pin", scopeFromData(data), nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "/tmp/ensemble:sha123" {
		t.Fatalf("bad render: %q", out)
	}
}

func TestRenderResourceMissingFailsLoud(t *testing.T) {
	data := map[string]any{"input": "", "param": map[string]string{}, "resource": map[string]map[string]string{}}
	_, err := render("x:~resource.missing.path:y", scopeFromData(data), nil)
	if err == nil {
		t.Fatal("expected UndefinedRefError on missing resource, got nil")
	}
	var ure *UndefinedRefError
	if !errorsAsStd(err, &ure) {
		t.Fatalf("expected UndefinedRefError, got %T: %v", err, err)
	}
}
