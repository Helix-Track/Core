package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"
)

// HTTP3Server wraps the http3.Server for QUIC-based HTTP/3 communication
type HTTP3Server struct {
	server    *http3.Server
	router    *gin.Engine
	logger    *zap.Logger
	tlsConfig *tls.Config
	addr      string
}

// NewHTTP3Server creates a new HTTP/3 server with TLS 1.3 and QUIC
func NewHTTP3Server(router *gin.Engine, certFile, keyFile string, logger *zap.Logger) (*HTTP3Server, error) {
	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS config with TLS 1.3 and HTTP/3 ALPN
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		NextProtos:   []string{"h3"}, // HTTP/3 protocol identifier
	}

	// Create HTTP/3 server with QUIC configuration
	server := &http3.Server{
		Handler:   router,
		TLSConfig: tlsConfig,
		QUICConfig: &quic.Config{
			MaxIdleTimeout:             30 * time.Second,
			MaxIncomingStreams:         1000,
			MaxIncomingUniStreams:      1000,
			MaxStreamReceiveWindow:     6 * 1024 * 1024,   // 6 MB per stream
			MaxConnectionReceiveWindow: 15 * 1024 * 1024,  // 15 MB total
			EnableDatagrams:            true,
			KeepAlivePeriod:            10 * time.Second,
		},
	}

	return &HTTP3Server{
		server:    server,
		router:    router,
		logger:    logger,
		tlsConfig: tlsConfig,
	}, nil
}

// Start starts the HTTP/3 server on the specified address
func (s *HTTP3Server) Start(addr string) error {
	s.addr = addr
	s.server.Addr = addr

	s.logger.Info("Starting HTTP/3 QUIC server",
		zap.String("addr", addr),
		zap.String("protocol", "HTTP/3"),
		zap.String("tls_version", "TLS 1.3"),
	)

	// Start HTTP/3 server
	return s.server.ListenAndServeTLS("", "") // Certs already in TLSConfig
}

// Shutdown gracefully shuts down the HTTP/3 server
func (s *HTTP3Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP/3 server...")

	// http3.Server doesn't support context-based shutdown like http.Server
	// We use Close() which immediately closes all connections
	return s.server.Close()
}

// GetAddress returns the server address
func (s *HTTP3Server) GetAddress() string {
	return s.addr
}

// SetQUICConfig allows customizing QUIC configuration
func (s *HTTP3Server) SetQUICConfig(config *quic.Config) {
	s.server.QUICConfig = config
}
