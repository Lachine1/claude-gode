package bash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// BashInput represents the input parameters for the Bash tool
type BashInput struct {
	Command string `json:"command"`
	Timeout *int   `json:"timeout,omitempty"`
}

// BashOutput represents the output from the Bash tool
type BashOutput struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// BashTool implements the types.Tool interface for shell command execution
type BashTool struct{}

// New creates a new BashTool
func New() *BashTool {
	return &BashTool{}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return "Execute a shell command. Supports any command that the system shell can run. Use with caution."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *BashTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Optional timeout in seconds (default: 120)",
			},
		},
		"required": []string{"command"},
	}
}

// Execute runs the shell command
func (t *BashTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params BashInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid input: %v", err),
		}, nil
	}

	if params.Command == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "Command cannot be empty",
		}, nil
	}

	timeout := 120 * time.Second
	if params.Timeout != nil && *params.Timeout > 0 {
		timeout = time.Duration(*params.Timeout) * time.Second
	}

	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmdCtx, cmdCancel := context.WithCancel(execCtx)
	defer cmdCancel()

	go func() {
		select {
		case <-ctx.AbortSignal:
			cmdCancel()
		case <-cmdCtx.Done():
		}
	}()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", params.Command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = ctx.Cwd

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("Command execution failed: %v", err),
			}, nil
		}
	}

	output := BashOutput{
		ExitCode: exitCode,
		Stdout:   strings.TrimSuffix(stdout.String(), "\n"),
		Stderr:   strings.TrimSuffix(stderr.String(), "\n"),
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}
