package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestPermissionMiddleware_Disabled tests middleware when permission service is disabled
func TestPermissionMiddleware_Disabled(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return false
		},
	}

	router := gin.New()
	router.Use(PermissionMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

// TestPermissionMiddleware_NoClaims tests middleware when no JWT claims exist
func TestPermissionMiddleware_NoClaims(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	router := gin.New()
	router.Use(PermissionMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "No authentication provided")
}

// TestPermissionMiddleware_InvalidClaims tests middleware with invalid claims type
func TestPermissionMiddleware_InvalidClaims(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set invalid claims (wrong type)
		c.Set("claims", "invalid_claims_string")
		c.Next()
	})
	router.Use(PermissionMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Invalid authentication claims")
}

// TestPermissionMiddleware_ValidClaims tests middleware with valid JWT claims
func TestPermissionMiddleware_ValidClaims(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set valid JWT claims
		claims := &models.JWTClaims{
			Username: "testuser",
			Name:     "Test User",
			Role:     "admin",
		}
		c.Set("claims", claims)
		c.Next()
	})
	router.Use(PermissionMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)

		permService, exists := c.Get("permissionService")
		assert.True(t, exists)
		assert.NotNil(t, permService)

		c.JSON(http.StatusOK, gin.H{"status": "ok", "username": username})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
}

// TestRequirePermission_Disabled tests RequirePermission when service is disabled
func TestRequirePermission_Disabled(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return false
		},
	}

	router := gin.New()
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequirePermission_NoUsername tests RequirePermission when no username in context
func TestRequirePermission_NoUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	router := gin.New()
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "No username in context")
}

// TestRequirePermission_InvalidUsername tests RequirePermission with invalid username type
func TestRequirePermission_InvalidUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", 12345) // Invalid type
		c.Next()
	})
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInternalError, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Invalid username type")
}

// TestRequirePermission_PermissionGranted tests RequirePermission with permission granted
func TestRequirePermission_PermissionGranted(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return true, nil
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequirePermission_PermissionDenied tests RequirePermission with permission denied
func TestRequirePermission_PermissionDenied(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Permission denied")
}

// TestRequirePermission_ServiceError tests RequirePermission with service error
func TestRequirePermission_ServiceError(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return false, assert.AnError
		},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})
	router.Use(RequirePermission(mockService, "node1", models.PermissionRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInternalError, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Permission check failed")
}

// TestRequirePermission_DifferentLevels tests RequirePermission with different permission levels
func TestRequirePermission_DifferentLevels(t *testing.T) {
	tests := []struct {
		name          string
		userLevel     models.PermissionLevel
		requiredLevel models.PermissionLevel
		expectedCode  int
	}{
		{
			name:          "User has READ, requires READ",
			userLevel:     models.PermissionRead,
			requiredLevel: models.PermissionRead,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "User has UPDATE, requires READ",
			userLevel:     models.PermissionUpdate,
			requiredLevel: models.PermissionRead,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "User has READ, requires UPDATE",
			userLevel:     models.PermissionRead,
			requiredLevel: models.PermissionUpdate,
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "User has DELETE, requires UPDATE",
			userLevel:     models.PermissionDelete,
			requiredLevel: models.PermissionUpdate,
			expectedCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &services.MockPermissionService{
				IsEnabledFunc: func() bool {
					return true
				},
				CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
					return tt.userLevel.HasPermission(requiredLevel), nil
				},
			}

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("username", "testuser")
				c.Next()
			})
			router.Use(RequirePermission(mockService, "node1", tt.requiredLevel))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestCheckPermissionForAction_Disabled tests CheckPermissionForAction when disabled
func TestCheckPermissionForAction_Disabled(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return false
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	allowed := CheckPermissionForAction(c, mockService, "read", "node1")
	assert.True(t, allowed, "Should allow when disabled")
}

// TestCheckPermissionForAction_NoUsername tests CheckPermissionForAction without username
func TestCheckPermissionForAction_NoUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	allowed := CheckPermissionForAction(c, mockService, "read", "node1")
	assert.False(t, allowed, "Should deny without username")
}

