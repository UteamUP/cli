package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/auth"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// ToolCallParams represents the params for a tools/call request.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// APIClient communicates with the UteamUP backend.
type APIClient struct {
	baseURL    string
	timeout    time.Duration
	insecure   bool
	retryOpts  RetryOptions
	logger     *logging.Logger
	requestID  int
}

// NewAPIClient creates a new APIClient.
func NewAPIClient(baseURL string, timeout time.Duration, insecure bool, retryOpts RetryOptions, logger *logging.Logger) *APIClient {
	return &APIClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		timeout:   timeout,
		insecure:  insecure,
		retryOpts: retryOpts,
		logger:    logger,
	}
}

func (c *APIClient) httpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested
	}
	return &http.Client{Transport: transport, Timeout: c.timeout}
}

// CallTool sends a JSON-RPC 2.0 tools/call request to the /mcp endpoint.
func (c *APIClient) CallTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, clierrors.NewAuthError("loading token", err)
	}
	if token == nil || !token.IsValid() {
		return nil, &clierrors.NotAuthenticatedError{}
	}

	c.requestID++
	rpcReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.requestID,
		Method:  "tools/call",
		Params:  ToolCallParams{Name: toolName, Arguments: args},
	}

	body, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	var result json.RawMessage

	err = RetryWithBackoff(ctx, c.logger, fmt.Sprintf("tool/%s", toolName), c.retryOpts, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		c.logger.Debug("POST %s/mcp tool=%s", c.baseURL, toolName)

		resp, err := c.httpClient().Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(resp.Body)
			return clierrors.NewAPIError(resp.StatusCode, resp.Status, string(respBody))
		}

		// Handle SSE or plain JSON
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/event-stream") {
			events, err := ParseSSE(resp.Body)
			if err != nil {
				return fmt.Errorf("parsing SSE: %w", err)
			}
			result = ExtractResult(events)
		} else {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			var rpcResp JSONRPCResponse
			if err := json.Unmarshal(respBody, &rpcResp); err != nil {
				// If not valid JSON-RPC, return raw body
				result = json.RawMessage(respBody)
				return nil
			}

			if rpcResp.Error != nil {
				return clierrors.NewAPIError(rpcResp.Error.Code, "JSON-RPC Error", rpcResp.Error.Message)
			}
			result = rpcResp.Result
		}

		return nil
	})

	return result, err
}
