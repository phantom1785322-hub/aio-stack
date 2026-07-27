// Package cli provides the CLI framework for AIO Stack applications.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aio-stack/aio-stack/pkg/types"
	"github.com/alecthomas/kong"
	"github.com/spf13/cobra"
)

// Framework is the CLI framework that supports both Kong and Cobra.
type Framework struct {
	config *Config
	app    AppInterface
	kong   *kong.Kong
	cobra  *cobra.Command
}

// AppInterface defines the interface the CLI framework needs from the application.
type AppInterface interface {
	Name() string
	Version() string
	Config() *Config
	Platform() string
	PluginManager() PluginManagerInterface
	AIEngine() AIEngineInterface
}

// Config holds CLI configuration.
type Config struct {
	EnableTUI       bool   `toml:"enable_tui" json:"enable_tui"`
	EnableWebUI     bool   `toml:"enable_web_ui" json:"enable_web_ui"`
	WebUIPort       int    `toml:"web_ui_port" json:"web_ui_port"`
	WebUIHost       string `toml:"web_ui_host" json:"web_ui_host"`
	ConfigFile      string `toml:"config_file" json:"config_file"`
	CompletionShell bool   `toml:"completion_shell" json:"completion_shell"`
	ColorMode       string `toml:"color_mode" json:"color_mode"`
	LogLevel        string `toml:"log_level" json:"log_level"`
	LogFormat       string `toml:"log_format" json:"log_format"`
}

// PluginManagerInterface defines the plugin manager interface.
type PluginManagerInterface interface {
	List() []PluginInfo
	Install(ctx context.Context, name, version string) error
	Uninstall(ctx context.Context, name string) error
	Enable(ctx context.Context, name string) error
	Disable(ctx context.Context, name string) error
	UpdateAll(ctx context.Context) error
	Search(ctx context.Context, query string, limit int) ([]PluginSearchResult, error)
	Info(ctx context.Context, name string) (*PluginInfo, error)
}

// PluginInfo represents plugin metadata.
type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Enabled     bool
	Path        string
	Type        types.PluginType
}

// PluginSearchResult represents a plugin search result.
type PluginSearchResult struct {
	Name        string
	Version     string
	Description string
	Author      string
}

// AIEngineInterface defines the AI engine interface.
type AIEngineInterface interface {
	GenerateCompletion(ctx context.Context, req *CompletionRequest) (string, error)
	ApplyPromptTemplate(templateName string, vars map[string]interface{}) (string, error)
	GetPromptTemplate(name string) (string, bool)
	AddPromptTemplate(name, template string)
	SetSystemPrompt(prompt string)
	ListModels() []*Model
	LoadModel(ctx context.Context, name string) (*Model, error)
	UnloadModel(name string) error
}

// CompletionRequest represents a completion request.
type CompletionRequest struct {
	Prompt      string
	MaxTokens   int
	Temperature float32
	Model       string
}

// Model represents an AI model.
type Model struct {
	Name        string
	Path        string
	Type        types.AIModelType
	ContextSize int
	Loaded      bool
	Metadata    map[string]interface{}
}

// NewFramework creates a new CLI framework.
func NewFramework(cfg *Config, app AppInterface) (*Framework, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	f := &Framework{
		config: cfg,
		app:    app,
	}

	// Initialize Kong parser
	if err := f.initKong(); err != nil {
		return nil, fmt.Errorf("kong init: %w", err)
	}

	// Initialize Cobra for compatibility
	if err := f.initCobra(); err != nil {
		return nil, fmt.Errorf("cobra init: %w", err)
	}

	return f, nil
}

// DefaultConfig returns default CLI configuration.
func DefaultConfig() *Config {
	return &Config{
		EnableTUI:       false,
		EnableWebUI:     false,
		WebUIPort:       8080,
		WebUIHost:       "127.0.0.1",
		ConfigFile:      "config.toml",
		CompletionShell: true,
		ColorMode:       "auto",
		LogLevel:        "info",
		LogFormat:       "console",
	}
}

