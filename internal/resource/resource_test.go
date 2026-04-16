package resource

import "testing"

func TestValidateName(t *testing.T) {
	bad := []string{
		"", "/", ".", "..", "foo/bar", "../foo", ".hidden", "foo/../bar", "a/b/c",
	}
	good := []string{"alpha", "my-repo", "my_repo", "repo.v2", "thing123"}
	for _, n := range bad {
		if err := ValidateName(n); err == nil {
			t.Errorf("want error for %q", n)
		}
	}
	for _, n := range good {
		if err := ValidateName(n); err != nil {
			t.Errorf("want no error for %q, got %v", n, err)
		}
	}
}
