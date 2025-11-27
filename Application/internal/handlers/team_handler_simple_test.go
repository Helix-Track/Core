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

// TestTeamHandler_Simple tests basic team operations
func TestTeamHandler_Simple(t *testing.T) {
	handler := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Step 1: Create a team
	teamReqBody := models.Request{
		Action: models.ActionTeamCreate,
		Data: map[string]interface{}{
			"title":       "Simple Test Team",
			"description": "Test team for simple testing",
		},
	}

	teamBody, _ := json.Marshal(teamReqBody)
	teamReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(teamBody))
	teamReq.Header.Set("Content-Type", "application/json")
	teamW := httptest.NewRecorder()

	teamC, _ := gin.CreateTestContext(teamW)
	teamC.Request = teamReq

	handler.TeamCreate(teamC, &teamReqBody)

	assert.Equal(t, http.StatusOK, teamW.Code)

	var teamResponse models.Response
	err := json.Unmarshal(teamW.Body.Bytes(), &teamResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, teamResponse.ErrorCode)
	assert.NotNil(t, teamResponse.Data)

	// Extract team ID
	teamDataMap := teamResponse.Data
	teamData, ok := teamDataMap["team"].(map[string]interface{})
	require.True(t, ok)
	teamID := teamData["id"].(string)
	require.NotEmpty(t, teamID)

	// Step 2: Read the team
	readReqBody := models.Request{
		Action: models.ActionTeamRead,
		Data: map[string]interface{}{
			"id": teamID,
		},
	}

	readBody, _ := json.Marshal(readReqBody)
	readReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(readBody))
	readReq.Header.Set("Content-Type", "application/json")
	readW := httptest.NewRecorder()

	readC, _ := gin.CreateTestContext(readW)
	readC.Request = readReq

	handler.TeamRead(readC, &readReqBody)

	assert.Equal(t, http.StatusOK, readW.Code)

	var readResponse models.Response
	err = json.Unmarshal(readW.Body.Bytes(), &readResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, readResponse.ErrorCode)
	assert.NotNil(t, readResponse.Data)

	// Step 3: List teams
	listReqBody := models.Request{
		Action: models.ActionTeamList,
		Data:   map[string]interface{}{},
	}

	listBody, _ := json.Marshal(listReqBody)
	listReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(listBody))
	listReq.Header.Set("Content-Type", "application/json")
	listW := httptest.NewRecorder()

	listC, _ := gin.CreateTestContext(listW)
	listC.Request = listReq

	handler.TeamList(listC, &listReqBody)

	assert.Equal(t, http.StatusOK, listW.Code)

	var listResponse models.Response
	err = json.Unmarshal(listW.Body.Bytes(), &listResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, listResponse.ErrorCode)
	assert.NotNil(t, listResponse.Data)

	// Step 4: Remove the team
	removeReqBody := models.Request{
		Action: models.ActionTeamRemove,
		Data: map[string]interface{}{
			"id": teamID,
		},
	}

	removeBody, _ := json.Marshal(removeReqBody)
	removeReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(removeBody))
	removeReq.Header.Set("Content-Type", "application/json")
	removeW := httptest.NewRecorder()

	removeC, _ := gin.CreateTestContext(removeW)
	removeC.Request = removeReq

	handler.TeamRemove(removeC, &removeReqBody)

	assert.Equal(t, http.StatusOK, removeW.Code)

	var removeResponse models.Response
	err = json.Unmarshal(removeW.Body.Bytes(), &removeResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, removeResponse.ErrorCode)
}