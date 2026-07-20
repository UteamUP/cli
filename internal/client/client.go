package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/uteamup/cli/internal/auth"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

const (
	maxUploadResponseBytes = 4 * 1024 * 1024
	maxUploadErrorBytes    = 64 * 1024
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

// ToolCallContent is one content item returned by an MCP tools/call result.
type ToolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolCallResult is the protocol result envelope returned by MCP tools/call.
type ToolCallResult struct {
	Content []ToolCallContent `json:"content"`
	IsError bool              `json:"isError,omitempty"`
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
	baseURL   string
	timeout   time.Duration
	insecure  bool
	retryOpts RetryOptions
	logger    *logging.Logger
	requestID int
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
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true} //nolint:gosec // user-requested
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

		// Send tenant context headers (required by backend)
		if token.TenantGUID != "" {
			req.Header.Set("X-Tenant-Guid", token.TenantGUID)
		}

		c.logger.Debug("POST %s/mcp tool=%s tenant=%d", c.baseURL, toolName, token.TenantID)

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

	if err != nil {
		return nil, err
	}
	return NormalizeToolResult(result)
}

// NormalizeToolResult unwraps the first JSON text item from an MCP tools/call
// response so CLI output matches direct REST output. It accepts both a plain
// tool result and a full JSON-RPC envelope because backends may deliver either
// shape over SSE.
func NormalizeToolResult(payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var rpcResponse JSONRPCResponse
	if json.Unmarshal(payload, &rpcResponse) == nil && rpcResponse.JSONRPC == "2.0" {
		if rpcResponse.Error != nil {
			return nil, clierrors.NewAPIError(
				rpcResponse.Error.Code,
				"JSON-RPC Error",
				rpcResponse.Error.Message,
			)
		}
		payload = rpcResponse.Result
	}

	var toolResult ToolCallResult
	if json.Unmarshal(payload, &toolResult) != nil || len(toolResult.Content) == 0 {
		return payload, nil
	}

	var firstText string
	for _, item := range toolResult.Content {
		if item.Type != "text" {
			continue
		}
		if firstText == "" {
			firstText = item.Text
		}
		if json.Valid([]byte(item.Text)) {
			if toolResult.IsError {
				return nil, fmt.Errorf("tool returned an error: %s", boundedToolError(item.Text))
			}
			return json.RawMessage(item.Text), nil
		}
	}

	if toolResult.IsError {
		return nil, fmt.Errorf("tool returned an error: %s", boundedToolError(firstText))
	}
	if firstText == "" {
		return payload, nil
	}
	encoded, err := json.Marshal(firstText)
	if err != nil {
		return nil, fmt.Errorf("encoding tool response: %w", err)
	}
	return encoded, nil
}

func boundedToolError(value string) string {
	const maxRunes = 2000
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > maxRunes {
		return string(runes[:maxRunes])
	}
	return string(runes)
}