// initKong initializes the Kong command parser.
func (f *Framework) initKong() error {
	cli := &CLI{
		framework: f,
	}

	var err error
	f.kong, err = kong.New(cli,
		kong.Name(f.app.Name()),
		kong.Description("AIO Stack Application"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": f.app.Version(),
		},
	)
	return err
}

// initCobra initializes the Cobra command parser (for shell completions).
func (f *Framework) initCobra() error {
	f.cobra = &cobra.Command{
		Use:   f.app.Name(),
		Short: "AIO Stack Application",
		Long:  "AIO Stack Application",
	}

	// Add completion command
	if f.config.CompletionShell {
		f.cobra.AddCommand(&cobra.Command{
			Use:   "completion [shell]",
			Short: "Generate shell completion script",
			Args:  cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				shell := "bash"
				if len(args) > 0 {
					shell = args[0]
				}
				return f.generateCompletion(shell)
			},
		})
	}

	return nil
}

// CLI is the root Kong CLI structure.
type CLI struct {
	framework *Framework

	Version VersionCmd `cmd:"" help:"Show version information."`
	Config  ConfigCmd  `cmd:"" help:"Manage configuration."`
	Plugin  PluginCmd  `cmd:"" help:"Manage plugins."`
	Doctor  DoctorCmd  `cmd:"" help:"Run system health diagnostics."`
	AI      AICmd      `cmd:"" help:"AI assistant commands."`
}

// VersionCmd shows version information.
type VersionCmd struct {
	framework *Framework
}

func (c *VersionCmd) Run(ctx *kong.Context) error {
	fmt.Printf("%s version %s\n", c.framework.app.Name(), c.framework.app.Version())
	fmt.Printf("Platform: %s\n", c.framework.app.Platform())
	return nil
}

// ConfigCmd manages configuration.
type ConfigCmd struct {
	framework *Framework

	Get  ConfigGetCmd  `cmd:"" help:"Get a configuration value."`
	Set  ConfigSetCmd  `cmd:"" help:"Set a configuration value."`
	List ConfigListCmd `cmd:"" help:"List all configuration values."`
	Edit ConfigEditCmd `cmd:"" help:"Open config file in editor."`
	Paths ConfigPathsCmd `cmd:"" help:"Show configuration file paths."`
}

type ConfigGetCmd struct {
	framework *Framework
	Key string `arg:"" help:"Configuration key to get."`
}

func (c *ConfigGetCmd) Run(ctx *kong.Context) error {
	fmt.Printf("Getting config: %s\n", c.Key)
	return nil
}

type ConfigSetCmd struct {
	framework *Framework
	Key   string `arg:"" help:"Configuration key to set."`
	Value string `arg:"" help:"Value to set."`
}

func (c *ConfigSetCmd) Run(ctx *kong.Context) error {
	fmt.Printf("Setting config: %s = %s\n", c.Key, c.Value)
	return nil
}

type ConfigListCmd struct {
	framework *Framework
}

func (c *ConfigListCmd) Run(ctx *kong.Context) error {
	fmt.Println("Listing all configuration...")
	return nil
}

type ConfigEditCmd struct {
	framework *Framework
}

func (c *ConfigEditCmd) Run(ctx *kong.Context) error {
	fmt.Println("Opening config file in editor...")
	return nil
}

type ConfigPathsCmd struct {
	framework *Framework
}

func (c *ConfigPathsCmd) Run(ctx *kong.Context) error {
	fmt.Println("Configuration paths:")
	return nil
}

// PluginCmd manages plugins.
type PluginCmd struct {
	framework *Framework

	List      PluginListCmd      `cmd:"" help:"List installed plugins."`
	Install   PluginInstallCmd   `cmd:"" help:"Install a plugin."`
	Uninstall PluginUninstallCmd `cmd:"" help:"Uninstall a plugin."`
	Enable    PluginEnableCmd    `cmd:"" help:"Enable a plugin."`
	Disable   PluginDisableCmd   `cmd:"" help:"Disable a plugin."`
	Update    PluginUpdateCmd    `cmd:"" help:"Update plugins."`
	Search    PluginSearchCmd    `cmd:"" help:"Search plugin registry."`
	Info      PluginInfoCmd      `cmd:"" help:"Show plugin information."`
}

