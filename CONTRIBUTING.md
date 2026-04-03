# Contributing

Thank you for your interest in contributing to Claude Code in Go! This project is a ground-up Go recreation of the Claude Code CLI, built for research and learning purposes.

## Development Setup

### Prerequisites

- Go ≥ 1.24.0
- `golangci-lint` (for linting)
- `goimports` (for import formatting)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/<your-username>/claude-gode.git
cd claude-gode

# Install dependencies
go mod download

# Build the binary
go build -o claude-gode ./cmd/claude-gode

# Run the CLI
./claude-gode

# Run tests
go test ./...

# Run linter
golangci-lint run
```

## Project Structure

This project follows standard Go project layout conventions:

- **`cmd/`** — CLI entry points
- **`internal/`** — Private application code (not importable by other projects)
- **`pkg/`** — Public library code (importable by other projects)

Each subsystem is organized as its own package within `internal/`. See [README.md](README.md) for the full architecture overview.

## Code Style

### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `go fmt ./...` before committing
- Run `goimports -w .` to manage imports
- Run `golangci-lint run` before submitting a PR

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | `lowercase`, no underscores | `queryengine`, not `query_engine` |
| Types | `PascalCase` | `QueryEngine`, `ToolResult` |
| Interfaces | `PascalCase`, `-er` suffix for single-method | `Reader`, `Executor` |
| Functions/Methods | `PascalCase` (exported), `camelCase` (unexported) | `NewQueryEngine`, `handleMessage` |
| Variables | `camelCase` | `sessionID`, `maxTokens` |
| Constants | `PascalCase` or `camelCase` | `DefaultMaxTokens`, `version` |
| Errors | Prefixed with `Err` | `ErrUnauthorized`, `ErrSessionNotFound` |

### Imports

Group imports in three sections, separated by blank lines:

1. Standard library
2. Third-party
3. Local project imports

```go
import (
    "context"
    "fmt"

    "github.com/charmbracelet/bubbletea/v2"
    "github.com/charmbracelet/lipgloss/v2"

    "github.com/lachine/claude-gode/internal/types"
    "github.com/lachine/claude-gode/pkg/utils"
)
```

### Error Handling

- Return errors, don't panic (except in `init()` or truly unrecoverable situations)
- Wrap errors with context using `fmt.Errorf("doing X: %w", err)`
- Define sentinel errors in the package where they're used: `var ErrNotFound = errors.New("not found")`

### Concurrency

- Use goroutines and channels over sync primitives when possible
- Document goroutine lifecycles — who starts them, who stops them, how
- Use `context.Context` for cancellation and timeouts
- Prefer `errgroup` for managing groups of goroutines

### TUI Components

Components follow the Bubble Tea v2 model:

```go
type Model struct {
    // state fields
}

func New() Model {
    return Model{}
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    return m, nil
}

func (m Model) View() string {
    return ""
}
```

## Testing

### Test Organization

- Place tests next to the code they test: `foo.go` → `foo_test.go`
- Use table-driven tests for functions with multiple cases
- Name test functions descriptively: `TestQueryEngine_CompactsOnTokenLimit`

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/engine/...

# Run with race detection
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Making Changes

### Small Changes (bug fixes, minor improvements)

1. Create a branch: `git checkout -b fix/something-broken`
2. Make your changes
3. Run tests and linter
4. Commit with a conventional commit message
5. Open a PR

### Large Changes (new features, refactors)

1. **Open an issue first** — describe what you want to do and why
2. Wait for discussion and alignment
3. Implement in small, reviewable PRs
4. Each PR should be independently functional (no broken builds between PRs)

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat(engine): add 3-tier compaction system
fix(tools/bash): handle SIGTERM propagation correctly
docs(readme): update architecture diagram
refactor(tui): extract permission dialog into separate component
```

## Architecture Decisions

When making significant changes, document the reasoning:

- **Why** this approach?
- **What** alternatives were considered?
- **How** does this fit with existing architecture?

For major decisions, open a discussion issue before implementing.

## Fidelity to Original

This project aims for close-to-1:1 functionality with the original Claude Code. When implementing features:

1. Study the TypeScript source in `references/claude-code/` to understand behavior
2. Replicate the behavior in Go, adapting to idiomatic Go patterns
3. Document any deviations from the original and why

Deviations are expected and acceptable where:
- Go idioms differ significantly from TypeScript
- The Bubble Tea ecosystem has different capabilities than React/Ink
- Performance characteristics allow for better approaches

## What to Work On

See [GitHub Issues](https://github.com/lachine/claude-gode/issues) for open tasks. Good first contributions:

- Tool implementations (each tool is self-contained)
- TUI components (visual, testable, good for learning Bubble Tea)
- Slash commands (straightforward, well-defined behavior)
- Utility functions and helpers

## Communication

- **Issues** — bug reports, feature requests, discussions
- **Pull Requests** — code changes (keep them small and focused)
- **Discussions** — architecture, design, direction questions

## License

This project is for research and learning purposes only. See [README.md](README.md) for details.
