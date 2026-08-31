// Package aex provides a Go client for the AEX Paraguay courier and
// shipping API (v1.5.4). It covers authorization, city and delivery point
// lookups, shipping quotes, service requests and confirmations, guide
// printing and updates, tracking, inventory, and webhook handling.
//
// All API operations are performed with context.Context for cancellation
// and timeout control. The zero value of Client is not usable; construct
// one with New.
package aex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Base URLs for each environment.
const (
	// SandboxBaseURL is the test environment base URL.
	SandboxBaseURL = "https://sandbox.aex.com.py/api/v1/"
	// ProductionBaseURL is the production environment base URL.
	ProductionBaseURL = "https://aex.com.py/api/v1/"
)

// DefaultTimeout is the HTTP timeout used when no custom http.Client is
// supplied via WithHTTPClient.
const DefaultTimeout = 30 * time.Second

// Config holds the credentials used to authenticate against the AEX API.
type Config struct {
	// PublicKey (clave_publica) identifies the entity or user.
	PublicKey string
	// PrivateKey (clave_privada) authorizes access. It is never sent in
	// plain text; it is transmitted as md5(PrivateKey + SessionCode).
	PrivateKey string
	// SessionCode (codigo_sesion) is the validation piece combined with
	// PrivateKey for hashing.
	SessionCode string
	// Sandbox selects the test environment when true, production when false.
	Sandbox bool
}

// Client is an AEX API client. It is safe for concurrent use.
type Client struct {
	config     Config
	baseURL    string
	httpClient *http.Client
}

// Option customizes a Client constructed by New.
type Option func(*Client)

// WithHTTPClient sets the http.Client used for API requests. Use it to
// control timeouts, proxies, or instrumentation.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL overrides the environment base URL. Intended for testing.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// New returns a Client for the given credentials. By default the client
// targets the production environment with Sandbox=false and an http.Client
// using DefaultTimeout; pass options to customize.
func New(config Config, opts ...Option) (*Client, error) {
	if config.PublicKey == "" {
		return nil, fmt.Errorf("aex: config.PublicKey is required")
	}
	if config.PrivateKey == "" {
		return nil, fmt.Errorf("aex: config.PrivateKey is required")
	}
	if config.SessionCode == "" {
		return nil, fmt.Errorf("aex: config.SessionCode is required")
	}

	baseURL := ProductionBaseURL
	if config.Sandbox {
		baseURL = SandboxBaseURL
	}

	client := &Client{
		config:     config,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

// doRequest performs a POST request with a JSON payload against the given
// API path, decodes the JSON response body into out (when out is non-nil),
// and returns transport-level errors. API-level errors (a non-zero result
// code in the response envelope) are surfaced by the caller through
// checkResponse, since response types embed their own envelope fields.
func (c *Client) doRequest(ctx context.Context, path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("aex: encode request %s: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("aex: build request %s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("aex: request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			Endpoint:   path,
			StatusCode: resp.StatusCode,
		}
	}

	if out == nil {
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			return fmt.Errorf("aex: drain response %s: %w", path, err)
		}
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("aex: decode response %s: %w", path, err)
	}
	return nil
}
