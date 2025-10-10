package server

import (
	"bytes"
	"context"
	"encoding/json"
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
