// Package ai provides the local AI engine using llama.cpp WASM.
package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aio-stack/aio-stack/internal/platform"
	"github.com/aio-stack/aio-stack/internal/optimizer"
	"github.com/aio-stack/aio-stack/pkg/types"
)

// Engine is the AI inference engine.
type Engine struct {
	config    *Config
	platform  *platform.Info
	optimizer *optimizer.Optimizer
	models    map[string]*Model
	mu        sync.RWMutex
	running   bool
}

// Config holds AI engine configuration.
type Config struct {
	Enabled         bool              `toml:"enabled" json:"enabled"`
	ModelDir        string            `toml:"model_dir" json:"model_dir"`
	DefaultModel    string            `toml:"default_model" json:"default_model"`
	Models          map[string]string `toml:"models" json:"models"`
	ContextWindow   int               `toml:"context_window" json:"context_window"`
	MaxTokens       int               `toml:"max_tokens" json:"max_tokens"`
	Temperature     float32           `toml:"temperature" json:"temperature"`
	TopP            float32           `toml:"top_p" json:"top_p"`
	TopK            int               `toml:"top_k" json:"top_k"`
	RepeatPenalty   float32           `toml:"repeat_penalty" json:"repeat_penalty"`
	NumThreads      int               `toml:"num_threads" json:"num_threads"`
	UseGPU          bool              `toml:"use_gpu" json:"use_gpu"`
	GPULayers       int               `toml:"gpu_layers" json:"gpu_layers"`
	EnableWASM      bool              `toml:"enable_wasm" json:"enable_wasm"`
	ModelCacheSize  int               `toml:"model_cache_size" json:"model_cache_size"`
	SystemPrompt    string            `toml:"system_prompt" json:"system_prompt"`
	PromptTemplates map[string]string `toml:"prompt_templates" json:"prompt_templates"`
}

// Model represents a loaded AI model.
type Model struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Type        types.AIModelType      `json:"type"`
	ContextSize int                    `json:"context_size"`
	Parameters  int64                  `json:"parameters"`
	Loaded      bool                   `json:"loaded"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// InferenceRequest represents an inference request.
type InferenceRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	SystemPrompt string                `json:"system_prompt,omitempty"`
	Context     string                 `json:"context,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float32                `json:"temperature,omitempty"`
	TopP        float32                `json:"top_p,omitempty"`
	TopK        int                    `json:"top_k,omitempty"`
	RepeatPenalty float32             `json:"repeat_penalty,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// InferenceResponse represents an inference response.
type InferenceResponse struct {
	Text         string                 `json:"text"`
	Tokens       int                    `json:"tokens"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StreamChunk represents a streaming inference chunk.
