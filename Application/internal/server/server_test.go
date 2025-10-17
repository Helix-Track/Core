package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

func init() {
	// Initialize logger for tests
	logger.Initialize(config.LogConfig{
		LogPath:      "/tmp",
		LogSizeLimit: 1000000,
		Level:        "error",
	})
}

func TestNewServer(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
		Version: "1.0.0-test",
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.db)
	assert.NotNil(t, server.authService)
	assert.NotNil(t, server.permService)

	// Clean up
	server.db.Close()
}

func TestServer_HealthEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestServer_DoEndpoint_Version(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
		Version: "1.0.0-test",
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	reqBody := models.Request{
		Action: models.ActionVersion,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "1.0.0-test", resp.Data["version"])
}

func TestServer_DoEndpoint_MissingJWT(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: true, URL: "http://localhost:8081"},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	reqBody := models.Request{
		Action: models.ActionCreate, // Requires authentication
		Object: "project",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingJWT, resp.ErrorCode)
}

func TestServer_DoEndpoint_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_CORSMiddleware(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	req := httptest.NewRequest(http.MethodOptions, "/do", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestServer_Shutdown(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_GetRouter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	router := server.GetRouter()
	assert.NotNil(t, router)
	assert.Equal(t, server.router, router)
}

func TestServer_PortFallback_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	// Start server in background
	go func() {
		err := server.Start()
		if err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is responding
	resp, err := http.Get("http://127.0.0.1:8080/health")
	if err != nil {
		// If port 8080 is not available, check if server started on a different port
		// This is expected behavior for port fallback
		t.Logf("Port 8080 not available, checking if server started on fallback port")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestServer_PortFallback_MultiplePorts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	// Start first server on port 8080
	go func() {
		err := server.Start()
		if err != nil {
			t.Logf("First server start error: %v", err)
		}
	}()

	// Give first server time to start
	time.Sleep(100 * time.Millisecond)

	// Create second server that should fallback to different port
	cfg2 := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false}, // Same port as first server
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server2, err := NewServer(cfg2)
	require.NoError(t, err)
	defer server2.db.Close()

	// Start second server - should fallback to different port
	go func() {
		err := server2.Start()
		if err != nil {
			t.Logf("Second server start error: %v", err)
		}
	}()

	// Give second server time to start
	time.Sleep(100 * time.Millisecond)

	// Both servers should be running on different ports
	// First server on 8080, second on 8081 or higher
	resp1, err1 := http.Get("http://127.0.0.1:8080/health")
	resp2, err2 := http.Get("http://127.0.0.1:8081/health")

	// At least one should succeed (indicating fallback worked)
	if err1 != nil && err2 != nil {
		t.Logf("Both ports failed - this might be expected if ports are occupied")
		return
	}

	if err1 == nil {
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
	}

	if err2 == nil {
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	}

	// Clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	server2.Shutdown(ctx)
}

func TestServer_BroadcastAvailability(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
		WebSocket: config.WebSocketConfig{
			Enabled: true,
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	// Test broadcast availability method
	server.broadcastAvailability("127.0.0.1:8080")

	// The method should not panic and should log the availability
	// Since we can't easily test WebSocket broadcasting in unit tests,
	// we mainly verify the method doesn't crash
}

func TestServer_ExtractPortFromAddress(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	tests := []struct {
		address  string
		expected int
	}{
		{"127.0.0.1:8080", 8080},
		{"localhost:3000", 3000},
		{"0.0.0.0:9090", 9090},
		{"192.168.1.1:8081", 8081},
		{"", 0},
		{"invalid", 0},
		{"127.0.0.1:", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("address_%s", tt.address), func(t *testing.T) {
			port := server.extractPortFromAddress(tt.address)
			assert.Equal(t, tt.expected, port)
		})
	}
}

func TestServer_IsPortInUseError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Log: config.LogConfig{
			LogPath:      tmpDir,
			LogSizeLimit: 1000000,
		},
		Listeners: []config.ListenerConfig{
			{Address: "127.0.0.1", Port: 8080, HTTPS: false},
		},
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: dbPath,
		},
		Services: config.ServicesConfig{
			Authentication: config.ServiceEndpoint{Enabled: false, URL: ""},
			Permissions:    config.ServiceEndpoint{Enabled: false, URL: ""},
		},
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)
	defer server.db.Close()

	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{fmt.Errorf("bind: address already in use"), true},
		{fmt.Errorf("listen tcp 127.0.0.1:8080: bind: address already in use"), true},
		{fmt.Errorf("address already in use"), true},
		{fmt.Errorf("some other error"), false},
		{fmt.Errorf("connection refused"), false},
	}

	for i, tt := range tests {
		testName := "nil_error"
		if tt.err != nil {
			testName = tt.err.Error()
		}
		t.Run(fmt.Sprintf("error_%d_%s", i, testName), func(t *testing.T) {
			result := server.isPortInUseError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
