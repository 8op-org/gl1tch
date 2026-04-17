package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
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
	".worktrees":   true,
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

// IndexOpts controls indexing behavior.
type IndexOpts struct {
	Repo        string
	ESURL       string
	Languages   []string // empty = all
	Full        bool     // force full re-index
	SymbolsOnly bool     // skip content chunks
	Stats       bool
}

// fileHash returns the SHA-256 hex digest of content.
func fileHash(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// classifyFiles compares existing file hashes (from ES) with current file hashes
// (from the filesystem) and returns which files need indexing and which need deletion.
func classifyFiles(existing, current map[string]string) (toIndex []string, toDelete []string) {
	for file, hash := range current {
		oldHash, ok := existing[file]
		if !ok || oldHash != hash {
			toIndex = append(toIndex, file)
		}
	}
	for file := range existing {
		if _, ok := current[file]; !ok {
			toDelete = append(toDelete, file)
		}
	}
	sort.Strings(toIndex)
	sort.Strings(toDelete)
	return toIndex, toDelete
}

// IndexRepo walks a repo, chunks files, and indexes them into ES.
// It delegates to IndexRepoGraph with default options.
func IndexRepo(root string, es *esearch.Client) error {
	return IndexRepoGraph(root, es, IndexOpts{Stats: true})
}

// fileEntry holds the content and metadata for a single file during indexing.
type fileEntry struct {
	relPath  string
	lang     string
	content  []byte
	hash     string
}

// IndexRepoGraph runs the three-phase tree-sitter pipeline:
// Phase 1: Extract symbols, imports, and calls from each file.
// Phase 2+3: Resolve cross-file references into edges.
// Optionally also indexes content chunks for BM25 search.
func IndexRepoGraph(root string, es *esearch.Client, opts IndexOpts) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Resolve repo name.
	repoName := opts.Repo
	if repoName == "" {
		repoName = filepath.Base(absRoot)
	}

	symbolsIndex := esearch.IndexSymbolsPrefix + repoName
	edgesIndex := esearch.IndexEdgesPrefix + repoName
	codeIndex := "glitch-code-" + repoName

	ctx := context.Background()

	fmt.Fprintf(os.Stderr, "Indexing %s → symbols:%s edges:%s\n", absRoot, symbolsIndex, edgesIndex)

	// Ensure indices exist.
	if err := es.EnsureIndex(ctx, symbolsIndex, esearch.SymbolsMapping); err != nil {
		return fmt.Errorf("ensure symbols index: %w", err)
	}
	if err := es.EnsureIndex(ctx, edgesIndex, esearch.EdgesMapping); err != nil {
		return fmt.Errorf("ensure edges index: %w", err)
	}
	if !opts.SymbolsOnly {
		if err := es.EnsureIndex(ctx, codeIndex, esMapping); err != nil {
			return fmt.Errorf("ensure code index: %w", err)
		}
	}

	// Build language filter set if specified.
	langFilter := make(map[string]bool, len(opts.Languages))
	for _, l := range opts.Languages {
		langFilter[l] = true
	}

	// Walk filesystem, collect file metadata.
	currentFiles := make(map[string]string)       // relPath → hash
	fileEntries := make(map[string]*fileEntry)     // relPath → entry

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
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

		relPath, _ := filepath.Rel(absRoot, path)
		lang := DetectLanguage(name)

		// Apply language filter.
		if len(langFilter) > 0 && !langFilter[lang] {
			return nil
		}

		h := fileHash(data)
		currentFiles[relPath] = h
		fileEntries[relPath] = &fileEntry{
			relPath: relPath,
			lang:    lang,
			content: data,
			hash:    h,
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Determine which files to index.
	var toIndex, toDelete []string
	if opts.Full {
		toIndex = make([]string, 0, len(currentFiles))
		for f := range currentFiles {
			toIndex = append(toIndex, f)
		}
		sort.Strings(toIndex)
	} else {
		existing, err := es.TermsAgg(ctx, symbolsIndex, "file", "file_hash", 100000)
		if err != nil {
			return fmt.Errorf("fetch existing hashes: %w", err)
		}
		toIndex, toDelete = classifyFiles(existing, currentFiles)
	}

	if opts.Stats {
		fmt.Fprintf(os.Stderr, "  files: %d total, %d to index, %d to delete\n",
			len(currentFiles), len(toIndex), len(toDelete))
	}

	// Delete stale symbols and edges for files in toDelete and toIndex.
	filesToClean := append(toDelete, toIndex...)
	for _, file := range filesToClean {
		q, _ := json.Marshal(map[string]any{
			"query": map[string]any{
				"term": map[string]any{"file": file},
			},
		})
		_, _ = es.DeleteByQuery(ctx, symbolsIndex, q)
		_, _ = es.DeleteByQuery(ctx, edgesIndex, q)
	}

	// Phase 1 — Extract symbols, imports, and calls.
	var allSymbols []SymbolDoc
	var allImports []UnresolvedImport
	var allCalls []CallSite

	for _, file := range toIndex {
		fe := fileEntries[file]
		if fe == nil {
			continue
		}

		ext := ExtractorForLanguage(fe.lang)
		if ext == nil {
			// No tree-sitter extractor for this language — still content-chunk later.
			continue
		}

		symbols, err := ext.Extract(fe.content, fe.relPath, repoName, fe.hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn: extract symbols %s: %v\n", fe.relPath, err)
			continue
		}
		allSymbols = append(allSymbols, symbols...)

		imports, err := ext.ExtractImports(fe.content, fe.relPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn: extract imports %s: %v\n", fe.relPath, err)
		} else {
			allImports = append(allImports, imports...)
		}

		calls, err := ext.ExtractCalls(fe.content, fe.relPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn: extract calls %s: %v\n", fe.relPath, err)
		} else {
			allCalls = append(allCalls, calls...)
		}
	}

	// Bulk index symbols.
	var symBatch []esearch.BulkDoc
	for _, s := range allSymbols {
		body, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("marshal symbol: %w", err)
		}
		symBatch = append(symBatch, esearch.BulkDoc{ID: s.ID, Body: body})
		if len(symBatch) >= bulkBatch {
			if err := es.BulkIndex(ctx, symbolsIndex, symBatch); err != nil {
				return fmt.Errorf("bulk index symbols: %w", err)
			}
			symBatch = symBatch[:0]
		}
	}
	if len(symBatch) > 0 {
		if err := es.BulkIndex(ctx, symbolsIndex, symBatch); err != nil {
			return fmt.Errorf("bulk index symbols (flush): %w", err)
		}
	}

	// Phase 2+3 — Resolve edges.
	resolver := NewResolver(allSymbols, repoName, absRoot)
	modPath := readFirstModuleLine(filepath.Join(absRoot, "go.mod"))
	if modPath != "" {
		resolver.ModulePath = modPath
	}

	var allEdges []EdgeDoc
	allEdges = append(allEdges, resolver.ResolveContains()...)
	allEdges = append(allEdges, resolver.ResolveImports(allImports)...)
	allEdges = append(allEdges, resolver.ResolveCalls(allCalls)...)

	// Bulk index edges.
	var edgeBatch []esearch.BulkDoc
	for i, e := range allEdges {
		body, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal edge: %w", err)
		}
		docID := fmt.Sprintf("%s-%s-%s-%d", e.SourceID, e.TargetID, e.Kind, i)
		edgeBatch = append(edgeBatch, esearch.BulkDoc{ID: docID, Body: body})
		if len(edgeBatch) >= bulkBatch {
			if err := es.BulkIndex(ctx, edgesIndex, edgeBatch); err != nil {
				return fmt.Errorf("bulk index edges: %w", err)
			}
			edgeBatch = edgeBatch[:0]
		}
	}
	if len(edgeBatch) > 0 {
		if err := es.BulkIndex(ctx, edgesIndex, edgeBatch); err != nil {
			return fmt.Errorf("bulk index edges (flush): %w", err)
		}
	}

	// Content chunks for BM25 search.
	if !opts.SymbolsOnly {
		now := time.Now().UTC().Format(time.RFC3339)
		var codeBatch []esearch.BulkDoc
		chunksTotal := 0

		for _, file := range toIndex {
			fe := fileEntries[file]
			if fe == nil {
				continue
			}

			content := string(fe.content)
			symbols := ExtractSymbols(content)
			chunks := ChunkContent(content)

			for i, chunk := range chunks {
				h := sha256.Sum256([]byte(chunk))
				hash := fmt.Sprintf("%x", h)

				docID := fe.relPath
				if len(chunks) > 1 {
					docID = fmt.Sprintf("%s:chunk%d", fe.relPath, i)
				}

				bodyJSON, err := json.Marshal(CodeDoc{
					Path:      fe.relPath,
					Content:   chunk,
					Repo:      repoName,
					Language:  fe.lang,
					Hash:      hash,
					Symbols:   symbols,
					IndexedAt: now,
				})
				if err != nil {
					return fmt.Errorf("marshal code doc: %w", err)
				}
				codeBatch = append(codeBatch, esearch.BulkDoc{ID: docID, Body: bodyJSON})
				chunksTotal++

				if len(codeBatch) >= bulkBatch {
					if err := es.BulkIndex(ctx, codeIndex, codeBatch); err != nil {
						return fmt.Errorf("bulk index code: %w", err)
					}
					codeBatch = codeBatch[:0]
				}
			}
		}

		if len(codeBatch) > 0 {
			if err := es.BulkIndex(ctx, codeIndex, codeBatch); err != nil {
				return fmt.Errorf("bulk index code (flush): %w", err)
			}
		}

		if opts.Stats {
			fmt.Fprintf(os.Stderr, "  chunks: %d indexed in %s\n", chunksTotal, codeIndex)
		}
	}

	if opts.Stats {
		fmt.Fprintf(os.Stderr, "Done: %d files indexed, %d symbols, %d edges\n",
			len(toIndex), len(allSymbols), len(allEdges))
	}

	return nil
}
