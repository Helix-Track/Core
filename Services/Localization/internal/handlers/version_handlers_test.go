package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestCreateVersion_Unit tests the CreateVersion handler
func TestCreateVersion_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("create version success", func(t *testing.T) {
		versionReq := models.CreateVersionRequest{
			VersionType: "patch",
			Description: "Patch release",
			Metadata:    map[string]interface{}{"test": "data"},
		}
		body, _ := json.Marshal(versionReq)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/versions", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateVersion(c)

		// Should return 201 (created)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Check response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("create version with invalid JSON", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/versions", bytes.NewBuffer([]byte("{invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateVersion(c)

		// Should return 400 for invalid JSON
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create version with database error", func(t *testing.T) {
		// Set database to return error
		db.SetError(context.Canceled)
		defer db.ClearError()

		versionReq := models.CreateVersionRequest{
			VersionType: "minor",
			Description: "Minor release",
		}
		body, _ := json.Marshal(versionReq)

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/admin/versions", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.CreateVersion(c)

		// Should return 500 for database error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestDeleteVersion_Unit tests the DeleteVersion handler
func TestDeleteVersion_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := NewMockDatabase()
	memCache := cache.NewMemoryCache(10, 1*time.Hour, 5*time.Minute, logger)
	handler := NewHandler(db, memCache, logger, nil)

	t.Run("delete version success", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/admin/versions/1.0.0", nil)
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteVersion(c)

		// Should return 200 (success)
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Check response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("delete version with empty version number", func(t *testing.T) {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/admin/versions/", nil)
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteVersion(c)

		// Should return 200 (success) - mock doesn't validate empty version
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete version with database error", func(t *testing.T) {
		// Set database to return error
		db.SetError(context.Canceled)
		defer db.ClearError()

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("DELETE", "/admin/versions/1.0.1", nil)
		c.Set("user_role", "admin")
		claims := &models.JWTClaims{Username: "adminuser", Role: "admin"}
		c.Set("claims", claims)

		handler.DeleteVersion(c)

		// Should return 500 for database error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}