# Claude Code in Go

[![Go](https://img.shields.io/badge/Go-%E2%89%A51.24.0-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![BubbleTea](https://img.shields.io/badge/TUI-BubbleTea_v2-FF6B6B)](https://charm.sh)
[![License](https://img.shields.io/github/license/Lachine1/claude-gode)](LICENSE)

A faithful Go recreation of the Claude Code CLI — reconstructed from the TypeScript source maps of the public npm release.

> [!WARNING]
> This repository is an **unofficial** Go recreation, built by studying the reconstructed TypeScript source of the public npm release. It is **for research and learning purposes only** and does not represent the internal development repository of Anthropic.

---

## Requirements

- Go ≥ 1.24.0
- Bubble Tea v2 ecosystem (Bubble Tea, Lip Gloss, Bubbles)

## Quickstart

```bash
$ go build -o claude-gode ./cmd/claude-gode   # Build the binary
$ ./claude-gode                                # Start CLI (Interactive)
$ ./claude-gode --version                      # Verify version
```

---

## Project Structure

```text
.
├── cmd/claude-gode/          # CLI entry point
├── internal/
│   ├── bootstrap/            # Startup sequence, init, profiling
│   ├── cli/                  # CLI transports & command handlers
│   ├── engine/               # Core LLM API loop (QueryEngine)
│   ├── tools/                # 53 tool implementations
│   │   ├── bash/             # Bash tool
│   │   ├── read/             # Read file tool
│   │   ├── edit/             # Edit file tool
│   │   ├── webfetch/         # Web fetch tool
│   │   ├── agent/            # Agent/sub-agent tool
│   │   └── ...               # (remaining tools)
│   ├── commands/             # 87 slash commands
│   ├── services/             # Backend services (API, MCP, auth)
│   ├── tui/                  # Terminal UI components (406 components)
│   │   ├── components/       # Messages, inputs, diffs, dialogs
│   │   ├── layout/           # Screen layouts & virtual scrolling
│   │   └── styles/           # Lip Gloss theme definitions
│   ├── coordinator/          # Multi-agent orchestration
│   ├── tasks/                # Task types (local, remote, agent, dream)
│   ├── memdir/               # Persistent memory system (5-tier)
│   ├── vim/                  # Vim keybinding engine
│   ├── skills/               # Skill system & bundled skills
│   ├── plugins/              # Plugin system
│   ├── bridge/               # IDE bidirectional communication
│   ├── remote/               # Remote session teleportation
│   ├── voice/                # Voice interaction (streaming STT)
│   ├── buddy/                # Gacha companion sprite (Easter egg)
│   └── assistant/            # KAIROS daemon mode
├── pkg/                      # Public packages
│   ├── types/                # Shared types & interfaces
│   ├── schemas/              # API schemas (JSON Schema / Zod equiv)
│   └── utils/                # Shared utilities
├── shims/                    # Compatibility shims (if needed)
├── vendor/                   # Vendored native binding sources
├── go.mod
├── go.sum
├── README.md
└── CONTRIBUTING.md
```

---

## Architecture

Claude Code is built on top of a highly optimized and robust architecture designed for LLM API interaction, token efficiency, and advanced execution boundaries.

### Boot Sequence

```text
cmd/claude-gode/main.go → bootstrap/ → cli/ → REPL (Bubble Tea)
  │                         │            │
  │                         │            └─ Full Init: Auth → Feature Flags → MCP → Settings → Commands
  │                         └─ Fast Path: --version / daemon / ps / logs
  └─ Startup Gate: validates all dependencies; blocks boot until resolved
```

### Core Engine & Token Optimization

Token efficiency is critical. The architecture employs industry-leading token saving techniques:

- **`engine/QueryEngine`**: The central engine managing the LLM API loop, session lifecycle, and automatic tool execution.
- **3-Tier Compaction System**:
  1. **Microcompact**: Removes messages from the server cache without invalidating the prompt cache context (zero API cost).
  2. **Session Memory**: Uses pre-extracted session memory as a summary to avoid LLM calls during mid-level compaction.
  3. **Full Compact**: Instructs a sub-agent to summarize the conversation into a structured 9-section format, employing `<analysis>` tag stripping to reduce token usage while maintaining quality.
- **Advanced Optimizations**:
  - `FILE_UNCHANGED_STUB`: Returns a brief 30-word stub for re-read files.
  - Dynamic max output caps (8K default with 64K retry) preventing slot-reservation waste.
  - Caching latches to prevent UI toggles from busting 70K context.
  - Circuit breakers preventing wasted API calls on consecutive compaction failures.

### Harness Engineering (Permissions & Security)

The "Harness" safely controls LLM operations within the local environment:

- **Permission Modes**: 6 primary modes (`acceptEdits`, `bypassPermissions`, `default`, `dontAsk`, `plan`) plus internal designations like `auto` (yoloClassifier) and `bubble` (sub-agent propagation).
- **Security Checkers**: PowerShell-specific security analysis to detect command injection, download cradles, and privilege escalation, plus redundant path validations.

### Teams & Multi-Agent Orchestration

- **Agents**: Orchestrated via `AgentTool`, created with three distinct paths: Teammate (tmux or in-process), Fork (inheriting context), and Normal (fresh context).
- **Coordinator Mode**: A designated coordinator delegates exact coding tasks to worker agents (`Agent`, `SendMessage`, `TaskStop`), effectively isolating high-level reasoning from raw file execution.

### Memory System (5-Tier Architecture)

Designed to persist AI knowledge across sessions and agents:

1. **Memdir**: Project-level indices and topic files (`MEMORY.md`).
2. **Auto Extract**: Fire-and-forget forked agent that consolidates memory post-session.
3. **Session Memory**: Real-time context tracking without extra LLM overhead.
4. **Team Memory**: Shared remote state leveraging SHA-256 delta uploads and git-leaks-based secret extraction guards.
5. **Agent Memory**: Agent-specific knowledge scoped to local, project, or user levels.

### UI Architecture

Terminal UI based on **Bubble Tea v2 + Lip Gloss + Bubbles**:

- `tui/components/` (~406 components) — Messages, inputs, diffs, permission dialogs, status bar
- `tui/layout/` — Screen layouts, virtual scrolling, focus management
- `tui/styles/` — Lip Gloss theme definitions and styling utilities

---

## Reconstruction Notes

This is a ground-up Go recreation, not a direct translation. The following may differ:

| Type | Description |
|------|-------------|
| **UI Framework** | React/Ink replaced with Bubble Tea v2 + Lip Gloss + Bubbles |
| **Type System** | TypeScript interfaces replaced with Go structs and interfaces |
| **Concurrency** | Promise/async replaced with goroutines and channels |
| **Native bindings** | Replaced with Go equivalents or CGO where needed |
| **Dynamic resources** | Embedded via Go `embed` package |

---

## Disclaimer

- The original source code copyright belongs to [Anthropic](https://www.anthropic.com).
- This is for technical research and learning purposes only. Please do not use it for commercial purposes.
