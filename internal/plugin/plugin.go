// Package plugin provides the plugin system for AIO Stack.
package plugin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aio-stack/aio-stack/internal/optimizer"
	"github.com/aio-stack/aio-stack/internal/platform"
	"github.com/aio-stack/aio-stack/pkg/types"
	"gopkg.in/yaml.v3"
)

// Manager manages plugins.
type Manager struct {
	config    *Config
	platform  *platform.Info
	optimizer *optimizer.Optimizer
	plugins   map[string]*Plugin
	registry  *Registry
	mu        sync.RWMutex
}

// Config holds plugin configuration.
type Config struct {
	Enabled         bool     `toml:"enabled" json:"enabled"`
	PluginDirs      []string `toml:"plugin_dirs" json:"plugin_dirs"`
	BuiltinPlugins  []string `toml:"builtin_plugins" json:"builtin_plugins"`
	AllowUnsigned   bool     `toml:"allow_unsigned" json:"allow_unsigned"`
	SignatureKeys   []string `toml:"signature_keys" json:"signature_keys"`
	MaxPlugins      int      `toml:"max_plugins" json:"max_plugins"`
	Timeout         types.Duration `toml:"timeout" json:"timeout"`
	EnableWASM      bool     `toml:"enable_wasm" json:"enable_wasm"`
	EnableGoPlugins bool     `toml:"enable_go_plugins" json:"enable_go_plugins"`
	Registry        string   `toml:"registry" json:"registry"`
}

// Plugin represents a loaded plugin.
type Plugin struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	License     string         `json:"license"`
	Homepage    string         `json:"homepage"`
	Repository  string         `json:"repository"`
	Type        types.PluginType `json:"type"`
	Path        string         `json:"path"`
	Enabled     bool           `json:"enabled"`
	Loaded      bool           `json:"loaded"`
	Manifest    *Manifest      `json:"manifest"`
}

// Manifest is the plugin manifest (plugin.yaml).
type Manifest struct {
	Name        string         `yaml:"name" json:"name"`
	Version     string         `yaml:"version" json:"version"`
	Description string         `yaml:"description" json:"description"`
	Author      string         `yaml:"author" json:"author"`
	License     string         `yaml:"license" json:"license"`
	Homepage    string         `yaml:"homepage" json:"homepage"`
	Repository  string         `yaml:"repository" json:"repository"`
	CoreVersion string         `yaml:"core_version" json:"core_version"`
	Dependencies []Dependency  `yaml:"dependencies" json:"dependencies"`
	Provides    Provides       `yaml:"provides" json:"provides"`
	Permissions []string       `yaml:"permissions" json:"permissions"`
}

// Dependency represents a plugin dependency.
type Dependency struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Optional bool  `yaml:"optional" json:"optional"`
}

// Provides represents what the plugin provides.
type Provides struct {
	Commands []CommandSpec `yaml:"commands" json:"commands"`
	Hooks    []HookSpec    `yaml:"hooks" json:"hooks"`
	UI       []UISpec      `yaml:"ui" json:"ui"`
	API      []APISpec     `yaml:"api" json:"api"`
}

// CommandSpec describes a command provided by the plugin.
type CommandSpec struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Aliases     []string `yaml:"aliases" json:"aliases"`
}

// HookSpec describes a hook provided by the plugin.
type HookSpec struct {
	Event   string `yaml:"event" json:"event"`
	Command string `yaml:"command" json:"command"`
}

// UISpec describes a UI panel provided by the plugin.
type UISpec struct {
	Panel    string `yaml:"panel" json:"panel"`
	Position string `yaml:"position" json:"position"`
	Size     string `yaml:"size" json:"size"`
}

// APISpec describes an API endpoint provided by the plugin.
type APISpec struct {
	Path   string `yaml:"path" json:"path"`
	Method string `yaml:"method" json:"method"`
}

// PluginSearchResult represents a plugin search result.
type PluginSearchResult struct {
	Name        string  `json:"name"`
	Version     string  `json:"version"`
	Description string  `json:"description"`
	Author      string  `json:"author"`
	Downloads   int     `json:"downloads"`
	Rating      float64 `json:"rating"`
}

// Registry manages the plugin registry.
type Registry struct {
	client   *http.Client
	registry string
	cache    map[string]*PluginSearchResult
	cacheTime time.Time
	mu       sync.RWMutex
}

// PluginInfo contains detailed plugin information.
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
	Path        string `json:"path"`
	Type        string `json:"type"`
}

// NewManager creates a new plugin manager.
func NewManager(cfg *Config, pf *platform.Info, opt *optimizer.Optimizer) (*Manager, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	m := &Manager{
		config:    cfg,
		platform:  pf,
		optimizer: opt,
		plugins:   make(map[string]*Plugin),
		registry:  NewRegistry(cfg.Registry),
	}

	// Load built-in plugins
	for _, name := range cfg.BuiltinPlugins {
		if err := m.loadBuiltin(name); err != nil {
			return nil, fmt.Errorf("builtin plugin %s: %w", name, err)
		}
	}

	return m, nil
}

