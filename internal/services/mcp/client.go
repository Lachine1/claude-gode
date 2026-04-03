package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// MCPResource represents a resource provided by an MCP server
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
}

// ToolResult represents the result of an MCP tool call
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"is_error,omitempty"`
}

// ToolResultContent represents content within a tool result
type ToolResultContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// MCPServer manages a connection to an MCP server
type MCPServer struct {
	Name         string
	Config       MCPConfig
	Connected    bool
	Tools        []types.Tool
	Resources    []MCPResource
	capabilities map[string]interface{}

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reqID  int
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

// Connect starts the MCP server process and initializes the connection
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

	var err error
	s.stdin, err = cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	s.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	s.cmd = cmd

	if err := s.initialize(ctx); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to initialize MCP server: %w", err)
	}

	if err := s.loadTools(ctx); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to load tools: %w", err)
	}

	if err := s.loadResources(ctx); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to load resources: %w", err)
	}

	s.Connected = true
	return nil
}

// Disconnect stops the MCP server process
func (s *MCPServer) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Connected {
		return nil
	}

	if s.stdin != nil {
		s.stdin.Close()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}

	s.Connected = false
	s.Tools = make([]types.Tool, 0)
	s.Resources = make([]MCPResource, 0)
	return nil
}

// CallTool calls a tool on the MCP server
func (s *MCPServer) CallTool(ctx context.Context, name string, input map[string]interface{}) (*ToolResult, error) {
	s.mu.Lock()
	if !s.Connected {
		s.mu.Unlock()
		return nil, fmt.Errorf("server not connected")
	}
	s.mu.Unlock()

	params := map[string]interface{}{
		"name":      name,
		"arguments": input,
	}

	result, err := s.sendRequest(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	var toolResult ToolResult
	if err := json.Unmarshal(result, &toolResult); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w", err)
	}

	return &toolResult, nil
}

// ListTools returns the tools available from this MCP server
func (s *MCPServer) ListTools() []types.Tool {
	return s.Tools
}

// ListResources returns the resources available from this MCP server
func (s *MCPServer) ListResources() []MCPResource {
	return s.Resources
}

// initialize sends the initialize request to the MCP server
func (s *MCPServer) initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "claude-gode",
			"version": "0.1.0",
		},
	}

	result, err := s.sendRequest(ctx, "initialize", params)
	if err != nil {
		return err
	}

	var initResult struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		Capabilities    map[string]interface{} `json:"capabilities"`
		ServerInfo      map[string]interface{} `json:"serverInfo"`
	}
	if err := json.Unmarshal(result, &initResult); err != nil {
		return fmt.Errorf("failed to parse initialize response: %w", err)
	}

	s.capabilities = initResult.Capabilities
	if name, ok := initResult.ServerInfo["name"].(string); ok {
		s.Name = name
	}

	_ = s.sendNotification("notifications/initialized", map[string]interface{}{})
	return nil
}

// loadTools fetches the list of tools from the MCP server
func (s *MCPServer) loadTools(ctx context.Context) error {
	result, err := s.sendRequest(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		return err
	}

	var toolsResp struct {
		Tools []struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			InputSchema map[string]interface{} `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(result, &toolsResp); err != nil {
		return fmt.Errorf("failed to parse tools response: %w", err)
	}

	s.Tools = make([]types.Tool, 0, len(toolsResp.Tools))
	for _, t := range toolsResp.Tools {
		s.Tools = append(s.Tools, &mcpTool{
			name:        t.Name,
			description: t.Description,
			inputSchema: t.InputSchema,
			server:      s,
		})
	}

	return nil
}

// loadResources fetches the list of resources from the MCP server
func (s *MCPServer) loadResources(ctx context.Context) error {
	result, err := s.sendRequest(ctx, "resources/list", map[string]interface{}{})
	if err != nil {
		return nil
	}

	var resourcesResp struct {
		Resources []MCPResource `json:"resources"`
	}
	if err := json.Unmarshal(result, &resourcesResp); err != nil {
		return nil
	}

	s.Resources = resourcesResp.Resources
	return nil
}

// sendRequest sends a JSON-RPC request and returns the result
func (s *MCPServer) sendRequest(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	s.mu.Lock()
	s.reqID++
	reqID := s.reqID
	s.mu.Unlock()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      reqID,
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data = append(data, '\n')
	if _, err := s.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	decoder := json.NewDecoder(s.stdout)
	for {
		var resp map[string]interface{}
		if err := decoder.Decode(&resp); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		id, ok := resp["id"]
		if !ok {
			continue
		}

		if float64(reqID) == id {
			if errVal, ok := resp["error"]; ok {
				errData, _ := json.Marshal(errVal)
				return nil, fmt.Errorf("JSON-RPC error: %s", string(errData))
			}

			resultData, _ := json.Marshal(resp["result"])
			return resultData, nil
		}
	}
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (s *MCPServer) sendNotification(method string, params interface{}) error {
	notif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	data = append(data, '\n')
	_, err = s.stdin.Write(data)
	return err
}

// mcpTool wraps an MCP tool to implement types.Tool
type mcpTool struct {
	name        string
	description string
	inputSchema map[string]interface{}
	server      *MCPServer
}

func (t *mcpTool) Name() string {
	return t.name
}

func (t *mcpTool) Description() string {
	return t.description
}

func (t *mcpTool) JSONSchema() map[string]interface{} {
	return t.inputSchema
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
		if c.Type == "text" {
			contentText += c.Text
		}
	}

	data, _ := json.Marshal(contentText)
	return &types.ToolResult[json.RawMessage]{
		Data:    data,
		IsError: result.IsError,
	}, nil
}
