package indexer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/swift"
	typescript "github.com/smacker/go-tree-sitter/typescript/typescript"
	tsx "github.com/smacker/go-tree-sitter/typescript/tsx"
)

var registry map[string]*LanguageExtractor

func init() {
	registry = make(map[string]*LanguageExtractor)

	// ── Go ──────────────────────────────────────────────────────────────
	goExt := &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(method_declaration name: (field_identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(type_declaration (type_spec name: (type_identifier) @name type: (struct_type))) @decl`, Kind: KindType},
			{Query: `(type_declaration (type_spec name: (type_identifier) @name type: (interface_type))) @decl`, Kind: KindInterface},
			{Query: `(const_spec name: (identifier) @name) @decl`, Kind: KindConst},
			{Query: `(var_spec name: (identifier) @name) @decl`, Kind: KindVar},
		},
		ImportQuery: `(import_spec path: (interpreted_string_literal) @path)`,
		CallQuery: strings.Join([]string{
			`(call_expression function: (identifier) @callee)`,
			`(call_expression function: (selector_expression field: (field_identifier) @callee))`,
		}, "\n"),
		PathResolver: goPathResolver,
	}
	registry["go"] = goExt

	// ── Python ──────────────────────────────────────────────────────────
	pyExt := &LanguageExtractor{
		Language:   "python",
		Grammar:    python.GetLanguage(),
		Extensions: []string{".py", ".pyi"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_definition name: (identifier) @name) @decl`, Kind: KindClass},
		},
		ImportQuery: strings.Join([]string{
			`(import_from_statement module_name: (dotted_name) @path)`,
			`(import_statement name: (dotted_name) @path)`,
		}, "\n"),
		CallQuery: strings.Join([]string{
			`(call function: (identifier) @callee)`,
			`(call function: (attribute attribute: (identifier) @callee))`,
		}, "\n"),
		PathResolver: pythonPathResolver,
	}
	registry["python"] = pyExt

	// ── JavaScript ──────────────────────────────────────────────────────
	jsExt := &LanguageExtractor{
		Language:   "javascript",
		Grammar:    javascript.GetLanguage(),
		Extensions: []string{".js", ".mjs", ".cjs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(method_definition name: (property_identifier) @name) @decl`, Kind: KindMethod},
		},
		ImportQuery: `(import_statement source: (string) @path)`,
		CallQuery: strings.Join([]string{
			`(call_expression function: (identifier) @callee)`,
			`(call_expression function: (member_expression property: (property_identifier) @callee))`,
		}, "\n"),
		PathResolver: jsPathResolver,
	}
	registry["javascript"] = jsExt
	// JSX uses the same grammar as JavaScript in smacker/go-tree-sitter.
	jsxExt := copyExtractor(jsExt)
	jsxExt.Language = "jsx"
	jsxExt.Extensions = []string{".jsx"}
	registry["jsx"] = jsxExt

	// ── TypeScript ──────────────────────────────────────────────────────
	tsExt := &LanguageExtractor{
		Language:   "typescript",
		Grammar:    typescript.GetLanguage(),
		Extensions: []string{".ts"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(interface_declaration name: (type_identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(type_alias_declaration name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(method_definition name: (property_identifier) @name) @decl`, Kind: KindMethod},
		},
		ImportQuery: `(import_statement source: (string) @path)`,
		CallQuery: strings.Join([]string{
			`(call_expression function: (identifier) @callee)`,
			`(call_expression function: (member_expression property: (property_identifier) @callee))`,
		}, "\n"),
		PathResolver: jsPathResolver,
	}
	registry["typescript"] = tsExt

	// ── TSX ─────────────────────────────────────────────────────────────
	tsxExt := &LanguageExtractor{
		Language:   "tsx",
		Grammar:    tsx.GetLanguage(),
		Extensions: []string{".tsx"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(interface_declaration name: (type_identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(type_alias_declaration name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(method_definition name: (property_identifier) @name) @decl`, Kind: KindMethod},
		},
		ImportQuery: `(import_statement source: (string) @path)`,
		CallQuery: strings.Join([]string{
			`(call_expression function: (identifier) @callee)`,
			`(call_expression function: (member_expression property: (property_identifier) @callee))`,
		}, "\n"),
		PathResolver: jsPathResolver,
	}
	registry["tsx"] = tsxExt

	// ── Ruby ────────────────────────────────────────────────────────────
	registry["ruby"] = &LanguageExtractor{
		Language:   "ruby",
		Grammar:    ruby.GetLanguage(),
		Extensions: []string{".rb"},
		SymbolQueries: []SymbolQuery{
			{Query: `(method name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class name: (constant) @name) @decl`, Kind: KindClass},
			{Query: `(module name: (constant) @name) @decl`, Kind: KindType},
		},
		CallQuery: `(call method: (identifier) @callee) @decl`,
	}

	// ── Rust ────────────────────────────────────────────────────────────
	registry["rust"] = &LanguageExtractor{
		Language:   "rust",
		Grammar:    rust.GetLanguage(),
		Extensions: []string{".rs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_item name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(struct_item name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(enum_item name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(trait_item name: (type_identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(impl_item type: (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(use_declaration argument: (_) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}

	// ── Java ────────────────────────────────────────────────────────────
	registry["java"] = &LanguageExtractor{
		Language:   "java",
		Grammar:    java.GetLanguage(),
		Extensions: []string{".java"},
		SymbolQueries: []SymbolQuery{
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(method_declaration name: (identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(interface_declaration name: (identifier) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(import_declaration (scoped_identifier) @path)`,
		CallQuery:   `(method_invocation name: (identifier) @callee)`,
	}

	// ── C ───────────────────────────────────────────────────────────────
	registry["c"] = &LanguageExtractor{
		Language:   "c",
		Grammar:    c.GetLanguage(),
		Extensions: []string{".c", ".h"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @decl`, Kind: KindFunction},
			{Query: `(struct_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(enum_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(preproc_include path: (_) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}

	// ── C++ ─────────────────────────────────────────────────────────────
	registry["cpp"] = &LanguageExtractor{
		Language:   "cpp",
		Grammar:    cpp.GetLanguage(),
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @decl`, Kind: KindFunction},
			{Query: `(class_specifier name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(struct_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(preproc_include path: (_) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}

	// ── C# ──────────────────────────────────────────────────────────────
	registry["csharp"] = &LanguageExtractor{
		Language:   "csharp",
		Grammar:    csharp.GetLanguage(),
		Extensions: []string{".cs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(method_declaration name: (identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(interface_declaration name: (identifier) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(using_directive (identifier) @path)`,
		CallQuery:   `(invocation_expression function: (identifier) @callee)`,
	}

	// ── PHP ─────────────────────────────────────────────────────────────
	registry["php"] = &LanguageExtractor{
		Language:   "php",
		Grammar:    php.GetLanguage(),
		Extensions: []string{".php"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (name) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (name) @name) @decl`, Kind: KindClass},
			{Query: `(method_declaration name: (name) @name) @decl`, Kind: KindMethod},
			{Query: `(interface_declaration name: (name) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(use_declaration (name) @path)`,
		CallQuery:   `(function_call_expression function: (name) @callee)`,
	}

	// ── Scala ───────────────────────────────────────────────────────────
	registry["scala"] = &LanguageExtractor{
		Language:   "scala",
		Grammar:    scala.GetLanguage(),
		Extensions: []string{".scala"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_definition name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(object_definition name: (identifier) @name) @decl`, Kind: KindType},
			{Query: `(trait_definition name: (identifier) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(import_declaration path: (_) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}

	// ── Swift ───────────────────────────────────────────────────────────
	// Swift grammar: structs are class_declaration with declaration_kind="struct".
	// We capture all class_declaration (class/struct) and protocol_declaration.
	registry["swift"] = &LanguageExtractor{
		Language:   "swift",
		Grammar:    swift.GetLanguage(),
		Extensions: []string{".swift"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (simple_identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(protocol_declaration name: (type_identifier) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(import_declaration (identifier) @path)`,
		CallQuery:   `(call_expression (simple_identifier) @callee)`,
	}

	// ── Kotlin ──────────────────────────────────────────────────────────
	registry["kotlin"] = &LanguageExtractor{
		Language:   "kotlin",
		Grammar:    kotlin.GetLanguage(),
		Extensions: []string{".kt", ".kts"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration (simple_identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(object_declaration (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(import_header (identifier) @path)`,
		CallQuery:   `(call_expression (simple_identifier) @callee)`,
	}
}

// ExtractorForLanguage returns the registered extractor for a language name.
// Returns nil for unknown languages.
func ExtractorForLanguage(lang string) *LanguageExtractor {
	return registry[lang]
}

// readFirstModuleLine reads go.mod and returns the module path.
func readFirstModuleLine(goModPath string) string {
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

// goPathResolver resolves Go import paths relative to the module root.
func goPathResolver(importPath, fromFile, repoRoot string) string {
	modPath := readFirstModuleLine(filepath.Join(repoRoot, "go.mod"))
	if modPath == "" {
		return ""
	}
	if !strings.HasPrefix(importPath, modPath) {
		return "" // external import
	}
	rel := strings.TrimPrefix(importPath, modPath)
	rel = strings.TrimPrefix(rel, "/")
	return rel
}

// pythonPathResolver converts dotted Python imports to file paths.
func pythonPathResolver(importPath, fromFile, repoRoot string) string {
	// Convert dotted path to directory path: "foo.bar.baz" → "foo/bar/baz"
	return strings.ReplaceAll(importPath, ".", "/")
}

// jsPathResolver resolves relative JS/TS imports.
func jsPathResolver(importPath, fromFile, repoRoot string) string {
	if !strings.HasPrefix(importPath, ".") {
		return "" // external package
	}
	dir := filepath.Dir(fromFile)
	resolved := filepath.Join(dir, importPath)
	return filepath.Clean(resolved)
}

// copyExtractor returns a shallow copy of a LanguageExtractor.
func copyExtractor(e *LanguageExtractor) *LanguageExtractor {
	cp := *e
	return &cp
}

// AllLanguages returns a list of all registered language names.
func AllLanguages() []string {
	langs := make([]string, 0, len(registry))
	for k := range registry {
		langs = append(langs, k)
	}
	return langs
}

// compile-time check: ensure grammar functions return the right type.
var _ *sitter.Language = golang.GetLanguage()