type PluginListCmd struct {
	framework *Framework
	All       bool `short:"a" help:"Show all plugins (including disabled)."`
	JSON      bool `short:"j" help:"Output as JSON."`
}

func (c *PluginListCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		fmt.Println("Plugin system not enabled")
		return nil
	}

	plugins := pm.List()
	if c.JSON {
		// TODO: JSON output
	}
	for _, p := range plugins {
		status := "enabled"
		if !p.Enabled {
			status = "disabled"
		}
		fmt.Printf("%s (%s) - %s\n", p.Name, p.Version, status)
	}
	return nil
}

type PluginInstallCmd struct {
	framework *Framework
	Name      string `arg:"" help:"Plugin name or URL to install."`
	Version   string `short:"v" help:"Specific version to install."`
	Force     bool   `short:"f" help:"Force reinstall if already installed."`
}

func (c *PluginInstallCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	fmt.Printf("Installing plugin: %s\n", c.Name)
	return pm.Install(context.Background(), c.Name, c.Version)
}

type PluginUninstallCmd struct {
	framework *Framework
	Name      string `arg:"" help:"Plugin name to uninstall."`
}

func (c *PluginUninstallCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	fmt.Printf("Uninstalling plugin: %s\n", c.Name)
	return pm.Uninstall(context.Background(), c.Name)
}

type PluginEnableCmd struct {
	framework *Framework
	Name      string `arg:"" help:"Plugin name to enable."`
}

func (c *PluginEnableCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	return pm.Enable(context.Background(), c.Name)
}

type PluginDisableCmd struct {
	framework *Framework
	Name      string `arg:"" help:"Plugin name to disable."`
}

func (c *PluginDisableCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	return pm.Disable(context.Background(), c.Name)
}

type PluginUpdateCmd struct {
	framework *Framework
	All       bool `short:"a" help:"Update all plugins."`
}

func (c *PluginUpdateCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	fmt.Println("Updating plugins...")
	return pm.UpdateAll(context.Background())
}

type PluginSearchCmd struct {
	framework *Framework
	Query     string `arg:"" help:"Search query."`
	Limit     int    `short:"l" default:"10" help:"Maximum results."`
}

func (c *PluginSearchCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	results, err := pm.Search(context.Background(), c.Query, c.Limit)
	if err != nil {
		return err
	}
	for _, r := range results {
		fmt.Printf("%s - %s\n", r.Name, r.Description)
	}
	return nil
}

type PluginInfoCmd struct {
	framework *Framework
	Name      string `arg:"" help:"Plugin name."`
}

