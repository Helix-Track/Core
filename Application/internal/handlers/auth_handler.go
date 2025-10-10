package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	db         database.Database
	jwtService *services.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db database.Database) *AuthHandler {
	jwtService := services.NewJWTService("", "", 24) // Default JWT service
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid registration data",
			"",
		))
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to process registration",
			"",
		))
		return
	}

	// Create user
	user := models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Email:        req.Email,
		Name:         req.Name,
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Deleted:      false,
	}

	// Insert into database
	query := `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.Email,
		user.Name,
		user.Role,
		user.CreatedAt.Unix(),
		user.UpdatedAt.Unix(),
		0,
	)

	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"Username or email already exists",
			"",
		))
		return
	}

	// Return user response
	response := models.NewSuccessResponse(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
	})

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid login data",
			"",
		))
		return
	}

	// Get user from database
	query := `
		SELECT id, username, password_hash, email, name, role, created_at, updated_at
		FROM users
		WHERE username = ? AND deleted = 0
	`

	var user models.User
	var createdAt, updatedAt int64

	err := h.db.QueryRow(context.Background(), query, req.Username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Name,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		logger.Error("User not found", zap.Error(err), zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid username or password",
			"",
		))
		return
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Error("Invalid password", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid username or password",
			"",
		))
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.Username, user.Email, user.Name, user.Role)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to generate authentication token",
			"",
		))
		return
	}

	// Return success with token
	response := models.NewSuccessResponse(map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
	})

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// For a stateless JWT system, logout just returns success
	// In a real system with token blacklisting, you'd invalidate the token
	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Successfully logged out",
	})
	c.JSON(http.StatusOK, response)
}

// InitializeUserTable creates the users table if it doesn't exist
func InitializeUserTable(db database.Database) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		);
	`

	_, err := db.Exec(context.Background(), createTableSQL)
	if err != nil {
		return err
	}

	// Create indexes
	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err = db.Exec(context.Background(), indexSQL)
	if err != nil {
		return err
	}

	// Create default test users if they don't exist
	testUsers := []struct {
		username string
		password string
		email    string
		name     string
		role     string
	}{
		{"admin_user", "Admin@123456", "admin@test.com", "Admin User", "user"},
		{"viewer", "Viewer@123456", "viewer@helixtrack.test", "Viewer User", "user"},
		{"project_manager", "PM@123456", "pm@helixtrack.test", "Project Manager", "user"},
		{"developer", "Dev@123456", "dev@helixtrack.test", "Developer", "user"},
	}

	for _, testUser := range testUsers {
		var count int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE username = ?", testUser.username).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			// Hash the password
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(testUser.password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			now := time.Now().Unix()
			_, err = db.Exec(context.Background(), `
				INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, uuid.New().String(), testUser.username, string(passwordHash), testUser.email, testUser.name, testUser.role, now, now, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
