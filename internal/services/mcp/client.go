package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/Lachine1/claude-gode/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPResource represents a resource provided by an MCP server
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
}

// MCPServer manages a connection to an MCP server using the official Go SDK
type MCPServer struct {
	Name      string
	Config    MCPConfig
	Connected bool
	Tools     []types.Tool
	Resources []MCPResource

	mu      sync.Mutex
	client  *mcp.Client
	session *mcp.ClientSession
	cmd     *exec.Cmd
}

// MCPConfig holds configuration for starting an MCP server
type MCPConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(config MCPConfig) *MCPServer {
	return &MCPServer{
		Config:    config,
		Tools:     make([]types.Tool, 0),
		Resources: make([]MCPResource, 0),
	}
}

// Connect starts the MCP server process and initializes the connection using the official SDK
func (s *MCPServer) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Connected {
		return nil
	}

	cmd := exec.CommandContext(ctx, s.Config.Command, s.Config.Args...)

	env := cmd.Environ()
	for k, v := range s.Config.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	transport := &mcp.CommandTransport{Command: cmd}

	s.client = mcp.NewClient(&mcp.Implementation{
		Name:    "claude-gode",
		Version: "999.0.0-restored",
	}, nil)

	session, err := s.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	s.session = session
	s.cmd = cmd

	// Get server info
	if err := session.Ping(ctx, nil); err != nil {
		// ping is optional, just log it
		_ = err
	}

	if err := s.loadTools(ctx); err != nil {
		session.Close()
		return fmt.Errorf("failed to load tools: %w", err)
	}

	if err := s.loadResources(ctx); err != nil {
		session.Close()
		return fmt.Errorf("failed to load resources: %w", err)
	}

	s.Connected = true
	return nil
}

// Disconnect closes the MCP session
func (s *MCPServer) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Connected {
		return nil
	}

	if s.session != nil {
		s.session.Close()
	}

	s.Connected = false
	s.Tools = make([]types.Tool, 0)
	s.Resources = make([]MCPResource, 0)
	return nil
}

// CallTool calls a tool on the MCP server using the official SDK
func (s *MCPServer) CallTool(ctx context.Context, name string, input map[string]interface{}) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	if !s.Connected || s.session == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("server not connected")
	}
	session := s.session
	s.mu.Unlock()

	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: input,
	}

	return session.CallTool(ctx, params)
}

// ListTools returns the tools available from this MCP server
func (s *MCPServer) ListTools() []types.Tool {
	return s.Tools
}

// ListResources returns the resources available from this MCP server
func (s *MCPServer) ListResources() []MCPResource {
	return s.Resources
}

// loadTools fetches the list of tools from the MCP server using the official SDK
func (s *MCPServer) loadTools(ctx context.Context) error {
	tools, err := s.session.ListTools(ctx, nil)
	if err != nil {
		return err
	}

	s.Tools = make([]types.Tool, 0, len(tools.Tools))
	for _, t := range tools.Tools {
		schema := t.InputSchema
		s.Tools = append(s.Tools, &mcpTool{
			name:        t.Name,
			description: t.Description,
			inputSchema: schema,
			server:      s,
		})
	}

	return nil
}

// loadResources fetches the list of resources from the MCP server using the official SDK
func (s *MCPServer) loadResources(ctx context.Context) error {
	resources, err := s.session.ListResources(ctx, nil)
	if err != nil {
		return nil // resources are optional
	}

	s.Resources = make([]MCPResource, 0, len(resources.Resources))
	for _, r := range resources.Resources {
		s.Resources = append(s.Resources, MCPResource{
			URI:         r.URI,
			Name:        r.Name,
			Description: r.Description,
			MimeType:    r.MIMEType,
		})
	}

	return nil
}

// mcpTool wraps an MCP tool to implement types.Tool
type mcpTool struct {
	name        string
	description string
	inputSchema any
	server      *MCPServer
}

func (t *mcpTool) Name() string {
	return t.name
}

func (t *mcpTool) Description() string {
	return t.description
}

func (t *mcpTool) JSONSchema() map[string]interface{} {
	if m, ok := t.inputSchema.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func (t *mcpTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var args map[string]interface{}
	if err := json.Unmarshal(input, &args); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to parse input: %v", err),
		}, nil
	}

	abortCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-ctx.AbortSignal:
			cancel()
		case <-abortCtx.Done():
		}
	}()

	result, err := t.server.CallTool(abortCtx, t.name, args)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	var contentText string
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			contentText += tc.Text
		}
	}

	data, _ := json.Marshal(contentText)
	return &types.ToolResult[json.RawMessage]{
		Data:    data,
		IsError: result.IsError,
	}, nil
}
