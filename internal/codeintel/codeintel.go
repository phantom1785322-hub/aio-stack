// Package codeintel provides code intelligence via tree-sitter.
package codeintel

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/aio-stack/aio-stack/internal/platform"
	"github.com/aio-stack/aio-stack/internal/optimizer"
)

// Engine is the code intelligence engine.
type Engine struct {
	config    *Config
	platform  *platform.Info
	optimizer *optimizer.Optimizer
	parsers   map[string]*Parser
	queries   map[string]*Query
	mu        sync.RWMutex
	running   bool
}

// Config holds code intelligence configuration.
type Config struct {
	Enabled         bool     `toml:"enabled" json:"enabled"`
	Languages       []string `toml:"languages" json:"languages"`
	ParserCacheSize int      `toml:"parser_cache_size" json:"parser_cache_size"`
	TreeCacheSize   int      `toml:"tree_cache_size" json:"tree_cache_size"`
	EnableQueries   bool     `toml:"enable_queries" json:"enable_queries"`
	QueryDirs       []string `toml:"query_dirs" json:"query_dirs"`
	MaxFileSize     int64    `toml:"max_file_size" json:"max_file_size"`
	EnableEdits     bool     `toml:"enable_edits" json:"enable_edits"`
}

// Parser wraps a tree-sitter parser.
type Parser struct {
	Language string
}

// SyntaxTree represents a parsed syntax tree.
type SyntaxTree struct {
	Language string
	Source   string
}

// Query represents a tree-sitter query.
type Query struct {
	Language string
	Source   string
}

// AnalysisResult represents the result of code analysis.
type AnalysisResult struct {
	Language      string                 `json:"language"`
	Symbols       []Symbol               `json:"symbols"`
	Imports       []Import               `json:"imports"`
	Exports       []Export               `json:"exports"`
	Complexity    ComplexityMetrics      `json:"complexity"`
	Issues        []Issue                `json:"issues"`
	Dependencies  []Dependency           `json:"dependencies"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Symbol represents a code symbol (function, class, variable, etc.)
type Symbol struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`       // function, class, variable, constant, etc.
	Signature  string   `json:"signature"`  // function signature, type, etc.
	Location   Location `json:"location"`   // file, line, column
	DocComment string   `json:"doc_comment,omitempty"`
}

// Import represents an import statement.
type Import struct {
	Path       string   `json:"path"`
	Alias      string   `json:"alias,omitempty"`
	Location   Location `json:"location"`
	IsExternal bool     `json:"is_external"`
}

// Export represents an exported symbol.
type Export struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Location Location `json:"location"`
}

// Location represents a source code location.
type Location struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}

// ComplexityMetrics holds code complexity metrics.
type ComplexityMetrics struct {
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
	CognitiveComplexity  int `json:"cognitive_complexity"`
	LinesOfCode          int `json:"lines_of_code"`
	FunctionCount        int `json:"function_count"`
	ClassCount           int `json:"class_count"`
	NestingDepth         int `json:"nesting_depth"`
}

// Issue represents a code issue (lint, security, style, etc.)
type Issue struct {
	Severity  string   `json:"severity"`  // error, warning, info
	Category  string   `json:"category"`  // security, style, performance, bug
	Message   string   `json:"message"`
	Location  Location `json:"location"`
	Rule      string   `json:"rule,omitempty"`
	Fix       string   `json:"fix,omitempty"`
}

// Dependency represents a code dependency.
type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Type        string `json:"type"`        // direct, transitive, dev
	License     string `json:"license,omitempty"`
	Vulnerable  bool   `json:"vulnerable"`
	Outdated    bool   `json:"outdated"`
}

// NewEngine creates a new code intelligence engine.
func NewEngine(cfg *Config, pf *platform.Info, opt *optimizer.Optimizer) (*Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	e := &Engine{
		config:    cfg,
		platform:  pf,
		optimizer: opt,
		parsers:   make(map[string]*Parser),
		queries:   make(map[string]*Query),
	}

	// Initialize tree-sitter if enabled
	if cfg.Enabled {
		if err := e.initTreeSitter(); err != nil {
			return nil, fmt.Errorf("tree-sitter init: %w", err)
		}
	}

	return e, nil
}

