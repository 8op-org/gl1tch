package indexer

import (
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"component.tsx", "tsx"},
		{"style.css", "css"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"README.md", "markdown"},
		{"script.sh", "shell"},
		{"run.bash", "shell"},
		{"query.sql", "sql"},
		{"app.rs", "rust"},
		{"Main.java", "java"},
		{"page.vue", "vue"},
		{"lib.rb", "ruby"},
		{"data.json", "json"},
		{"page.html", "html"},
		{"app.ts", "typescript"},
		{"component.jsx", "jsx"},
		{"Makefile", "other"},
		{"Dockerfile", "other"},
		{".env", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestChunkContent(t *testing.T) {
	t.Run("small file single chunk", func(t *testing.T) {
		content := "hello world"
		chunks := ChunkContent(content)
		if len(chunks) != 1 {
			t.Fatalf("expected 1 chunk, got %d", len(chunks))
		}
		if chunks[0] != content {
			t.Errorf("chunk content mismatch")
		}
	})

	t.Run("exact boundary", func(t *testing.T) {
		content := strings.Repeat("a", 1500)
		chunks := ChunkContent(content)
		if len(chunks) != 1 {
			t.Fatalf("expected 1 chunk for exactly 1500 chars, got %d", len(chunks))
		}
	})

	t.Run("large file multiple chunks", func(t *testing.T) {
		content := strings.Repeat("x", 4000)
		chunks := ChunkContent(content)
		if len(chunks) < 2 {
			t.Fatalf("expected multiple chunks, got %d", len(chunks))
		}

		// First chunk should be chunkSize
		if len(chunks[0]) != 1500 {
			t.Errorf("first chunk length = %d, want 1500", len(chunks[0]))
		}

		// Verify overlap: end of chunk[0] should match start of chunk[1]
		overlap := chunks[0][1500-150:]
		if !strings.HasPrefix(chunks[1], overlap) {
			t.Error("chunks do not overlap correctly")
		}

		// All content should be covered
		total := len(chunks[0])
		for i := 1; i < len(chunks); i++ {
			total += len(chunks[i]) - 150 // subtract overlap
		}
		// last chunk may be shorter, so total >= original content
		if total < len(content) {
			t.Errorf("chunks don't cover all content: total covered %d, content %d", total, len(content))
		}
	})
}

func TestClassifyFiles(t *testing.T) {
	existing := map[string]string{
		"main.go":    "hash1",
		"old.go":     "hash2",
		"changed.go": "hash3",
	}
	current := map[string]string{
		"main.go":    "hash1",   // unchanged
		"changed.go": "hash999", // changed
		"new.go":     "hash4",   // new
	}
	toIndex, toDelete := classifyFiles(existing, current)
	if len(toIndex) != 2 {
		t.Errorf("toIndex: got %d, want 2", len(toIndex))
	}
	if len(toDelete) != 1 {
		t.Errorf("toDelete: got %d, want 1", len(toDelete))
	}
	if toDelete[0] != "old.go" {
		t.Errorf("toDelete[0] = %q, want old.go", toDelete[0])
	}
}

func TestFileHashCompute(t *testing.T) {
	h1 := fileHash([]byte("hello world"))
	h2 := fileHash([]byte("hello world"))
	h3 := fileHash([]byte("different"))
	if h1 != h2 {
		t.Error("same content should produce same hash")
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
	if h1 == "" {
		t.Error("hash should not be empty")
	}
}

func TestExtractSymbols(t *testing.T) {
	t.Run("go code", func(t *testing.T) {
		content := `package main

func IndexRepo(root string) error {
	return nil
}

type CodeDoc struct {
	Path string
}

var defaultTimeout = 30

const maxRetries = 3

interface Indexer {
	Index()
}
`
		symbols := ExtractSymbols(content)
		for _, want := range []string{"IndexRepo", "CodeDoc", "defaultTimeout", "maxRetries", "Indexer"} {
			if !strings.Contains(symbols, want) {
				t.Errorf("symbols %q missing %q", symbols, want)
			}
		}
	})

	t.Run("python code", func(t *testing.T) {
		content := `
class MyService:
    def handle_request(self, req):
        pass

def main():
    svc = MyService()
`
		symbols := ExtractSymbols(content)
		for _, want := range []string{"MyService", "handle_request", "main"} {
			if !strings.Contains(symbols, want) {
				t.Errorf("symbols %q missing %q", symbols, want)
			}
		}
	})

	t.Run("empty content", func(t *testing.T) {
		symbols := ExtractSymbols("")
		if symbols != "" {
			t.Errorf("expected empty symbols, got %q", symbols)
		}
	})
}
