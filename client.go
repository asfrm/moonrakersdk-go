package moonraker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client is the main Moonraker API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	mu         sync.RWMutex
	connected  bool
	lastInfo   *ServerInfo
}

// ClientOption is a functional option for configuring the client.
type ClientOption func(*Client)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the default request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Moonraker API client.
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// IsConnected returns whether the client believes it's connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetServerInfo fetches and caches server info.
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var info ServerInfo
	if err := c.doRequest(ctx, http.MethodGet, "/server/info", nil, &info); err != nil {
		return nil, err
	}

	c.connected = info.KlippyConnected
	c.lastInfo = &info
	return &info, nil
}

// GetLastInfo returns the last cached server info.
func (c *Client) GetLastInfo() *ServerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastInfo
}

// doRequest performs an HTTP request with proper error handling and response parsing.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	var contentType string

	if body != nil {
		if buf, ok := body.(*bytes.Buffer); ok {
			reqBody = buf
		} else {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
			contentType = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if result != nil {
		var wrapper struct {
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(respBody, &wrapper); err != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return nil
		}

		if len(wrapper.Result) > 0 {
			if err := json.Unmarshal(wrapper.Result, result); err != nil {
				return fmt.Errorf("failed to parse result: %w", err)
			}
		}
	}

	return nil
}

// doRequestForm performs a form-data request (for uploads).
func (c *Client) doRequestForm(ctx context.Context, method, path string, body *bytes.Buffer, contentType string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if result != nil {
		var wrapper struct {
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(respBody, &wrapper); err != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return nil
		}

		if len(wrapper.Result) > 0 {
			if err := json.Unmarshal(wrapper.Result, result); err != nil {
				return fmt.Errorf("failed to parse result: %w", err)
			}
		}
	}

	return nil
}

// QueryParams represents URL query parameters.
type QueryParams map[string]interface{}

// EncodeQueryParams encodes query parameters with type hints.
func EncodeQueryParams(params QueryParams) string {
	v := url.Values{}
	for key, val := range params {
		switch t := val.(type) {
		case string:
			v.Set(key, t)
		case int:
			v.Set(key, fmt.Sprintf("%d", t))
		case int64:
			v.Set(key, fmt.Sprintf("%d", t))
		case float64:
			v.Set(key, fmt.Sprintf("%f", t))
		case bool:
			v.Set(key, fmt.Sprintf("%t", t))
		default:
			if jsonData, err := json.Marshal(val); err == nil {
				v.Set(key, string(jsonData))
			}
		}
	}
	return v.Encode()
}

// APIError represents a Moonraker API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("moonraker API error: HTTP %d - %s", e.StatusCode, e.Message)
}
