package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"
)

// HTTP3Client is an HTTP/3 QUIC client for making requests
type HTTP3Client struct {
	client    *http.Client
	logger    *zap.Logger
	tlsConfig *tls.Config
}

// HTTP3ClientConfig holds HTTP/3 client configuration
type HTTP3ClientConfig struct {
	SkipVerify      bool          // Skip TLS certificate verification (for testing)
	Timeout         time.Duration // Request timeout
	MaxIdleTimeout  time.Duration // QUIC max idle timeout
	EnableDatagrams bool          // Enable QUIC datagrams
}

// DefaultHTTP3ClientConfig returns default configuration
func DefaultHTTP3ClientConfig() *HTTP3ClientConfig {
	return &HTTP3ClientConfig{
		SkipVerify:      false,
		Timeout:         30 * time.Second,
		MaxIdleTimeout:  30 * time.Second,
		EnableDatagrams: true,
	}
}

// NewHTTP3Client creates a new HTTP/3 client with QUIC transport
func NewHTTP3Client(logger *zap.Logger, config *HTTP3ClientConfig) *HTTP3Client {
	if config == nil {
		config = DefaultHTTP3ClientConfig()
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: config.SkipVerify,
		NextProtos:         []string{"h3"}, // HTTP/3
	}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig: &quic.Config{
			MaxIdleTimeout:  config.MaxIdleTimeout,
			EnableDatagrams: config.EnableDatagrams,
			KeepAlivePeriod: 10 * time.Second,
		},
	}

	client := &http.Client{
		Transport: roundTripper,
		Timeout:   config.Timeout,
	}

	return &HTTP3Client{
		client:    client,
		logger:    logger,
		tlsConfig: tlsConfig,
	}
}

// Get performs an HTTP/3 GET request
func (c *HTTP3Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.logger.Debug("HTTP/3 GET request",
		zap.String("url", url),
		zap.String("protocol", "HTTP/3"),
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP/3 GET failed: %w", err)
	}

	c.logger.Debug("HTTP/3 GET response",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode),
		zap.String("protocol", resp.Proto),
	)

	return resp, nil
}

// Post performs an HTTP/3 POST request with JSON body
func (c *HTTP3Client) Post(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debug("HTTP/3 POST request",
		zap.String("url", url),
		zap.String("protocol", "HTTP/3"),
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP/3 POST failed: %w", err)
	}

	c.logger.Debug("HTTP/3 POST response",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode),
		zap.String("protocol", resp.Proto),
	)

	return resp, nil
}

// PostRaw performs an HTTP/3 POST request with raw body
func (c *HTTP3Client) PostRaw(ctx context.Context, url string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	c.logger.Debug("HTTP/3 POST raw request",
		zap.String("url", url),
		zap.String("protocol", "HTTP/3"),
	)

	return c.client.Do(req)
}

// Do performs a custom HTTP/3 request
func (c *HTTP3Client) Do(req *http.Request) (*http.Response, error) {
	c.logger.Debug("HTTP/3 custom request",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.String("protocol", "HTTP/3"),
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP/3 request failed: %w", err)
	}

	c.logger.Debug("HTTP/3 custom response",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Int("status", resp.StatusCode),
		zap.String("protocol", resp.Proto),
	)

	return resp, nil
}

// Close closes the HTTP/3 client and releases resources
func (c *HTTP3Client) Close() error {
	if transport, ok := c.client.Transport.(*http3.RoundTripper); ok {
		transport.Close()
	}
	c.logger.Debug("HTTP/3 client closed")
	return nil
}

// GetProtocol returns the negotiated protocol from a response
func GetProtocol(resp *http.Response) string {
	if resp == nil {
		return "unknown"
	}
	return resp.Proto
}

// IsHTTP3 checks if the response used HTTP/3 protocol
func IsHTTP3(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	// HTTP/3 protocol identifier can be "h3" or "h3-29" (draft version)
	proto := resp.Proto
	return proto == "HTTP/3.0" || proto == "h3" || proto == "h3-29"
}