// DefaultConfig returns default code intelligence configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		Languages:       []string{}, // Empty = all supported
		ParserCacheSize: 32,
		TreeCacheSize:   128,
		EnableQueries:   true,
		MaxFileSize:     1048576, // 1MB
		EnableEdits:     false,
	}
}

// initTreeSitter initializes tree-sitter parsers.
func (e *Engine) initTreeSitter() error {
	// TODO: Initialize tree-sitter parsers for supported languages
	supportedLanguages := []string{
		"go", "typescript", "javascript", "python", "rust",
		"java", "cpp", "c", "csharp", "ruby", "php",
		"swift", "kotlin", "scala", "lua", "bash",
	}

	// Filter by configured languages
	if len(e.config.Languages) > 0 {
		filtered := make([]string, 0)
		for _, lang := range supportedLanguages {
			for _, cfgLang := range e.config.Languages {
				if lang == cfgLang {
					filtered = append(filtered, lang)
					break
				}
			}
		}
		supportedLanguages = filtered
	}

	// Create parsers for each language
	for _, lang := range supportedLanguages {
		p, err := e.createParser(lang)
		if err != nil {
			// Log but continue
			continue
		}
		e.parsers[lang] = p
	}

	return nil
}

// createParser creates a parser for a language.
func (e *Engine) createParser(lang string) (*Parser, error) {
	// TODO: Create actual tree-sitter parser
	return &Parser{Language: lang}, nil
}

// Start starts the code intelligence engine.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("engine already running")
	}

	e.running = true
	return nil
}

// Shutdown shuts down the code intelligence engine.
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	// Clear caches
	e.parsers = make(map[string]*Parser)
	e.queries = make(map[string]*Query)

	e.running = false
	return nil
}

// AnalyzeFile analyzes a source file.
func (e *Engine) AnalyzeFile(ctx context.Context, filePath, source string) (*AnalysisResult, error) {
	// Detect language from file extension
	lang := e.detectLanguage(filePath)
	if lang == "" {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	_, ok := e.parsers[lang]
	if !ok {
		return nil, fmt.Errorf("no parser for language: %s", lang)
	}

	// TODO: Parse with tree-sitter and extract symbols, imports, etc.
	return &AnalysisResult{
		Language: lang,
		Symbols:  []Symbol{},
		Imports:  []Import{},
		Exports:  []Export{},
		Complexity: ComplexityMetrics{},
		Issues:   []Issue{},
		Dependencies: []Dependency{},
	}, nil
}

// AnalyzeProject analyzes a project directory.
func (e *Engine) AnalyzeProject(ctx context.Context, rootPath string) (*ProjectAnalysis, error) {
	// TODO: Walk directory, analyze files, build dependency graph
	return &ProjectAnalysis{
		RootPath:  rootPath,
		Files:     []*AnalysisResult{},
		Symbols:   []Symbol{},
		Imports:   []Import{},
		Graph:     &DependencyGraph{},
	}, nil
}

// ProjectAnalysis represents analysis of an entire project.
type ProjectAnalysis struct {
	RootPath   string              `json:"root_path"`
	Files      []*AnalysisResult   `json:"files"`
	Symbols    []Symbol            `json:"symbols"`
	Imports    []Import            `json:"imports"`
	Graph      *DependencyGraph    `json:"graph"`
}

// DependencyGraph represents the project dependency graph.
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a node in the dependency graph.
type GraphNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // file, package, module
	Language string `json:"language"`
}

// GraphEdge represents an edge in the dependency graph.
type GraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"` // import, reference, inheritance
	Weight int    `json:"weight"`
}

// detectLanguage detects the programming language from file extension.
func (e *Engine) detectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".py", ".pyw", ".pyi":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx", ".hpp", ".h", ".hh":
		return "cpp"
	case ".c":
		return "c"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php", ".phtml":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala", ".sc":
		return "scala"
	case ".lua":
		return "lua"
	case ".sh", ".bash", ".zsh", ".fish":
		return "bash"
	default:
		return ""
	}
}

// GetParser returns a parser for a language.
func (e *Engine) GetParser(lang string) (*Parser, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	parser, ok := e.parsers[lang]
	return parser, ok
}

// SupportedLanguages returns the list of supported languages.
func (e *Engine) SupportedLanguages() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	langs := make([]string, 0, len(e.parsers))
	for lang := range e.parsers {
		langs = append(langs, lang)
	}
	return langs
}