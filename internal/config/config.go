// Package config provides configuration loading and validation for AIO Stack.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

// Loader handles configuration loading from multiple sources.
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetEnvPrefix("AIO")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &Loader{v: v}
}

// Load loads configuration from file and environment.
func (l *Loader) Load(cfg interface{}, paths ...string) error {
	// Set defaults first
	l.setDefaults(l.v)

	// Load from files
	for _, path := range paths {
		if path == "" {
			continue
		}
		l.v.AddConfigPath(filepath.Dir(path))
		l.v.SetConfigName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
		if err := l.v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("config file %s: %w", path, err)
			}
		}
	}

	// Unmarshal
	if err := l.v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
}

// LoadFromBytes loads configuration from bytes.
func (l *Loader) LoadFromBytes(cfg interface{}, data []byte) error {
	l.v.SetConfigType("toml")
	if err := l.v.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	return l.v.Unmarshal(cfg)
}

// setDefaults sets default values.
func (l *Loader) setDefaults(v *viper.Viper) {
	// Platform defaults
	v.SetDefault("platform.auto_detect", true)
	v.SetDefault("platform.enable_simd", true)
	v.SetDefault("platform.enable_wasm", true)
	v.SetDefault("platform.wasm_runtime", "wazero")
	v.SetDefault("platform.max_procs", 0)
	v.SetDefault("platform.memory_limit", 0)
	v.SetDefault("platform.termux_mode", false)
	v.SetDefault("platform.windows_console_mode", "auto")

	// Plugin defaults
	v.SetDefault("plugin.enabled", true)
	v.SetDefault("plugin.allow_unsigned", false)
	v.SetDefault("plugin.max_plugins", 0)
	v.SetDefault("plugin.timeout", "30s")
	v.SetDefault("plugin.enable_wasm", true)
	v.SetDefault("plugin.enable_go_plugins", true)
	v.SetDefault("plugin.registry", "https://registry.aio-stack.dev")

	// AI defaults
	v.SetDefault("ai.enabled", true)
	v.SetDefault("ai.default_model", "phi-3-mini-4k-instruct-q4_k_m.gguf")
	v.SetDefault("ai.context_window", 4096)
	v.SetDefault("ai.max_tokens", 2048)
	v.SetDefault("ai.temperature", 0.7)
	v.SetDefault("ai.top_p", 0.9)
	v.SetDefault("ai.top_k", 40)
	v.SetDefault("ai.repeat_penalty", 1.1)
	v.SetDefault("ai.num_threads", 0)
	v.SetDefault("ai.use_gpu", false)
	v.SetDefault("ai.gpu_layers", -1)
	v.SetDefault("ai.enable_wasm", true)
	v.SetDefault("ai.model_cache_size", 2)

	// CodeIntel defaults
	v.SetDefault("code_intel.enabled", true)
	v.SetDefault("code_intel.parser_cache_size", 32)
	v.SetDefault("code_intel.tree_cache_size", 128)
	v.SetDefault("code_intel.enable_queries", true)
	v.SetDefault("code_intel.max_file_size", 1048576)
	v.SetDefault("code_intel.enable_edits", false)

	// CLI defaults
	v.SetDefault("cli.enable_tui", false)
	v.SetDefault("cli.enable_web_ui", false)
	v.SetDefault("cli.web_ui_port", 8080)
	v.SetDefault("cli.web_ui_host", "127.0.0.1")
	v.SetDefault("cli.config_file", "config.toml")
	v.SetDefault("cli.completion_shell", true)
	v.SetDefault("cli.color_mode", "auto")
	v.SetDefault("cli.log_level", "info")
	v.SetDefault("cli.log_format", "console")

	// State defaults
	v.SetDefault("state.type", "sqlite")
	v.SetDefault("state.enable_wal", true)
	v.SetDefault("state.busy_timeout", "5s")
	v.SetDefault("state.max_open_conns", 10)
	v.SetDefault("state.enable_migrations", true)

	// Optimizer defaults
	v.SetDefault("optimizer.enabled", true)
	v.SetDefault("optimizer.enable_pgo", true)
	v.SetDefault("optimizer.enable_simd", true)
	v.SetDefault("optimizer.enable_pooling", true)
	v.SetDefault("optimizer.enable_interning", true)
	v.SetDefault("optimizer.gc_percent", 100)
	v.SetDefault("optimizer.memory_ballast", 0)

	// Telemetry defaults
	v.SetDefault("telemetry.enabled", false)
	v.SetDefault("telemetry.sample_rate", 0.01)
	v.SetDefault("telemetry.environment", "production")
}

// Save saves configuration to a file.
func (l *Loader) Save(cfg interface{}, path string) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Validate validates a configuration struct using CUE.
func Validate(cfg interface{}) error {
	// TODO: Implement CUE validation
	return nil
}

// GetConfigPaths returns the default configuration search paths.
func GetConfigPaths(appName string) []string {
	var paths []string

	// Current directory
	paths = append(paths, ".")

	// User config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, appName))
	}

	// System config directory
	if configDir := os.Getenv("AIO_CONFIG_DIR"); configDir != "" {
		paths = append(paths, configDir)
	}

	// /etc on Unix
	if runtime.GOOS != "windows" {
		paths = append(paths, "/etc/"+appName)
	}

	return paths
}

// FindConfigFile finds the configuration file in standard locations.
func FindConfigFile(appName string) (string, error) {
	for _, path := range GetConfigPaths(appName) {
		for _, name := range []string{"config.toml", "config.yaml", "config.yml"} {
			fullPath := filepath.Join(path, name)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, nil
			}
		}
	}
	return "", fmt.Errorf("config file not found for %s", appName)
}