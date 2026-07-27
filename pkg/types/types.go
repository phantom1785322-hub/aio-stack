// Package types provides shared types for the AIO Stack.
package types

import (
	"strings"
	"time"
)

// Duration is a wrapper around time.Duration that supports TOML/JSON unmarshaling.
type Duration time.Duration

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns the string representation.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// PluginType represents the type of plugin.
type PluginType string

const (
	// PluginTypeWASM is a WebAssembly plugin.
	PluginTypeWASM PluginType = "wasm"

	// PluginTypeGo is a Go plugin (hashicorp/go-plugin).
	PluginTypeGo PluginType = "go"

	// PluginTypeBuiltin is a built-in plugin.
	PluginTypeBuiltin PluginType = "builtin"
)

// PluginStatus represents the status of a plugin.
type PluginStatus string

const (
	// PluginStatusUnloaded means the plugin is not loaded.
	PluginStatusUnloaded PluginStatus = "unloaded"

	// PluginStatusLoading means the plugin is being loaded.
	PluginStatusLoading PluginStatus = "loading"

	// PluginStatusLoaded means the plugin is loaded and ready.
	PluginStatusLoaded PluginStatus = "loaded"

	// PluginStatusError means the plugin failed to load.
	PluginStatusError PluginStatus = "error"

	// PluginStatusDisabled means the plugin is disabled.
	PluginStatusDisabled PluginStatus = "disabled"
)

// AIModelType represents the type of AI model.
type AIModelType string

const (
	// AIModelTypeGGUF is a GGUF model (llama.cpp).
	AIModelTypeGGUF AIModelType = "gguf"

	// AIModelTypeONNX is an ONNX model.
	AIModelTypeONNX AIModelType = "onnx"

	// AIModelTypeWASM is a WASM-compiled model.
	AIModelTypeWASM AIModelType = "wasm"
)

// Language represents a programming language.
type Language string

const (
	LanguageGo          Language = "go"
	LanguageTypeScript  Language = "typescript"
	LanguageJavaScript  Language = "javascript"
	LanguagePython      Language = "python"
	LanguageRust        Language = "rust"
	LanguageJava        Language = "java"
	LanguageC           Language = "c"
	LanguageCpp         Language = "cpp"
	LanguageCSharp      Language = "csharp"
	LanguageRuby        Language = "ruby"
	LanguagePHP         Language = "php"
	LanguageSwift       Language = "swift"
	LanguageKotlin      Language = "kotlin"
	LanguageScala       Language = "scala"
	LanguageShell       Language = "shell"
	LanguageDockerfile  Language = "dockerfile"
	LanguageYAML        Language = "yaml"
	LanguageJSON        Language = "json"
	LanguageTOML        Language = "toml"
	LanguageMarkdown    Language = "markdown"
	LanguageHTML        Language = "html"
	LanguageCSS         Language = "css"
	LanguageSQL         Language = "sql"
	LanguageGraphQL     Language = "graphql"
	LanguageProtobuf    Language = "protobuf"
)

// AllLanguages returns all supported languages.
func AllLanguages() []Language {
	return []Language{
		LanguageGo, LanguageTypeScript, LanguageJavaScript, LanguagePython,
		LanguageRust, LanguageJava, LanguageC, LanguageCpp, LanguageCSharp,
		LanguageRuby, LanguagePHP, LanguageSwift, LanguageKotlin, LanguageScala,
		LanguageShell, LanguageDockerfile, LanguageYAML, LanguageJSON,
		LanguageTOML, LanguageMarkdown, LanguageHTML, LanguageCSS,
		LanguageSQL, LanguageGraphQL, LanguageProtobuf,
	}
}

// IsValidLanguage checks if a language is supported.
func IsValidLanguage(lang Language) bool {
	for _, l := range AllLanguages() {
		if l == lang {
			return true
		}
	}
	return false
}

