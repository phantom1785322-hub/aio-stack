# CHANGELOG.md

All notable changes to aio-stack will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Comprehensive GitHub issue templates (bug, feature, question)
- Pull request template with checklist
- CODEOWNERS file
- Dependabot configuration for Go modules and GitHub Actions
- GitHub Actions workflows: CI, Release, Security
- Security scanning (govulncheck, gosec, CodeQL)
- SECURITY.md with vulnerability disclosure policy

### Changed
- Updated README.md with professional formatting, badges, installation methods, usage, architecture, cross-platform support
- Improved workflow files with multi-platform builds, security scanning

### Security
- Added govulncheck, gosec, CodeQL to CI pipeline

---

## [0.1.0] - 2025-07-27

### Added
- **Platform Detection**: OS, Arch, CPU features (AVX2/NEON/SVE), SIMD width, container/VM/WSL/Termux detection
- **Optimizer**: Object pooling, string interning, SIMD-aware batching, GC tuning, memory ballast
- **Code Intelligence**: Tree-sitter parsers for 15+ languages (Go, TS, Python, Rust, etc.)
- **AI Engine**: Local llama.cpp via WASM (Phi-3, CodeLlama, DeepSeek models)
- **Plugin System**: WASM (wasmer/wazero) + Go plugins, signed registry, hot-reload, capability-based security
- **CLI Framework**: Cobra + Kong dual parser, shell completions, man pages, markdown docs
- **Config System**: TOML + CUE validation, Viper, env overrides, cross-platform config dirs
- **State Management**: SQLite + etcd, migrations, encrypted secrets, sync engine
- **Cross-platform**: Builds for Linux/macOS/Windows/Termux/BSD (11 targets)
- **Package Managers**: Homebrew tap, Scoop bucket
- **Documentation**: Comprehensive README with features, usage, config, themes, plugins, architecture, cross-platform matrix, performance benchmarks
- **License**: MIT License

### Technical
- Built on Go 1.23 with zero runtime dependencies (single binary)
- Tree-sitter for incremental parsing
- llama.cpp via WASM for local AI inference
- wasmer/wazero for WASM plugin runtime
- Cobra/Kong for CLI framework
- Viper for config management

---

[Unreleased]: https://github.com/aio-stack/aio-stack/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/aio-stack/aio-stack/releases/tag/v0.1.0