type StreamChunk struct {
	Text         string `json:"text"`
	Done         bool   `json:"done"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// CompletionRequest is a simple completion request.
type CompletionRequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
}

// NewEngine creates a new AI engine.
func NewEngine(cfg *Config, pf *platform.Info, opt *optimizer.Optimizer) (*Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	e := &Engine{
		config:    cfg,
		platform:  pf,
		optimizer: opt,
		models:    make(map[string]*Model),
	}

	// Initialize WASM runtime if enabled
	if cfg.EnableWASM {
		if err := e.initWASM(); err != nil {
			return nil, fmt.Errorf("WASM init: %w", err)
		}
	}

	return e, nil
}

// DefaultConfig returns default AI configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:        true,
		DefaultModel:   "phi-3-mini-4k-instruct-q4_k_m.gguf",
		ContextWindow:  4096,
		MaxTokens:      2048,
		Temperature:    0.7,
		TopP:           0.9,
		TopK:           40,
		RepeatPenalty:  1.1,
		NumThreads:     0,
		UseGPU:         false,
		GPULayers:      -1,
		EnableWASM:     true,
		ModelCacheSize: 2,
		SystemPrompt:   "You are a helpful coding assistant. Provide concise, accurate answers.",
		PromptTemplates: map[string]string{
			"code_explanation": `Explain the following {{.language}} code in simple terms:

{{.code}}`,
			"commit_message":     `Generate a conventional commit message for these changes:

{{.diff}}`,
			"bug_analysis":       `Analyze this error and suggest fixes:

{{.error}}

Context: {{.context}}`,
			"code_review":        `Review this code for issues:

{{.code}}`,
			"documentation":      `Generate documentation for this code:

{{.code}}`,
		},
	}
}

// initWASM initializes the WASM runtime for llama.cpp.
func (e *Engine) initWASM() error {
	// TODO: Initialize wasmer/wazero runtime and load llama.cpp WASM module
	// This would:
	// 1. Create wasmer/wazero engine
	// 2. Load llama.cpp WASM module (compiled from llama.cpp with WASM target)
	// 3. Set up memory, imports, exports
	// 4. Pre-load default model
	return nil
}

// Start starts the AI engine.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("engine already running")
	}

	// Load default model if specified
	if e.config.DefaultModel != "" {
		if _, err := e.LoadModel(ctx, e.config.DefaultModel); err != nil {
			// Log but don't fail - model can be loaded later
		}
	}

	e.running = true
	return nil
}

// Shutdown shuts down the AI engine.
func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	// Unload all models
	for name := range e.models {
		e.unloadModel(name)
	}

	e.running = false
	return nil
}

// LoadModel loads an AI model.
func (e *Engine) LoadModel(ctx context.Context, name string) (*Model, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if already loaded
	if model, ok := e.models[name]; ok && model.Loaded {
		return model, nil
	}

	// Check cache size limit
	if len(e.models) >= e.config.ModelCacheSize {
		// Evict oldest model (simple LRU would be better)
		for n := range e.models {
			e.unloadModel(n)
			break
		}
	}

	// Resolve model path
	path := e.resolveModelPath(name)
	if path == "" {
		return nil, fmt.Errorf("model not found: %s", name)
	}

	// TODO: Actually load model via llama.cpp WASM
	model := &Model{
		Name:        name,
		Path:        path,
		Type:        types.AIModelTypeGGUF,
		ContextSize: e.config.ContextWindow,
		Loaded:      true,
		Metadata:    make(map[string]interface{}),
	}

	e.models[name] = model
	return model, nil
}

// UnloadModel unloads a model.
func (e *Engine) UnloadModel(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.unloadModel(name)
}

func (e *Engine) unloadModel(name string) error {
	if model, ok := e.models[name]; ok {
		// TODO: Free WASM resources
		model.Loaded = false
		delete(e.models, name)
	}
	return nil
}

// resolveModelPath resolves a model name to a file path.
func (e *Engine) resolveModelPath(name string) string {
	// Check explicit models map
	if path, ok := e.config.Models[name]; ok {
		return path
	}

	// Check model directory
	if e.config.ModelDir != "" {
		// Try common extensions
		for _, ext := range []string{".gguf", ".bin", ".onnx", ".wasm"} {
			path := filepath.Join(e.config.ModelDir, name+ext)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// Infer runs inference on a prompt.
func (e *Engine) Infer(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.running {
		return nil, fmt.Errorf("engine not running")
	}

	// Use default model if not specified
	modelName := req.Model
	if modelName == "" {
		modelName = e.config.DefaultModel
	}

	// Ensure model is loaded
	if _, ok := e.models[modelName]; !ok {
		if _, err := e.LoadModel(ctx, modelName); err != nil {
			return nil, fmt.Errorf("model load failed: %w", err)
		}
	}

	// TODO: Run actual inference via llama.cpp WASM
	// For now, return a mock response
	return &InferenceResponse{
		Text:         "[AI Response would appear here - inference via llama.cpp WASM]",
		Tokens:       0,
		FinishReason: "stop",
	}, nil
}

// InferStream runs streaming inference.
func (e *Engine) InferStream(ctx context.Context, req *InferenceRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)

	go func() {
		defer close(ch)
		resp, err := e.Infer(ctx, req)
		if err != nil {
			ch <- StreamChunk{Text: "", Done: true, FinishReason: "error"}
			return
		}
		// Simulate streaming by chunking
		ch <- StreamChunk{Text: resp.Text, Done: true, FinishReason: resp.FinishReason}
	}()

	return ch, nil
}

// GenerateCompletion generates a completion for a prompt.
func (e *Engine) GenerateCompletion(ctx context.Context, req *CompletionRequest) (string, error) {
	inferReq := &InferenceRequest{
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	resp, err := e.Infer(ctx, inferReq)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// ApplyPromptTemplate applies a prompt template with variables.
func (e *Engine) ApplyPromptTemplate(templateName string, vars map[string]interface{}) (string, error) {
	e.mu.RLock()
	template, ok := e.config.PromptTemplates[templateName]
	e.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Simple template substitution
	result := template
	for key, value := range vars {
		placeholder := "{{." + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result, nil
}

// GetModel returns a loaded model.
func (e *Engine) GetModel(name string) (*Model, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	model, ok := e.models[name]
	return model, ok
}

// ListModels returns all loaded models.
func (e *Engine) ListModels() []*Model {
	e.mu.RLock()
	defer e.mu.RUnlock()

	models := make([]*Model, 0, len(e.models))
	for _, m := range e.models {
		models = append(models, m)
	}
	return models
}

// SetSystemPrompt sets the default system prompt.
func (e *Engine) SetSystemPrompt(prompt string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.SystemPrompt = prompt
}

// AddPromptTemplate adds a prompt template.
func (e *Engine) AddPromptTemplate(name, template string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.config.PromptTemplates == nil {
		e.config.PromptTemplates = make(map[string]string)
	}
	e.config.PromptTemplates[name] = template
}

// GetPromptTemplate returns a prompt template.
func (e *Engine) GetPromptTemplate(name string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	template, ok := e.config.PromptTemplates[name]
	return template, ok
}