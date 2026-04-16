package registry

import "testing"

func TestActiveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if v, _ := GetActive(); v != "" {
		t.Fatalf("want empty, got %q", v)
	}
	if err := SetActive("stokagent"); err != nil {
		t.Fatal(err)
	}
	v, err := GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if v != "stokagent" {
		t.Fatalf("want stokagent, got %q", v)
	}
}
