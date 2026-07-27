# Contributing to aio-stack

We ❤️ contributors! Thank you for taking the time to contribute to aio-stack.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Commit Messages](#commit-messages)
- [Release Process](#release-process)
- [Community](#community)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code. Please report unacceptable behavior to security@aio-stack.dev.

---

## Getting Started

### Prerequisites
- **Go 1.23+** (for core development)
- **Git** (obviously!)
- **Make** or **Task** (for build commands)

### Quick Start
```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/aio-stack
cd aio-stack

# 3. Add upstream remote
git remote add upstream https://github.com/aio-stack/aio-stack

# 4. Install dev tools
make dev-setup

# 5. Build and test
make build
make test
```

---

## Development Setup

### Install Dev Tools
```bash
make dev-setup
# Installs: golangci-lint, goimports, goreleaser, govulncheck
```

### Run Tests
```bash
# All tests
make test

# Short tests only
make test-short

# With race detector
make test-race

# With coverage
make coverage
```

### Run Linter
```bash
make lint
```

### Format Code
```bash
make fmt
```

### Build Binary
```bash
# Current platform
make build

# All platforms
make build-all
```

### Run CLI (Development)
```bash
make dev
# or
./bin/aio-stack doctor
```

---

## How to Contribute

### Reporting Bugs
1. Search [existing issues](https://github.com/aio-stack/aio-stack/issues) first
2. Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.yml)
3. Include `aio-stack doctor` output and reproduction steps

### Suggesting Features
1. Search [existing issues](https://github.com/aio-stack/aio-stack/issues) and [discussions](https://github.com/aio-stack/aio-stack/discussions)
2. Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.yml)
3. Explain the problem and proposed solution

### Improving Documentation
- Fix typos, clarify explanations, add examples
- Update README, CHANGELOG, or docs/ files
- Documentation changes don't need tests

### Code Contributions

#### Good First Issues
Look for labels: `good first issue`, `help wanted`, `documentation`

#### Before You Start
1. **Comment on the issue** — Let others know you're working on it
2. **Check for existing PRs** — Avoid duplicate work
3. **Discuss large changes** — Open an issue first for major features

#### Making Changes
1. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Make your changes with clear, focused commits
3. Follow coding standards (see below)
4. Add tests for new functionality
4. Update documentation
5. Run `make check` before pushing

---

## Pull Request Process

### Before Submitting
- [ ] All CI checks pass (`make check`)
- [ ] Tests added/updated for new functionality
- [ ] Documentation updated
- [ ] Changelog entry added (for user-facing changes)
- [ ] Conventional commit messages used
- [ ] No merge conflicts with `main`

### PR Requirements
- **Clear title** following conventional commits
- **Description** explaining what and why
- **Related issues** linked (`Fixes #123`)
- **Tests** for new functionality
- **Screenshots** for UI changes

### Review Process
1. Automated CI runs (tests, lint, build, security)
2. Code owner review (auto-requested via CODEOWNERS)
3. Address review feedback
4. Squash and merge (maintainers handle)

### After Merge
- Delete your feature branch
- Pull latest `main` to your fork
- Celebrate! 🎉

---

## Coding Standards

### Go (Core)
- **Standard library style** — Follow [Effective Go](https://go.dev/doc/effective_go)
- **Format** — `gofmt` / `goimports` (run `make fmt`)
- **Lint** — Pass `golangci-lint` (run `make lint`)
- **Error handling** — Check errors explicitly, wrap with context
- **Testing** — Table-driven tests, `testify` for assertions
- **Comments** — Export comments for public APIs

### Project Structure
```
aio-stack/
├── cmd/aio-stack/           # CLI entry point
├── internal/                # Private packages
│   ├── platform/            # Platform detection
│   ├── optimizer/           # Optimizer (pools, interning, GC tuning)
│   ├── codeintel/           # Code intelligence (Tree-sitter)
│   ├── ai/                  # AI engine (llama.cpp via WASM)
│   ├── plugin/              # Plugin system (WASM + Go)
│   ├── cli/                 # CLI framework (Cobra + Kong)
│   ├── config/              # Config system (TOML + CUE)
│   └── state/               # State management (SQLite + etcd)
├── .github/                 # GitHub Actions, templates
├── docs/                    # Documentation
└── scripts/                 # Build/release scripts
```

### Naming Conventions
- **Packages**: lowercase, single word (`platform`, `optimizer`, `codeintel`)
- **Files**: lowercase with underscores (`platform.go`, `config.go`)
- **Types**: PascalCase (`Platform`, `Optimizer`, `CodeIntel`)
- **Functions**: PascalCase for public, camelCase for private
- **Constants**: PascalCase or UPPER_SNAKE_CASE
- **Variables**: camelCase

### Error Handling
```go
// Good: wrap with context
if err != nil {
    return fmt.Errorf("failed to open repo at %s: %w", path, err)
}

// Good: sentinel errors for expected cases
var ErrNotFound = errors.New("not found")
```

### Testing
```go
// Table-driven tests
func TestParseDiff(t *testing.T) {
    tests := []struct{
        name     string
        input    string
        expected Diff
    }{
        {"added file", "...", Diff{...}},
        {"deleted file", "...", Diff{...}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Coverage Targets
- **Core packages** (`internal/platform`, `internal/optimizer`): >80%
- **CLI commands**: >70%
- **Plugin system**: >60% (harder to test)
- **Overall**: >70%

---

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, missing semicolons, etc. |
| `refactor` | Code change that neither fixes nor adds feature |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `chore` | Maintenance, deps, build |
| `ci` | CI/CD changes |
| `build` | Build system changes |

### Examples
```
feat(platform): add RISC-V V extension detection

fix(optimizer): handle zero-size allocations in object pool

docs(readme): add Windows Scoop installation instructions

refactor(ai): extract prompt templates to separate package

perf(optimizer): use SIMD-aware batching for string interning

test(platform): add tests for ARM64 NEON detection

chore(deps): update bubbletea to v0.25.0
```

### Scopes (optional)
- `platform`, `optimizer`, `codeintel`, `ai`, `plugin`, `cli`, `config`, `build`, `ci`, `docs`, `deps`

---

## Release Process

### Versioning
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** — Breaking changes
- **MINOR** — New features (backward compatible)
- **PATCH** — Bug fixes (backward compatible)

### Release Steps (Maintainers)
```bash
# 1. Ensure main is up to date
git checkout main && git pull

# 2. Create release tag
git tag v0.2.0
git push origin v0.2.0

# 3. GitHub Actions handles the rest:
# - Runs full CI
# - Builds 11 platform binaries
# - Creates GitHub Release with assets
# - Generates changelog
# - Updates Homebrew formula
# - Updates Scoop manifest
# - Posts to Discord, Twitter, Dev.to, Mastodon
```

### Changelog
Follow [Keep a Changelog](https://keepachangelog.com/):
```markdown
## [0.2.0] - 2025-08-15
### Added
- New `aio-stack gpu` command for GPU detection
- Plugin marketplace integration

### Fixed
- ARM64 NEON detection on Apple Silicon
- Windows ARM64 build issues

### Changed
- Updated to Go 1.23
- Improved optimizer performance by 40%
```