package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func TestJWTMiddleware_Validate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing Authorization header", func(t *testing.T) {
		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return false },
		}
		middleware := NewJWTMiddleware(mockAuth, "secret")

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Authorization header format", func(t *testing.T) {
		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return false },
		}
		middleware := NewJWTMiddleware(mockAuth, "secret")

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid token with auth service", func(t *testing.T) {
		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return true },
			ValidateTokenFunc: func(ctx context.Context, token string) (*models.JWTClaims, error) {
				return &models.JWTClaims{
					Username: "testuser",
					Role:     "admin",
				}, nil
			},
		}
		middleware := NewJWTMiddleware(mockAuth, "")

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			claims, exists := GetClaims(c)
			assert.True(t, exists)
			assert.Equal(t, "testuser", claims.Username)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid token with auth service", func(t *testing.T) {
		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return true },
			ValidateTokenFunc: func(ctx context.Context, token string) (*models.JWTClaims, error) {
				return nil, jwt.ErrTokenExpired
			},
		}
		middleware := NewJWTMiddleware(mockAuth, "")

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid local token", func(t *testing.T) {
		secretKey := "test-secret-key"

		// Create a valid token
		claims := &models.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "authentication",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Username: "testuser",
			Role:     "user",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return false },
		}
		middleware := NewJWTMiddleware(mockAuth, secretKey)

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			claims, exists := GetClaims(c)
			assert.True(t, exists)
			assert.Equal(t, "testuser", claims.Username)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid local token - wrong secret", func(t *testing.T) {
		secretKey := "test-secret-key"

		// Create a token with different secret
		claims := &models.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "authentication",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Username: "testuser",
			Role:     "user",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		mockAuth := &services.MockAuthService{
			IsEnabledFunc: func() bool { return false },
		}
		middleware := NewJWTMiddleware(mockAuth, secretKey)

		router := gin.New()
		router.Use(middleware.Validate())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Claims exist in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		expectedClaims := &models.JWTClaims{
			Username: "testuser",
			Role:     "admin",
		}
		c.Set("claims", expectedClaims)

		claims, exists := GetClaims(c)
		assert.True(t, exists)
		assert.Equal(t, expectedClaims, claims)
	})

	t.Run("Claims don't exist in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		claims, exists := GetClaims(c)
		assert.False(t, exists)
		assert.Nil(t, claims)
	})

	t.Run("Claims with wrong type in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("claims", "wrong-type")

		claims, exists := GetClaims(c)
		assert.False(t, exists)
		assert.Nil(t, claims)
	})
}

func TestGetUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Username exists in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("username", "testuser")

		username, exists := GetUsername(c)
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)
	})

	t.Run("Username doesn't exist in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		username, exists := GetUsername(c)
		assert.False(t, exists)
		assert.Empty(t, username)
	})

	t.Run("Username with wrong type in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("username", 12345)

		username, exists := GetUsername(c)
		assert.False(t, exists)
		assert.Empty(t, username)
	})
}

func TestJWTMiddleware_validateTokenLocally_ExpiredToken(t *testing.T) {
	secretKey := "test-secret-key"

	// Create an expired token
	claims := &models.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		Username: "testuser",
		Role:     "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return false },
	}
	middleware := NewJWTMiddleware(mockAuth, secretKey)

	router := gin.New()
	router.Use(middleware.Validate())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewJWTMiddleware(t *testing.T) {
	mockAuth := &services.MockAuthService{}
	secretKey := "test-secret"

	middleware := NewJWTMiddleware(mockAuth, secretKey)

	assert.NotNil(t, middleware)
	assert.Equal(t, mockAuth, middleware.authService)
	assert.Equal(t, secretKey, middleware.secretKey)
}
