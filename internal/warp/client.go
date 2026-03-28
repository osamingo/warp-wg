package warp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL  = "https://api.cloudflareclient.com"
	apiVersion      = "v0a2158"
	userAgent       = "okhttp/0.7.21"
	cfClientVersion = "a-7.21-0721"

	maxResponseSize = 1 << 20 // 1 MB
)

// Client is an HTTP client for the Cloudflare WARP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new WARP API client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// RegisterRequest is the request body for POST /reg.
type RegisterRequest struct {
	Key       string `json:"key"`
	InstallID string `json:"install_id"`
	FcmToken  string `json:"fcm_token"`
	TOS       string `json:"tos"`
	Model     string `json:"model"`
	Type      string `json:"type"`
	Locale    string `json:"locale"`
}

// RegisterResponse is the response from POST /reg.
type RegisterResponse struct {
	ID      string       `json:"id"`
	Token   string       `json:"token"`
	Account Account      `json:"account"`
	Config  DeviceConfig `json:"config"`
}

// DeviceResponse is the response from GET /reg/{deviceId}.
type DeviceResponse struct {
	ID      string       `json:"id"`
	Account Account      `json:"account"`
	Config  DeviceConfig `json:"config"`
}

// Account holds the WARP account information.
type Account struct {
	ID          string `json:"id"`
	AccountType string `json:"account_type"`
	License     string `json:"license"`
	PremiumData uint64 `json:"premium_data"`
	Quota       uint64 `json:"quota"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// DeviceConfig holds the WireGuard configuration from the API.
type DeviceConfig struct {
	ClientID  string `json:"client_id"`
	Interface struct {
		Addresses struct {
			V4 string `json:"v4"`
			V6 string `json:"v6"`
		} `json:"addresses"`
	} `json:"interface"`
	Peers []Peer `json:"peers"`
}

// Peer represents a WireGuard peer from the API response.
type Peer struct {
	PublicKey string `json:"public_key"`
	Endpoint  struct {
		Host string `json:"host"`
		V4   string `json:"v4"`
		V6   string `json:"v6"`
	} `json:"endpoint"`
}

// BoundDevice represents a device linked to the account.
type BoundDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	Type      string `json:"type"`
	Active    bool   `json:"active"`
	Created   string `json:"created"`
	Activated string `json:"activated"`
}

// UpdateDeviceRequest is the request body for PATCH /reg/{deviceId}.
type UpdateDeviceRequest struct {
	Key string `json:"key"`
}

// UpdateAccountRequest is the request body for PUT /reg/{deviceId}/account.
type UpdateAccountRequest struct {
	License string `json:"license"`
}

// Register creates a new WARP device registration.
func (c *Client) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	if err := c.request(ctx, http.MethodPost, c.regBaseURL(), nil, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Device retrieves the device information for the given device ID.
func (c *Client) Device(ctx context.Context, deviceID, token string) (*DeviceResponse, error) {
	var resp DeviceResponse
	if err := c.request(ctx, http.MethodGet, c.regDeviceURL(deviceID), authHeader(token), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteDevice deletes the device registration.
func (c *Client) DeleteDevice(ctx context.Context, deviceID, token string) error {
	return c.request(ctx, http.MethodDelete, c.regDeviceURL(deviceID), authHeader(token), nil, nil)
}

// UpdateDeviceKey updates the WireGuard public key for the device.
func (c *Client) UpdateDeviceKey(ctx context.Context, deviceID, token string, req *UpdateDeviceRequest) (*DeviceResponse, error) {
	var resp DeviceResponse
	if err := c.request(ctx, http.MethodPatch, c.regDeviceURL(deviceID), authHeader(token), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateAccount updates the account license key.
func (c *Client) UpdateAccount(ctx context.Context, deviceID, token string, req *UpdateAccountRequest) (*Account, error) {
	url := c.regDeviceURL(deviceID) + "/account"
	var resp Account
	if err := c.request(ctx, http.MethodPut, url, authHeader(token), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BoundDevices retrieves the list of devices linked to the account.
func (c *Client) BoundDevices(ctx context.Context, deviceID, token string) ([]BoundDevice, error) {
	url := c.regDeviceURL(deviceID) + "/account/devices"
	var resp []BoundDevice
	if err := c.request(ctx, http.MethodGet, url, authHeader(token), nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// APIError represents an error response from the Cloudflare WARP API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("warp api: %d: %s", e.StatusCode, e.Body)
}

func (c *Client) regBaseURL() string {
	return fmt.Sprintf("%s/%s/reg", c.baseURL, apiVersion)
}

func (c *Client) regDeviceURL(deviceID string) string {
	// Validate deviceID to prevent path traversal or URL manipulation.
	if strings.ContainsAny(deviceID, "/?#") {
		return c.regBaseURL() + "/invalid-device-id"
	}
	return fmt.Sprintf("%s/%s/reg/%s", c.baseURL, apiVersion, deviceID)
}

func authHeader(token string) http.Header {
	h := make(http.Header)
	h.Set("Authorization", "Bearer "+token)
	return h
}

func (c *Client) request(ctx context.Context, method, url string, headers http.Header, reqBody, respBody any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	setCommonHeaders(httpReq)
	if reqBody != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range headers {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := c.httpClient.Do(httpReq) //nolint:gosec // URL is constructed from trusted baseURL + API path
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", slog.String("error", err.Error()))
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	if respBody != nil && len(body) > 0 {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

func setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("CF-Client-Version", cfClientVersion)
}