// TestCheckPermissionForAction_InvalidUsername tests CheckPermissionForAction with invalid username
func TestCheckPermissionForAction_InvalidUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", 12345) // Invalid type
	allowed := CheckPermissionForAction(c, mockService, "read", "node1")
	assert.False(t, allowed, "Should deny with invalid username type")
}

// TestCheckPermissionForAction_Success tests CheckPermissionForAction with valid permissions
func TestCheckPermissionForAction_Success(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		allowed   bool
		expected  bool
	}{
		{name: "Read action allowed", action: "read", allowed: true, expected: true},
		{name: "Create action allowed", action: "create", allowed: true, expected: true},
		{name: "Update action allowed", action: "update", allowed: true, expected: true},
		{name: "Delete action allowed", action: "delete", allowed: true, expected: true},
		{name: "Read action denied", action: "read", allowed: false, expected: false},
		{name: "Create action denied", action: "create", allowed: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &services.MockPermissionService{
				IsEnabledFunc: func() bool {
					return true
				},
				CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
					return tt.allowed, nil
				},
			}

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("username", "testuser")
			allowed := CheckPermissionForAction(c, mockService, tt.action, "node1")
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestCheckPermissionForAction_ServiceError tests CheckPermissionForAction with service error
func TestCheckPermissionForAction_ServiceError(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return false, assert.AnError
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "testuser")
	allowed := CheckPermissionForAction(c, mockService, "read", "node1")
	assert.False(t, allowed, "Should deny on service error")
}

// TestGetUserPermissions_NoUsername tests GetUserPermissions without username
func TestGetUserPermissions_NoUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	permissions, err := GetUserPermissions(c, mockService)
	assert.NoError(t, err)
	assert.Nil(t, permissions)
}

// TestGetUserPermissions_InvalidUsername tests GetUserPermissions with invalid username
func TestGetUserPermissions_InvalidUsername(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", 12345) // Invalid type
	permissions, err := GetUserPermissions(c, mockService)
	assert.NoError(t, err)
	assert.Nil(t, permissions)
}

// TestGetUserPermissions_Success tests GetUserPermissions with valid username
func TestGetUserPermissions_Success(t *testing.T) {
	expectedPermissions := []models.Permission{
		{
			ID:      "perm1",
			Context: "node1",
			Level:   models.PermissionRead,
		},
		{
			ID:      "perm2",
			Context: "node1→account1",
			Level:   models.PermissionUpdate,
		},
	}

	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		GetUserPermissionsFunc: func(ctx context.Context, username string) ([]models.Permission, error) {
			return expectedPermissions, nil
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "testuser")
	permissions, err := GetUserPermissions(c, mockService)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Equal(t, "perm1", permissions[0].ID)
	assert.Equal(t, "perm2", permissions[1].ID)
}

// TestGetUserPermissions_ServiceError tests GetUserPermissions with service error
func TestGetUserPermissions_ServiceError(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		GetUserPermissionsFunc: func(ctx context.Context, username string) ([]models.Permission, error) {
			return nil, assert.AnError
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "testuser")
	permissions, err := GetUserPermissions(c, mockService)
	assert.Error(t, err)
	assert.Nil(t, permissions)
}

// TestPermissionMiddleware_Integration tests full middleware integration
func TestPermissionMiddleware_Integration(t *testing.T) {
	mockService := &services.MockPermissionService{
		IsEnabledFunc: func() bool {
			return true
		},
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			// Grant permission to "admin" but deny to "user"
			return username == "admin", nil
		},
	}

	tests := []struct {
		name         string
		username     string
		context      string
		level        models.PermissionLevel
		expectedCode int
	}{
		{
			name:         "Admin has permission",
			username:     "admin",
			context:      "node1",
			level:        models.PermissionRead,
			expectedCode: http.StatusOK,
		},
		{
			name:         "User denied permission",
			username:     "user",
			context:      "node1",
			level:        models.PermissionRead,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				claims := &models.JWTClaims{
					Username: tt.username,
				}
				c.Set("claims", claims)
				c.Next()
			})
			router.Use(PermissionMiddleware(mockService))
			router.Use(RequirePermission(mockService, tt.context, tt.level))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
