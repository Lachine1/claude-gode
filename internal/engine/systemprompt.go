package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// buildSystemPrompt constructs the Claude Code system prompt.
func (e *QueryEngine) buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString("You are Claude Code, Anthropic's official CLI for Claude.\n\n")

	sb.WriteString(e.buildContextSection())
	sb.WriteString("\n")
	sb.WriteString(e.buildToolSection())
	sb.WriteString("\n")

	if e.config.CustomPrompt != "" {
		sb.WriteString(e.config.CustomPrompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString(e.buildPermissionSection())

	if e.config.AppendPrompt != "" {
		sb.WriteString("\n")
		sb.WriteString(e.config.AppendPrompt)
	}

	return sb.String()
}

func (e *QueryEngine) buildContextSection() string {
	var sb strings.Builder

	sb.WriteString("## Context\n\n")

	sb.WriteString(fmt.Sprintf("- **CWD**: %s\n", e.config.Cwd))
	sb.WriteString(fmt.Sprintf("- **OS**: %s/%s\n", runtime.GOOS, runtime.GOARCH))

	if gitStatus := e.getGitStatus(); gitStatus != "" {
		sb.WriteString(fmt.Sprintf("- **Git**: %s\n", gitStatus))
	}

	if username := os.Getenv("USER"); username != "" {
		sb.WriteString(fmt.Sprintf("- **User**: %s\n", username))
	}

	sb.WriteString(fmt.Sprintf("- **Shell**: %s\n", os.Getenv("SHELL")))

	return sb.String()
}

func (e *QueryEngine) getGitStatus() string {
	if e.config.Cwd == "" {
		return ""
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = e.config.Cwd
	output, err := cmd.Output()
	if err != nil {
		return "not a git repository"
	}

	gitRoot := strings.TrimSpace(string(output))

	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = e.config.Cwd
	output, err = cmd.Output()
	if err != nil {
		return fmt.Sprintf("git root: %s", gitRoot)
	}

	branch := strings.TrimSpace(string(output))

	cmd = exec.Command("git", "log", "-1", "--format=%h %s")
	cmd.Dir = e.config.Cwd
	output, err = cmd.Output()
	if err != nil {
		return fmt.Sprintf("git root: %s, branch: %s", gitRoot, branch)
	}

	commit := strings.TrimSpace(string(output))

	return fmt.Sprintf("root: %s, branch: %s, HEAD: %s", filepath.Base(gitRoot), branch, commit)
}

func (e *QueryEngine) buildToolSection() string {
	var sb strings.Builder

	sb.WriteString("## Available Tools\n\n")

	for _, tool := range e.config.Tools {
		sb.WriteString(fmt.Sprintf("### %s\n\n", tool.Name()))
		sb.WriteString(tool.Description())
		sb.WriteString("\n\n")

		schema := tool.JSONSchema()
		schemaJSON, err := json.MarshalIndent(schema, "", "  ")
		if err == nil {
			sb.WriteString("Input schema:\n")
			sb.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(schemaJSON)))
		}
	}

	return sb.String()
}

func (e *QueryEngine) buildPermissionSection() string {
	var sb strings.Builder

	sb.WriteString("## Permissions\n\n")
	sb.WriteString("You have access to the tools listed above. Use them to help the user accomplish their tasks.\n\n")
	sb.WriteString("When using the bash tool:\n")
	sb.WriteString("- Be careful with destructive commands\n")
	sb.WriteString("- Explain what commands will do before running them\n")
	sb.WriteString("- Use appropriate error handling\n\n")

	sb.WriteString("When editing files:\n")
	sb.WriteString("- Always read the file first to understand its contents\n")
	sb.WriteString("- Make minimal, focused changes\n")
	sb.WriteString("- Preserve existing code style and conventions\n")

	return sb.String()
}

// buildAPIMessages prepares messages for the API call.
func (e *QueryEngine) buildAPIMessages() []types.Message {
	messages := make([]types.Message, 0, len(e.messages))

	for _, msg := range e.messages {
		if msg.Role == types.RoleSystem {
			continue
		}

		content := make([]types.ContentBlock, 0, len(msg.Content))
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentTypeText:
				content = append(content, types.ContentBlock{
					Type: types.ContentTypeText,
					Text: block.Text,
				})
			case types.ContentTypeToolUse:
				if block.ToolUse != nil {
					content = append(content, types.ContentBlock{
						Type: types.ContentTypeToolUse,
						ToolUse: &types.ToolUseContent{
							ID:    block.ToolUse.ID,
							Name:  block.ToolUse.Name,
							Input: block.ToolUse.Input,
						},
					})
				}
			case types.ContentTypeToolResult:
				if block.ToolResult != nil {
					content = append(content, types.ContentBlock{
						Type: types.ContentTypeToolResult,
						ToolResult: &types.ToolResultContent{
							ToolUseID: block.ToolResult.ToolUseID,
							Content:   block.ToolResult.Content,
							IsError:   block.ToolResult.IsError,
						},
					})
				}
			case types.ContentTypeThinking:
				if block.Thinking != nil {
					content = append(content, types.ContentBlock{
						Type: types.ContentTypeThinking,
						Thinking: &types.ThinkingContent{
							Text: block.Thinking.Text,
						},
					})
				}
			}
		}

		messages = append(messages, types.Message{
			Role:    msg.Role,
			Content: content,
		})
	}

	return messages
}