// CallREST sends a direct REST API request (used with email/password login auth).
// This mirrors how the frontend's apiCall() works.
//
// extraHeaders are applied AFTER the standard auth/tenant/CSRF headers so a
// caller can attach things like `Idempotency-Key` that the backend reads via
// `[FromHeader]`. Values must already be valid HTTP header strings; callers
// are responsible for any encoding.
func (c *APIClient) CallREST(ctx context.Context, method, path string, params map[string]any, extraHeaders map[string]string, actionName string) (json.RawMessage, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, clierrors.NewAuthError("loading token", err)
	}
	if token == nil || !token.IsValid() {
		return nil, &clierrors.NotAuthenticatedError{}
	}

	// For GET/DELETE, encode params as query string; for POST/PUT, send as JSON body
	fullURL := c.baseURL + path
	var bodyReader io.Reader

	if method == "GET" || method == "DELETE" {
		query := buildQueryString(params, actionName)
		if query != "" {
			fullURL += "?" + query
		}
	} else {
		// Remove positional args (id) from body — already in the URL
		bodyParams := make(map[string]any)
		for k, v := range params {
			if k != "id" {
				bodyParams[k] = v
			}
		}
		if len(bodyParams) > 0 {
			bodyBytes, err := json.Marshal(bodyParams)
			if err != nil {
				return nil, fmt.Errorf("marshaling body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	var result json.RawMessage

	err = RetryWithBackoff(ctx, c.logger, fmt.Sprintf("REST %s %s", method, path), c.retryOpts, func() error {
		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		// CSRF guard on mutating endpoints requires a non-simple-header marker —
		// the backend rejects any POST/PUT/PATCH/DELETE without
		// "X-Requested-With: XMLHttpRequest" with HTTP 400. The frontend
		// apiCall() sets it unconditionally; mirror that here so the CLI's
		// REST path (used by domains like bugsandfeatures update-status) isn't
		// silently blocked.
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		// Send tenant context headers (same as frontend apiCall)
		if token.TenantID > 0 {
			req.Header.Set("X-Tenant-ID", fmt.Sprintf("%d", token.TenantID))
		}
		if token.TenantGUID != "" {
			req.Header.Set("X-Tenant-Guid", token.TenantGUID)
		}

		// Caller-supplied headers (e.g. `Idempotency-Key` from a `HeaderName`
		// flag) win over the defaults; standard auth/tenant headers above
		// remain in place because callers don't set them.
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}

		c.logger.Debug("%s %s tenant=%d", method, fullURL, token.TenantID)

		resp, err := c.httpClient().Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode >= 400 {
			return clierrors.NewAPIError(resp.StatusCode, resp.Status, string(respBody))
		}

		result = json.RawMessage(respBody)
		return nil
	})

	return result, err
}

// CallRESTUpload sends a multipart/form-data request with one local file as
// the body part named fileField, plus params on the query string (the body is
// owned by the multipart payload). It is the REST counterpart for endpoints
// that bind IFormFile — e.g. the stock CSV import. Auth, tenant, and CSRF
// headers mirror CallREST.
func (c *APIClient) CallRESTUpload(ctx context.Context, method, path, fileField, filePath string, params map[string]any, extraHeaders map[string]string, actionName string) (json.RawMessage, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, clierrors.NewAuthError("loading token", err)
	}
	if token == nil || !token.IsValid() {
		return nil, &clierrors.NotAuthenticatedError{}
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("building multipart body: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("building multipart body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("building multipart body: %w", err)
	}

	fullURL := c.baseURL + path
	if query := buildQueryString(params, actionName); query != "" {
		fullURL += "?" + query
	}

	var result json.RawMessage

	err = RetryWithBackoff(ctx, c.logger, fmt.Sprintf("REST %s %s", method, path), c.retryOpts, func() error {
		req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		if token.TenantID > 0 {
			req.Header.Set("X-Tenant-ID", fmt.Sprintf("%d", token.TenantID))
		}
		if token.TenantGUID != "" {
			req.Header.Set("X-Tenant-Guid", token.TenantGUID)
		}

		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}

		c.logger.Debug("%s %s multipart file=%s tenant=%d", method, fullURL, filepath.Base(filePath), token.TenantID)

		resp, err := c.httpClient().Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode >= 400 {
			return clierrors.NewAPIError(resp.StatusCode, resp.Status, string(respBody))
		}

		result = json.RawMessage(respBody)
		return nil
	})

	return result, err
}

// CallRESTUploadLimited uploads one local file without buffering the complete
// multipart payload in memory. The caller supplies the maximum accepted file
// size and the filename/content type exposed to the backend. Error bodies are
// deliberately not returned because provider failures may contain secrets or
// untrusted model output.
func (c *APIClient) CallRESTUploadLimited(
	ctx context.Context,
	method string,
	path string,
	fileField string,
	filePath string,
	uploadFileName string,
	contentType string,
	maxFileBytes int64,
	extraHeaders map[string]string,
) (json.RawMessage, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, clierrors.NewAuthError("loading token", err)
	}
	if token == nil || !token.IsValid() {
		return nil, &clierrors.NotAuthenticatedError{}
	}

	if err := validateUploadFile(filePath, maxFileBytes); err != nil {
		return nil, err
	}
	uploadFileName = safeUploadFileName(uploadFileName)
	if uploadFileName == "" {
		uploadFileName = "media.bin"
	}

	var result json.RawMessage
	err = RetryWithBackoff(ctx, c.logger, fmt.Sprintf("REST %s %s", method, path), c.retryOpts, func() error {
		body, multipartContentType, streamErr := openMultipartFileStream(
			ctx,
			fileField,
			filePath,
			uploadFileName,
			contentType,
			maxFileBytes,
		)
		if streamErr != nil {
			return streamErr
		}
		defer body.Close()

		req, reqErr := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if reqErr != nil {
			return fmt.Errorf("creating request: %w", reqErr)
		}
		req.Header.Set("Content-Type", multipartContentType)
		for key, value := range extraHeaders {
			req.Header.Set(key, value)
		}
		// Security-scoping headers always win over caller-supplied metadata.
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		if token.TenantGUID != "" {
			req.Header.Set("X-Tenant-Guid", token.TenantGUID)
		}

		c.logger.Debug("%s %s multipart upload", method, c.baseURL+path)
		resp, requestErr := c.httpClient().Do(req)
		if requestErr != nil {
			return fmt.Errorf("request failed: %w", requestErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxUploadErrorBytes))
			return clierrors.NewAPIError(resp.StatusCode, resp.Status, "request rejected")
		}

		responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxUploadResponseBytes+1))
		if readErr != nil {
			return fmt.Errorf("reading response: %w", readErr)
		}
		if len(responseBody) > maxUploadResponseBytes {
			return fmt.Errorf("response exceeds %d bytes", maxUploadResponseBytes)
		}
		result = json.RawMessage(responseBody)
		return nil
	})

	return result, err
}

