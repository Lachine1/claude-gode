package tools

import (
	"github.com/Lachine1/claude-gode/internal/tools/bash"
	"github.com/Lachine1/claude-gode/internal/tools/edit"
	"github.com/Lachine1/claude-gode/internal/tools/glob"
	"github.com/Lachine1/claude-gode/internal/tools/grep"
	"github.com/Lachine1/claude-gode/internal/tools/ls"
	"github.com/Lachine1/claude-gode/internal/tools/read"
	"github.com/Lachine1/claude-gode/internal/tools/webfetch"
	"github.com/Lachine1/claude-gode/internal/tools/write"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// RegisterTools returns all available tools
func RegisterTools() []types.Tool {
	return []types.Tool{
		bash.New(),
		read.New(),
		write.New(),
		edit.New(),
		glob.New(),
		grep.New(),
		ls.New(),
		webfetch.New(),
	}
}

// GetTool returns a tool by name
func GetTool(name string, tools []types.Tool) types.Tool {
	for _, t := range tools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}
