package indexer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	chunkSize    = 1500
	chunkOverlap = 150
	bulkBatch    = 100
	maxFileSize  = 100 * 1024 // 100KB
)

// skipDirs are directory names to skip during walk.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".next":        true,
	"dist":         true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	".tox":         true,
}

// skipExts are file extensions to skip.
var skipExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".gif":  true,
	".ico":  true,
	".woff": true,
	".ttf":  true,
	".eot":  true,
	".svg":  true,
	".pdf":  true,
	".zip":  true,
	".tar":  true,
	".gz":   true,
	".wasm": true,
	".lock": true,
	".map":  true,
}

// skipNames are exact filenames to skip.
var skipNames = map[string]bool{
	"package-lock.json": true,
	".DS_Store":         true,
}

// langMap maps file extensions to language names.
var langMap = map[string]string{
	".go":   "go",
	".py":   "python",
	".js":   "javascript",
	".ts":   "typescript",
	".tsx":  "tsx",
	".jsx":  "jsx",
	".rb":   "ruby",
	".rs":   "rust",
	".java": "java",
	".vue":  "vue",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
	".md":   "markdown",
	".sh":   "shell",
	".bash": "shell",
	".sql":  "sql",
	".html": "html",
	".css":  "css",
}

// symbolPattern matches common declarations across languages.
var symbolPattern = regexp.MustCompile(
	`(?m)^[ \t]*(?:func|def|class|type|const|var|interface|export\s+(?:function|class|const|default\s+function|default\s+class))\s+(\w+)`,
)

// CodeDoc is the document structure indexed into ES.
type CodeDoc struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Repo      string `json:"repo"`
	Language  string `json:"language"`
	Hash      string `json:"hash"`
	Symbols   string `json:"symbols"`
	IndexedAt string `json:"indexed_at"`
}

// DetectLanguage returns the language name for a file extension.
func DetectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "other"
}

// ChunkContent splits content into fixed-size chunks with overlap.
func ChunkContent(content string) []string {
	if len(content) <= chunkSize {
		return []string{content}
	}

	var chunks []string
	for i := 0; i < len(content); i += chunkSize - chunkOverlap {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
		if end == len(content) {
			break
		}
	}
	return chunks
}

// ExtractSymbols greps for function/class/type declarations and returns
// a space-separated string of identifier names.
func ExtractSymbols(content string) string {
	matches := symbolPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}
	seen := make(map[string]bool)
	var symbols []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			symbols = append(symbols, name)
		}
	}
	return strings.Join(symbols, " ")
}

func shouldSkipFile(name string, ext string) bool {
	if skipNames[name] {
		return true
	}
	if skipExts[ext] {
		return true
	}
	if strings.HasSuffix(name, ".min.js") || strings.HasSuffix(name, ".min.css") {
		return true
	}
	return false
}

// esMapping is the index mapping for the code index.
var esMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "path":       { "type": "keyword" },
      "content":    { "type": "text", "analyzer": "standard" },
      "repo":       { "type": "keyword" },
      "language":   { "type": "keyword" },
      "hash":       { "type": "keyword" },
      "symbols":    { "type": "text", "analyzer": "simple" },
      "indexed_at": { "type": "date" }
    }
  }
}`

// ensureIndex creates the ES index if it doesn't exist.
func ensureIndex(esURL, index string) error {
	resp, err := http.Head(esURL + "/" + index)
	if err != nil {
		return fmt.Errorf("ES unreachable: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil // index exists
	}

	req, err := http.NewRequest("PUT", esURL+"/"+index, strings.NewReader(esMapping))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("create index %s: %s — %s", index, resp.Status, buf.String())
	}
	return nil
}

// bulkIndex sends a batch of docs to the ES bulk API.
func bulkIndex(esURL, index string, docs []bulkItem) error {
	var buf bytes.Buffer
	for _, d := range docs {
		meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`, index, d.id)
		buf.WriteString(meta)
		buf.WriteByte('\n')
		body, err := json.Marshal(d.doc)
		if err != nil {
			return err
		}
		buf.Write(body)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequest("POST", esURL+"/_bulk", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		var respBuf bytes.Buffer
		respBuf.ReadFrom(resp.Body)
		return fmt.Errorf("bulk index: %s — %s", resp.Status, respBuf.String())
	}
	return nil
}

type bulkItem struct {
	id  string
	doc CodeDoc
}

// IndexRepo walks a repo, chunks files, and indexes them into ES.
func IndexRepo(root, esURL string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	repoName := filepath.Base(absRoot)
	index := "glitch-code-" + repoName

	fmt.Fprintf(os.Stderr, "Indexing %s → %s...\n", absRoot, index)

	if err := ensureIndex(esURL, index); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var batch []bulkItem
	filesProcessed := 0
	chunksTotal := 0

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}

		name := d.Name()

		if d.IsDir() {
			if skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if shouldSkipFile(name, ext) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		relPath, _ := filepath.Rel(absRoot, path)
		lang := DetectLanguage(name)
		symbols := ExtractSymbols(content)

		chunks := ChunkContent(content)
		for i, chunk := range chunks {
			h := sha256.Sum256([]byte(chunk))
			hash := fmt.Sprintf("%x", h)

			docID := relPath
			if len(chunks) > 1 {
				docID = fmt.Sprintf("%s:chunk%d", relPath, i)
			}

			batch = append(batch, bulkItem{
				id: docID,
				doc: CodeDoc{
					Path:      relPath,
					Content:   chunk,
					Repo:      repoName,
					Language:  lang,
					Hash:      hash,
					Symbols:   symbols,
					IndexedAt: now,
				},
			})
			chunksTotal++
		}

		filesProcessed++
		if filesProcessed%100 == 0 {
			fmt.Fprintf(os.Stderr, "  ... %d files processed\n", filesProcessed)
		}

		if len(batch) >= bulkBatch {
			if err := bulkIndex(esURL, index, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}

		return nil
	})
	if err != nil {
		return err
	}

	// flush remaining
	if len(batch) > 0 {
		if err := bulkIndex(esURL, index, batch); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "Done: %d files processed, %d chunks indexed in %s\n", filesProcessed, chunksTotal, index)
	return nil
}