func validateUploadFile(filePath string, maxFileBytes int64) error {
	if maxFileBytes <= 0 {
		return fmt.Errorf("maximum upload size must be positive")
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("reading upload file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("upload source must be a regular file")
	}
	if info.Size() == 0 {
		return fmt.Errorf("upload source is empty")
	}
	if info.Size() > maxFileBytes {
		return fmt.Errorf("upload source exceeds the %d byte limit", maxFileBytes)
	}
	return nil
}

func safeUploadFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, name)
	runes := []rune(name)
	if len(runes) > 200 {
		name = string(runes[:200])
	}
	return name
}

func openMultipartFileStream(
	ctx context.Context,
	fileField string,
	filePath string,
	uploadFileName string,
	contentType string,
	maxFileBytes int64,
) (io.ReadCloser, string, error) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	multipartContentType := multipartWriter.FormDataContentType()

	go func() {
		defer func() {
			_ = multipartWriter.Close()
			_ = writer.Close()
		}()

		file, err := os.Open(filePath)
		if err != nil {
			_ = writer.CloseWithError(fmt.Errorf("opening upload file: %w", err))
			return
		}
		defer file.Close()

		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fileField, uploadFileName))
		partHeader.Set("Content-Type", contentType)
		part, err := multipartWriter.CreatePart(partHeader)
		if err != nil {
			_ = writer.CloseWithError(fmt.Errorf("creating multipart file: %w", err))
			return
		}

		written, err := io.Copy(part, io.LimitReader(file, maxFileBytes+1))
		if err != nil {
			_ = writer.CloseWithError(fmt.Errorf("streaming upload file: %w", err))
			return
		}
		if written > maxFileBytes {
			_ = writer.CloseWithError(fmt.Errorf("upload source exceeds the %d byte limit", maxFileBytes))
			return
		}

		select {
		case <-ctx.Done():
			_ = writer.CloseWithError(ctx.Err())
		default:
		}
	}()

	return reader, multipartContentType, nil
}

// buildQueryString converts params to URL query string for GET requests.
func buildQueryString(params map[string]any, actionName string) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, 0, len(params))
	for k, v := range params {
		// Skip "id" — it's in the URL path for get/update/delete
		if k == "id" {
			continue
		}
		// Skip "query" for search — it goes as the search term
		if k == "query" && actionName == "search" {
			parts = append(parts, fmt.Sprintf("search=%v", v))
			continue
		}
		// Map camelCase CLI flag names to backend query params
		switch k {
		case "pageSize":
			parts = append(parts, fmt.Sprintf("pageSize=%v", v))
		default:
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return strings.Join(parts, "&")
}
