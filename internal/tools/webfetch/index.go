package webfetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

const defaultMaxLength = 10000

var (
	scriptRe  = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRe   = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	tagRe     = regexp.MustCompile(`(?is)<[^>]+>`)
	commentRe = regexp.MustCompile(`(?is)<!--.*?-->`)
	headRe    = regexp.MustCompile(`(?is)<head[^>]*>.*?</head>`)
	titleRe   = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	brRe      = regexp.MustCompile(`(?is)<br\s*/?>`)
	pRe       = regexp.MustCompile(`(?is)</?p[^>]*>`)
	hRe       = regexp.MustCompile(`(?is)</?h[1-6][^>]*>`)
	liRe      = regexp.MustCompile(`(?is)<li[^>]*>`)
	aRe       = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)
)

// WebFetchTool fetches web content
type WebFetchTool struct {
	client *http.Client
}

// New creates a new WebFetchTool
func New() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the tool name
func (t *WebFetchTool) Name() string {
	return "webfetch"
}

// Description returns the tool description
func (t *WebFetchTool) Description() string {
	return "Fetch content from a URL and convert it to readable text. Supports HTML to text conversion."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *WebFetchTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to fetch",
			},
			"max_length": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of characters to return (default: 10000)",
			},
		},
		"required": []string{"url"},
	}
}

type webFetchInput struct {
	URL       string `json:"url"`
	MaxLength *int   `json:"max_length,omitempty"`
}

// Execute fetches the URL and converts to text
func (t *WebFetchTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params webFetchInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("invalid input: %v", err),
		}, nil
	}

	if params.URL == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "url cannot be empty",
		}, nil
	}

	// Validate URL
	parsedURL, err := url.Parse(params.URL)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("invalid URL: %v", err),
		}, nil
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "only http and https URLs are supported",
		}, nil
	}

	// Check abort signal
	select {
	case <-ctx.AbortSignal:
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "execution aborted",
		}, nil
	default:
	}

	// Fetch the URL
	req, err := http.NewRequest("GET", params.URL, nil)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	req.Header.Set("User-Agent", "claude-gode/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to fetch URL: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Read the response
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}

	contentType := resp.Header.Get("Content-Type")
	var text string

	if strings.Contains(contentType, "html") {
		text = htmlToText(string(body))
	} else {
		text = string(body)
	}

	// Extract title
	title := extractTitle(string(body))

	// Apply length limit
	maxLen := defaultMaxLength
	if params.MaxLength != nil && *params.MaxLength > 0 {
		maxLen = *params.MaxLength
	}

	truncated := false
	if len(text) > maxLen {
		text = text[:maxLen]
		truncated = true
	}

	result := map[string]interface{}{
		"url":       params.URL,
		"title":     title,
		"content":   text,
		"truncated": truncated,
		"status":    resp.StatusCode,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}

func htmlToText(html string) string {
	// Remove comments
	text := commentRe.ReplaceAllString(html, "")

	// Remove head section
	text = headRe.ReplaceAllString(text, "")

	// Extract title before removing tags
	title := extractTitle(html)

	// Convert <br> and block elements to newlines
	text = brRe.ReplaceAllString(text, "\n")
	text = pRe.ReplaceAllString(text, "\n\n")
	text = hRe.ReplaceAllString(text, "\n\n")
	text = liRe.ReplaceAllString(text, "\n• ")

	// Convert links to [text](url) format
	text = aRe.ReplaceAllString(text, "[$2]($1)")

	// Remove scripts and styles
	text = scriptRe.ReplaceAllString(text, "")
	text = styleRe.ReplaceAllString(text, "")

	// Remove all remaining HTML tags
	text = tagRe.ReplaceAllString(text, "")

	// Decode HTML entities
	text = decodeHTMLEntities(text)

	// Clean up whitespace
	text = strings.TrimSpace(text)

	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	// Add title at the top
	if title != "" {
		text = fmt.Sprintf("# %s\n\n%s", title, text)
	}

	return text
}

func extractTitle(html string) string {
	matches := titleRe.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(decodeHTMLEntities(matches[1]))
	}
	return ""
}

func decodeHTMLEntities(text string) string {
	replacements := map[string]string{
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&#39;":    "'",
		"&nbsp;":   " ",
		"&#160;":   " ",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&hellip;": "…",
		"&laquo;":  "«",
		"&raquo;":  "»",
	}

	for entity, char := range replacements {
		text = strings.ReplaceAll(text, entity, char)
	}

	return text
}
