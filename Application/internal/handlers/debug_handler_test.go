package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func TestDebugDocumentSpaceCreate(t *testing.T) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	ctx := context.Background()
	query := `CREATE TABLE document_space (
		id TEXT PRIMARY KEY,
		key TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		owner_id TEXT NOT NULL,
		is_public INTEGER DEFAULT 0,
		created INTEGER NOT NULL,
		modified INTEGER NOT NULL,
		deleted INTEGER DEFAULT 0
	)`
	_, err = db.Exec(ctx, query)
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return true, nil
		},
	}

	handler := NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.POST("/do", func(c *gin.Context) {
		var reqBody models.Request
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			c.Set("request", &reqBody)
		}
		handler.DoAction(c)
	})

	reqBody := models.Request{
		Action: models.ActionDocumentSpaceCreate,
		JWT:    "test-jwt-token",
		Data: map[string]interface{}{
			"key":       "TEST",
			"name":      "Test Space",
			"owner_id":  "user-1",
			"is_public": true,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	fmt.Printf("Error Code: %d\n", resp.ErrorCode)
	fmt.Printf("Error Message: %s\n", resp.ErrorMessage)
	fmt.Printf("Data: %+v\n", resp.Data)
}