// ParseLanguage parses a language string.
func ParseLanguage(s string) (Language, bool) {
	lang := Language(strings.ToLower(s))
	return lang, IsValidLanguage(lang)
}

// Severity represents the severity of a check or issue.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
	SeverityCritical Severity = "critical"
)

// CheckCategory represents the category of a check.
type CheckCategory string

const (
	CheckCategoryShell       CheckCategory = "shell"
	CheckCategoryGit         CheckCategory = "git"
	CheckCategoryRuntime     CheckCategory = "runtime"
	CheckCategoryContainer   CheckCategory = "container"
	CheckCategoryKubernetes  CheckCategory = "kubernetes"
	CheckCategoryCloud       CheckCategory = "cloud"
	CheckCategorySSH         CheckCategory = "ssh"
	CheckCategoryIDE         CheckCategory = "ide"
	CheckCategoryDatabase    CheckCategory = "database"
	CheckCategorySecrets     CheckCategory = "secrets"
	CheckCategoryFilesystem  CheckCategory = "filesystem"
	CheckCategoryNetwork     CheckCategory = "network"
	CheckCategoryProcesses   CheckCategory = "processes"
	CheckCategorySecurity    CheckCategory = "security"
	CheckCategoryPerformance CheckCategory = "performance"
)

// FixType represents the type of fix.
type FixType string

const (
	FixTypeAuto      FixType = "auto"
	FixTypeInteractive FixType = "interactive"
	FixTypeManual    FixType = "manual"
	FixTypeScript    FixType = "script"
)

// OutputFormat represents the output format.
type OutputFormat string

const (
	OutputFormatHuman   OutputFormat = "human"
	OutputFormatJSON    OutputFormat = "json"
	OutputFormatJUnit   OutputFormat = "junit"
	OutputFormatSARIF   OutputFormat = "sarif"
	OutputFormatPrometheus OutputFormat = "prometheus"
	OutputFormatHTML    OutputFormat = "html"
)

// Platform represents a target platform.
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformDarwin  Platform = "darwin"
	PlatformFreeBSD Platform = "freebsd"
	PlatformOpenBSD Platform = "openbsd"
	PlatformNetBSD  Platform = "netbsd"
	PlatformAndroid Platform = "android"
	PlatformTermux  Platform = "termux"
	PlatformWSL     Platform = "wsl"
)

// Architecture represents a CPU architecture.
type Architecture string

const (
	ArchAMD64   Architecture = "amd64"
	ArchARM64   Architecture = "arm64"
	ArchARM     Architecture = "arm"
	ArchARMv7   Architecture = "armv7"
	Arch386     Architecture = "386"
	ArchRISCV64 Architecture = "riscv64"
	ArchPPC64le Architecture = "ppc64le"
	ArchS390x   Architecture = "s390x"
)

// String returns the string representation of the platform.
func (p Platform) String() string {
	return string(p)
}

// String returns the string representation of the architecture.
func (a Architecture) String() string {
	return string(a)
}

// AllPlatforms returns all supported platforms.
func AllPlatforms() []Platform {
	return []Platform{
		PlatformLinux, PlatformWindows, PlatformDarwin,
		PlatformFreeBSD, PlatformOpenBSD, PlatformNetBSD,
		PlatformAndroid, PlatformTermux, PlatformWSL,
	}
}

// AllArchitectures returns all supported architectures.
func AllArchitectures() []Architecture {
	return []Architecture{
		ArchAMD64, ArchARM64, ArchARM, ArchARMv7,
		Arch386, ArchRISCV64, ArchPPC64le, ArchS390x,
	}
}

// IsValidPlatform checks if a platform is supported.
func IsValidPlatform(p Platform) bool {
	for _, platform := range AllPlatforms() {
		if platform == p {
			return true
		}
	}
	return false
}

// IsValidArchitecture checks if an architecture is supported.
func IsValidArchitecture(a Architecture) bool {
	for _, arch := range AllArchitectures() {
		if arch == a {
			return true
		}
	}
	return false
}