// DefaultConfig returns default plugin configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		AllowUnsigned:   false,
		MaxPlugins:      0,
		Timeout:         types.Duration(30 * time.Second),
		EnableWASM:      true,
		EnableGoPlugins: true,
		Registry:        "https://registry.aio-stack.dev",
	}
}

// loadBuiltin loads a built-in plugin.
func (m *Manager) loadBuiltin(name string) error {
	// Built-in plugins are registered via init()
	return nil
}

// RegisterBuiltin registers a built-in plugin.
func (m *Manager) RegisterBuiltin(p *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[p.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", p.Name)
	}

	m.plugins[p.Name] = p
	return nil
}

// LoadAll loads all plugins from plugin directories.
func (m *Manager) LoadAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Scan plugin directories
	for _, dir := range m.config.PluginDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				pluginPath := filepath.Join(dir, entry.Name())
				if err := m.loadPlugin(ctx, pluginPath); err != nil {
					// Log error but continue
				}
			}
		}
	}

	return nil
}

func (m *Manager) loadPlugin(ctx context.Context, path string) error {
	// Read manifest
	manifestPath := filepath.Join(path, "plugin.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// Create plugin
	plugin := &Plugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		License:     manifest.License,
		Homepage:    manifest.Homepage,
		Repository:  manifest.Repository,
		Path:        path,
		Manifest:    &manifest,
		Enabled:     true,
	}

	// Load based on type
	if m.config.EnableWASM {
		wasmPath := filepath.Join(path, "plugin.wasm")
		if _, err := os.Stat(wasmPath); err == nil {
			plugin.Type = types.PluginTypeWASM
			return m.loadWASMPlugin(ctx, plugin)
		}
	}

	if m.config.EnableGoPlugins {
		goPluginPath := filepath.Join(path, "plugin.so")
		if _, err := os.Stat(goPluginPath); err == nil {
			plugin.Type = types.PluginTypeGo
			return m.loadGoPlugin(ctx, plugin)
		}
	}

	return fmt.Errorf("no valid plugin module found in %s", path)
}

func (m *Manager) loadWASMPlugin(ctx context.Context, plugin *Plugin) error {
	// TODO: Load WASM plugin using wasmer/wazero
	plugin.Loaded = true
	return nil
}

func (m *Manager) loadGoPlugin(ctx context.Context, plugin *Plugin) error {
	// TODO: Load Go plugin using hashicorp/go-plugin
	plugin.Loaded = true
	return nil
}

// List returns all loaded plugins.
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Get returns a plugin by name.
func (m *Manager) Get(name string) (*Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	return p, ok
}

// Install installs a plugin from the registry.
func (m *Manager) Install(ctx context.Context, name, version string) error {
	// TODO: Download and install plugin from registry
	return nil
}

// Uninstall uninstalls a plugin.
func (m *Manager) Uninstall(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.plugins[name]; !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	delete(m.plugins, name)
	return nil
}

// Enable enables a plugin.
func (m *Manager) Enable(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	p.Enabled = true
	return nil
}

// Disable disables a plugin.
func (m *Manager) Disable(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	p.Enabled = false
	return nil
}

// UpdateAll updates all installed plugins.
func (m *Manager) UpdateAll(ctx context.Context) error {
	return nil
}

// Search searches the plugin registry.
func (m *Manager) Search(ctx context.Context, query string, limit int) ([]PluginSearchResult, error) {
	return m.registry.Search(ctx, query, limit)
}

// Info returns detailed information about a plugin.
func (m *Manager) Info(ctx context.Context, name string) (*PluginInfo, error) {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if !ok {
		return m.registry.Info(ctx, name)
	}

	return &PluginInfo{
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		Enabled:     p.Enabled,
		Path:        p.Path,
		Type:        string(p.Type),
	}, nil
}

// UnloadAll unloads all plugins.
func (m *Manager) UnloadAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.plugins {
		m.plugins[name].Loaded = false
	}

	return nil
}

// NewRegistry creates a new plugin registry client.
func NewRegistry(registryURL string) *Registry {
	return &Registry{
		client:   &http.Client{Timeout: 30 * time.Second},
		registry: registryURL,
		cache:    make(map[string]*PluginSearchResult),
	}
}

// Search searches the registry.
func (r *Registry) Search(ctx context.Context, query string, limit int) ([]PluginSearchResult, error) {
	return []PluginSearchResult{}, nil
}

// Info gets plugin info from registry.
func (r *Registry) Info(ctx context.Context, name string) (*PluginInfo, error) {
	return nil, fmt.Errorf("not implemented")
}