func (c *PluginInfoCmd) Run(ctx *kong.Context) error {
	pm := c.framework.app.PluginManager()
	if pm == nil {
		return fmt.Errorf("plugin system not enabled")
	}
	info, err := pm.Info(context.Background(), c.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Name: %s\nVersion: %s\nDescription: %s\nAuthor: %s\n", info.Name, info.Version, info.Description, info.Author)
	return nil
}

// DoctorCmd runs system health diagnostics.
type DoctorCmd struct {
	Fix   bool   `short:"f" help:"Attempt to auto-fix issues."`
	JSON  bool   `short:"j" help:"Output as JSON."`
	Check string `short:"c" help:"Run specific check only."`
}

func (c *DoctorCmd) Run(ctx *kong.Context) error {
	fmt.Println("Running system health diagnostics...")
	return nil
}

// AICmd provides AI assistant commands.
type AICmd struct {
	framework *Framework

	Chat      AIChatCmd      `cmd:"" help:"Start an interactive chat session."`
	Explain   AIExplainCmd   `cmd:"" help:"Explain code or error."`
	Commit    AICommitCmd    `cmd:"" help:"Generate commit message."`
	Review    AIReviewCmd    `cmd:"" help:"Review code for issues."`
	Changelog AIChangelogCmd `cmd:"" help:"Generate changelog."`
	Estimate  AIEstimateCmd  `cmd:"" help:"Estimate story points."`
	Breakdown AIBreakdownCmd `cmd:"" help:"Break down epic into tasks."`
}

type AIChatCmd struct {
	framework *Framework
	Model     string `short:"m" help:"Model to use."`
}

func (c *AIChatCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Starting AI chat session... (type 'exit' to quit)")
	return nil
}

type AIExplainCmd struct {
	framework *Framework
	File     string `short:"f" help:"File to explain."`
	Code     string `short:"c" help:"Code snippet to explain."`
	Language string `short:"l" help:"Programming language."`
	Simple   bool   `short:"s" help:"Simple explanation (non-technical)."`
}

func (c *AIExplainCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}

	var code string
	if c.File != "" {
		data, err := os.ReadFile(c.File)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		code = string(data)
	} else if c.Code != "" {
		code = c.Code
	} else {
		return fmt.Errorf("provide --file or --code")
	}

	lang := c.Language
	if lang == "" && c.File != "" {
		lang = detectLanguageFromFile(c.File)
	}

	vars := map[string]interface{}{
		"code":     code,
		"language": lang,
	}
	template := "code_explanation"
	if c.Simple {
		template = "simple_explanation"
	}

	prompt, err := ai.ApplyPromptTemplate(template, vars)
	if err != nil {
		return err
	}

	resp, err := ai.GenerateCompletion(context.Background(), &CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   2048,
		Temperature: 0.3,
	})
	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

type AICommitCmd struct {
	framework *Framework
	Staged    bool `short:"s" help:"Use staged changes only."`
	All       bool `short:"a" help:"Include all changes."`
}

func (c *AICommitCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Generating commit message...")
	return nil
}

type AIReviewCmd struct {
	framework *Framework
	File string `short:"f" help:"File to review."`
	Diff string `short:"d" help:"Diff to review."`
}

func (c *AIReviewCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Reviewing code...")
	return nil
}

type AIChangelogCmd struct {
	framework *Framework
	Since  string `short:"s" help:"Generate changelog since tag/commit."`
	Format string `short:"f" default:"markdown" help:"Output format (markdown, json)."`
}

func (c *AIChangelogCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Generating changelog...")
	return nil
}

type AIEstimateCmd struct {
	framework *Framework
	Task string `arg:"" help:"Task description to estimate."`
}

func (c *AIEstimateCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Estimating...")
	return nil
}

type AIBreakdownCmd struct {
	framework *Framework
	Epic  string `arg:"" help:"Epic description to break down."`
	Count int    `short:"c" default:"5" help:"Number of tasks to generate."`
}

func (c *AIBreakdownCmd) Run(ctx *kong.Context) error {
	ai := c.framework.app.AIEngine()
	if ai == nil {
		return fmt.Errorf("AI engine not enabled")
	}
	fmt.Println("Breaking down epic...")
	return nil
}

// Run runs the CLI framework.
func (f *Framework) Run(ctx context.Context) error {
	_, err := f.kong.Parse(os.Args[1:])
	if err != nil {
		// Fall back to Cobra for shell completion
		if strings.HasPrefix(err.Error(), "unknown command") {
			return f.cobra.Execute()
		}
		return err
	}
	return nil
}

// Shutdown shuts down the CLI framework.
func (f *Framework) Shutdown(ctx context.Context) error {
	return nil
}

// generateCompletion generates shell completion scripts.
func (f *Framework) generateCompletion(shell string) error {
	switch shell {
	case "bash":
		return f.cobra.GenBashCompletion(os.Stdout)
	case "zsh":
		return f.cobra.GenZshCompletion(os.Stdout)
	case "fish":
		return f.cobra.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return f.cobra.GenPowerShellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

// detectLanguageFromFile detects programming language from file extension.
func detectLanguageFromFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
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
		return "text"
	}
}