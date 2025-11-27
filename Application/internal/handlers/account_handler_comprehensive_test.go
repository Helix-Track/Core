package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// TestAccountHandler_Comprehensive tests the complete account lifecycle
func TestAccountHandler_Comprehensive(t *testing.T) {
	handler := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Step 1: Create an account
	createReqBody := models.Request{
		Action: models.ActionAccountCreate,
		Data: map[string]interface{}{
			"title":       "Comprehensive Test Account",
			"description": "Test account for comprehensive testing",
		},
	}

	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	createC, _ := gin.CreateTestContext(createW)
	createC.Request = createReq

	handler.AccountCreate(createC, &createReqBody)

	assert.Equal(t, http.StatusOK, createW.Code)

	var createResponse models.Response
	err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, createResponse.ErrorCode)
	assert.NotNil(t, createResponse.Data)

	// Extract account ID from creation response
	dataMap := createResponse.Data
	accountData, ok := dataMap["account"].(map[string]interface{})
	require.True(t, ok)
	accountID := accountData["id"].(string)
	require.NotEmpty(t, accountID)

	// Step 2: Read the created account
	readReqBody := models.Request{
		Action: models.ActionAccountRead,
		Data: map[string]interface{}{
			"id": accountID,
		},
	}

	readBody, _ := json.Marshal(readReqBody)
	readReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(readBody))
	readReq.Header.Set("Content-Type", "application/json")
	readW := httptest.NewRecorder()

	readC, _ := gin.CreateTestContext(readW)
	readC.Request = readReq

	handler.AccountRead(readC, &readReqBody)

	assert.Equal(t, http.StatusOK, readW.Code)

	var readResponse models.Response
	err = json.Unmarshal(readW.Body.Bytes(), &readResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, readResponse.ErrorCode)
	assert.NotNil(t, readResponse.Data)

	// Verify read account data
	readDataMap := readResponse.Data
	readAccountData, ok := readDataMap["account"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, accountID, readAccountData["id"])
	assert.Equal(t, "Comprehensive Test Account", readAccountData["title"])
	assert.Equal(t, "Test account for comprehensive testing", readAccountData["description"])

	// Step 3: Modify the account
	modifyReqBody := models.Request{
		Action: models.ActionAccountModify,
		Data: map[string]interface{}{
			"id":          accountID,
			"title":       "Modified Comprehensive Account",
			"description": "Modified description for comprehensive testing",
		},
	}

	modifyBody, _ := json.Marshal(modifyReqBody)
	modifyReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(modifyBody))
	modifyReq.Header.Set("Content-Type", "application/json")
	modifyW := httptest.NewRecorder()

	modifyC, _ := gin.CreateTestContext(modifyW)
	modifyC.Request = modifyReq

	handler.AccountModify(modifyC, &modifyReqBody)

	assert.Equal(t, http.StatusOK, modifyW.Code)

	var modifyResponse models.Response
	err = json.Unmarshal(modifyW.Body.Bytes(), &modifyResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, modifyResponse.ErrorCode)
	assert.NotNil(t, modifyResponse.Data)

	// Step 4: Read the modified account to verify changes
	readAfterModifyReqBody := models.Request{
		Action: models.ActionAccountRead,
		Data: map[string]interface{}{
			"id": accountID,
		},
	}

	readAfterModifyBody, _ := json.Marshal(readAfterModifyReqBody)
	readAfterModifyReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(readAfterModifyBody))
	readAfterModifyReq.Header.Set("Content-Type", "application/json")
	readAfterModifyW := httptest.NewRecorder()

	readAfterModifyC, _ := gin.CreateTestContext(readAfterModifyW)
	readAfterModifyC.Request = readAfterModifyReq

	handler.AccountRead(readAfterModifyC, &readAfterModifyReqBody)

	assert.Equal(t, http.StatusOK, readAfterModifyW.Code)

	var readAfterModifyResponse models.Response
	err = json.Unmarshal(readAfterModifyW.Body.Bytes(), &readAfterModifyResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, readAfterModifyResponse.ErrorCode)
	assert.NotNil(t, readAfterModifyResponse.Data)

	// Verify modified account data
	readAfterModifyDataMap := readAfterModifyResponse.Data
	readAfterModifyAccountData, ok := readAfterModifyDataMap["account"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, accountID, readAfterModifyAccountData["id"])
	assert.Equal(t, "Modified Comprehensive Account", readAfterModifyAccountData["title"])
	assert.Equal(t, "Modified description for comprehensive testing", readAfterModifyAccountData["description"])

	// Step 5: Remove the account
	removeReqBody := models.Request{
		Action: models.ActionAccountRemove,
		Data: map[string]interface{}{
			"id": accountID,
		},
	}

	removeBody, _ := json.Marshal(removeReqBody)
	removeReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(removeBody))
	removeReq.Header.Set("Content-Type", "application/json")
	removeW := httptest.NewRecorder()

	removeC, _ := gin.CreateTestContext(removeW)
	removeC.Request = removeReq

	handler.AccountRemove(removeC, &removeReqBody)

	assert.Equal(t, http.StatusOK, removeW.Code)

	var removeResponse models.Response
	err = json.Unmarshal(removeW.Body.Bytes(), &removeResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, removeResponse.ErrorCode)
	assert.NotNil(t, removeResponse.Data)

	// Verify removal response
	removeDataMap := removeResponse.Data
	assert.Equal(t, accountID, removeDataMap["id"])
	assert.Equal(t, true, removeDataMap["deleted"])

	// Step 6: Try to read the removed account (should fail)
	readAfterRemoveReqBody := models.Request{
		Action: models.ActionAccountRead,
		Data: map[string]interface{}{
			"id": accountID,
		},
	}

	readAfterRemoveBody, _ := json.Marshal(readAfterRemoveReqBody)
	readAfterRemoveReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(readAfterRemoveBody))
	readAfterRemoveReq.Header.Set("Content-Type", "application/json")
	readAfterRemoveW := httptest.NewRecorder()

	readAfterRemoveC, _ := gin.CreateTestContext(readAfterRemoveW)
	readAfterRemoveC.Request = readAfterRemoveReq

	handler.AccountRead(readAfterRemoveC, &readAfterRemoveReqBody)

	assert.Equal(t, http.StatusNotFound, readAfterRemoveW.Code)

	var readAfterRemoveResponse models.Response
	err = json.Unmarshal(readAfterRemoveW.Body.Bytes(), &readAfterRemoveResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, readAfterRemoveResponse.ErrorCode)
}