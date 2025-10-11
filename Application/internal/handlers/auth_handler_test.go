package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
)

// setupAuthTestHandler creates a test auth handler with in-memory database
func setupAuthTestHandler(t *testing.T) *AuthHandler {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err, "Failed to create in-memory database")

	// Initialize user table
	err = InitializeUserTable(db)
	require.NoError(t, err, "Failed to initialize user table")

	return NewAuthHandler(db)
}

// TestAuthHandler_Register_Success tests successful user registration
func TestAuthHandler_Register_Success(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.UserRegistrationRequest{
		Username: "testuser",
		Password: "Test@123456",
		Email:    "testuser@test.com",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.NotNil(t, response.Data)

	data := response.Data
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, "testuser@test.com", data["email"])
	assert.Equal(t, "Test User", data["name"])
	assert.Equal(t, "user", data["role"])
	assert.NotEmpty(t, data["id"])

	// Verify user was inserted into database
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE username = ?", "testuser").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify password was hashed
	var passwordHash string
	err = handler.db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE username = ?", "testuser").Scan(&passwordHash)
	require.NoError(t, err)
	assert.NotEqual(t, "Test@123456", passwordHash) // Should be hashed
	assert.True(t, len(passwordHash) > 50)           // bcrypt hashes are long
}

// TestAuthHandler_Register_MissingUsername tests registration with missing username
func TestAuthHandler_Register_MissingUsername(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"password": "Test@123456",
		"email":    "testuser@test.com",
		"name":     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_MissingPassword tests registration with missing password
func TestAuthHandler_Register_MissingPassword(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"username": "testuser",
		"email":    "testuser@test.com",
		"name":     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_MissingEmail tests registration with missing email
func TestAuthHandler_Register_MissingEmail(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"username": "testuser",
		"password": "Test@123456",
		"name":     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_MissingName tests registration with missing name
func TestAuthHandler_Register_MissingName(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"username": "testuser",
		"password": "Test@123456",
		"email":    "testuser@test.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_InvalidEmail tests registration with invalid email format
func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.UserRegistrationRequest{
		Username: "testuser",
		Password: "Test@123456",
		Email:    "not-an-email",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_PasswordTooShort tests registration with password too short
func TestAuthHandler_Register_PasswordTooShort(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.UserRegistrationRequest{
		Username: "testuser",
		Password: "Short1",
		Email:    "testuser@test.com",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Register_DuplicateUsername tests registration with duplicate username
func TestAuthHandler_Register_DuplicateUsername(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert existing user
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Existing@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "existing-user-id", "existinguser", string(passwordHash), "existing@test.com", "Existing User", "user", now, now, 0)
	require.NoError(t, err)

	// Try to register with same username
	reqBody := models.UserRegistrationRequest{
		Username: "existinguser",
		Password: "Test@123456",
		Email:    "newuser@test.com",
		Name:     "New User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, response.ErrorCode)
}

// TestAuthHandler_Register_DuplicateEmail tests registration with duplicate email
func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert existing user
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Existing@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "existing-user-id", "existinguser", string(passwordHash), "existing@test.com", "Existing User", "user", now, now, 0)
	require.NoError(t, err)

	// Try to register with same email
	reqBody := models.UserRegistrationRequest{
		Username: "newuser",
		Password: "Test@123456",
		Email:    "existing@test.com",
		Name:     "New User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, response.ErrorCode)
}

// TestAuthHandler_Login_Success tests successful user login
func TestAuthHandler_Login_Success(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Register a user first
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Test@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-user-id", "testuser", string(passwordHash), "testuser@test.com", "Test User", "user", now, now, 0)
	require.NoError(t, err)

	// Login
	reqBody := models.UserLoginRequest{
		Username: "testuser",
		Password: "Test@123456",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.NotNil(t, response.Data)

	data := response.Data
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, "testuser@test.com", data["email"])
	assert.Equal(t, "Test User", data["name"])
	assert.Equal(t, "user", data["role"])
	assert.NotEmpty(t, data["token"]) // JWT token should be present
}

// TestAuthHandler_Login_MissingUsername tests login with missing username
func TestAuthHandler_Login_MissingUsername(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"password": "Test@123456",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Login_MissingPassword tests login with missing password
func TestAuthHandler_Login_MissingPassword(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"username": "testuser",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidRequest, response.ErrorCode)
}

// TestAuthHandler_Login_UserNotFound tests login with non-existent user
func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.UserLoginRequest{
		Username: "nonexistentuser",
		Password: "Test@123456",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
}

// TestAuthHandler_Login_IncorrectPassword tests login with incorrect password
func TestAuthHandler_Login_IncorrectPassword(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Register a user first
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Test@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-user-id", "testuser", string(passwordHash), "testuser@test.com", "Test User", "user", now, now, 0)
	require.NoError(t, err)

	// Login with wrong password
	reqBody := models.UserLoginRequest{
		Username: "testuser",
		Password: "WrongPassword@123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
}

// TestAuthHandler_Login_DeletedUser tests login with deleted user
func TestAuthHandler_Login_DeletedUser(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Register a deleted user
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Test@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-user-id", "deleteduser", string(passwordHash), "deleted@test.com", "Deleted User", "user", now, now, 1)
	require.NoError(t, err)

	// Try to login
	reqBody := models.UserLoginRequest{
		Username: "deleteduser",
		Password: "Test@123456",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeUnauthorized, response.ErrorCode)
}

// TestAuthHandler_Logout_Success tests successful logout
func TestAuthHandler_Logout_Success(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.NotNil(t, response.Data)

	data := response.Data
	assert.Equal(t, "Successfully logged out", data["message"])
}

// TestAuthHandler_FullAuthenticationCycle tests full registration -> login -> logout cycle
func TestAuthHandler_FullAuthenticationCycle(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Step 1: Register
	registerReq := models.UserRegistrationRequest{
		Username: "cycleuser",
		Password: "Cycle@123456",
		Email:    "cycle@test.com",
		Name:     "Cycle User",
	}

	registerBody, _ := json.Marshal(registerReq)
	req1 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1

	handler.Register(c1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var registerResponse models.Response
	err := json.Unmarshal(w1.Body.Bytes(), &registerResponse)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, registerResponse.ErrorCode)

	// Step 2: Login
	loginReq := models.UserLoginRequest{
		Username: "cycleuser",
		Password: "Cycle@123456",
	}

	loginBody, _ := json.Marshal(loginReq)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2

	handler.Login(c2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var loginResponse models.Response
	err = json.Unmarshal(w2.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, loginResponse.ErrorCode)

	loginData := loginResponse.Data
	assert.NotEmpty(t, loginData["token"])

	// Step 3: Logout
	req3 := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w3 := httptest.NewRecorder()

	c3, _ := gin.CreateTestContext(w3)
	c3.Request = req3

	handler.Logout(c3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var logoutResponse models.Response
	err = json.Unmarshal(w3.Body.Bytes(), &logoutResponse)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, logoutResponse.ErrorCode)
}

// TestInitializeUserTable tests user table initialization
func TestInitializeUserTable(t *testing.T) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err, "Failed to create in-memory database")

	// Initialize table
	err = InitializeUserTable(db)
	require.NoError(t, err, "Failed to initialize user table")

	// Verify table exists
	var tableName string
	err = db.QueryRow(context.Background(), "SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "users", tableName)

	// Verify default users were created
	var count int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE deleted = 0").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 4, count) // 4 default test users

	// Verify default users exist
	testUsernames := []string{"admin_user", "viewer", "project_manager", "developer"}
	for _, username := range testUsernames {
		var userCount int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&userCount)
		require.NoError(t, err)
		assert.Equal(t, 1, userCount, "Default user %s should exist", username)
	}
}

// TestAuthHandler_Register_PasswordHashing tests that passwords are properly hashed
func TestAuthHandler_Register_PasswordHashing(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.UserRegistrationRequest{
		Username: "hashtest",
		Password: "HashTest@123",
		Email:    "hashtest@test.com",
		Name:     "Hash Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Retrieve stored password hash
	var storedHash string
	err := handler.db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE username = ?", "hashtest").Scan(&storedHash)
	require.NoError(t, err)

	// Verify password is hashed (bcrypt format starts with $2a$ or $2b$)
	assert.True(t, len(storedHash) > 50, "Password hash should be long")
	assert.Contains(t, storedHash, "$2", "Password should be bcrypt hashed")

	// Verify password matches the hash
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("HashTest@123"))
	assert.NoError(t, err, "Password should match the hash")

	// Verify wrong password doesn't match
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("WrongPassword"))
	assert.Error(t, err, "Wrong password should not match the hash")
}

// TestAuthHandler_Login_JWTTokenGeneration tests that JWT token is generated on login
func TestAuthHandler_Login_JWTTokenGeneration(t *testing.T) {
	handler := setupAuthTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Register a user first
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("JWT@123456"), bcrypt.DefaultCost)
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO users (id, username, password_hash, email, name, role, created_at, updated_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "jwt-user-id", "jwtuser", string(passwordHash), "jwt@test.com", "JWT User", "user", now, now, 0)
	require.NoError(t, err)

	// Login
	reqBody := models.UserLoginRequest{
		Username: "jwtuser",
		Password: "JWT@123456",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response.Data
	token := data["token"].(string)

	// Verify token is not empty
	assert.NotEmpty(t, token, "JWT token should be generated")

	// Verify token has JWT format (header.payload.signature)
	assert.Contains(t, token, ".", "JWT token should contain dots")
	assert.True(t, len(token) > 20, "JWT token should be long enough")
}
