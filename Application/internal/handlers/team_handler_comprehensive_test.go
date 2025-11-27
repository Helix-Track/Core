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

// TestTeamHandler_Comprehensive tests the complete team lifecycle with mappings
func TestTeamHandler_Comprehensive(t *testing.T) {
	handler := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Step 1: Create an organization
	orgReqBody := models.Request{
		Action: models.ActionOrganizationCreate,
		Data: map[string]interface{}{
			"title":       "Test Organization",
			"description": "Test organization for team testing",
		},
	}

	orgBody, _ := json.Marshal(orgReqBody)
	orgReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(orgBody))
	orgReq.Header.Set("Content-Type", "application/json")
	orgW := httptest.NewRecorder()

	orgC, _ := gin.CreateTestContext(orgW)
	orgC.Request = orgReq

	handler.OrganizationCreate(orgC, &orgReqBody)

	assert.Equal(t, http.StatusOK, orgW.Code)

	var orgResponse models.Response
	err := json.Unmarshal(orgW.Body.Bytes(), &orgResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, orgResponse.ErrorCode)
	assert.NotNil(t, orgResponse.Data)

	// Extract organization ID
	orgDataMap := orgResponse.Data
	orgData, ok := orgDataMap["organization"].(map[string]interface{})
	require.True(t, ok)
	orgID := orgData["id"].(string)
	require.NotEmpty(t, orgID)

	// Step 2: Create a team
	teamReqBody := models.Request{
		Action: models.ActionTeamCreate,
		Data: map[string]interface{}{
			"title":       "Test Team",
			"description": "Test team for comprehensive testing",
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
	err = json.Unmarshal(teamW.Body.Bytes(), &teamResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, teamResponse.ErrorCode)
	assert.NotNil(t, teamResponse.Data)

	// Extract team ID
	teamDataMap := teamResponse.Data
	teamData, ok := teamDataMap["team"].(map[string]interface{})
	require.True(t, ok)
	teamID := teamData["id"].(string)
	require.NotEmpty(t, teamID)

	// Step 3: Assign team to organization
	assignReqBody := models.Request{
		Action: models.ActionTeamAssignOrganization,
		Data: map[string]interface{}{
			"teamId":         teamID,
			"organizationId": orgID,
		},
	}

	assignBody, _ := json.Marshal(assignReqBody)
	assignReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(assignBody))
	assignReq.Header.Set("Content-Type", "application/json")
	assignW := httptest.NewRecorder()

	assignC, _ := gin.CreateTestContext(assignW)
	assignC.Request = assignReq

	handler.TeamAssignOrganization(assignC, &assignReqBody)

	assert.Equal(t, http.StatusOK, assignW.Code)

	var assignResponse models.Response
	err = json.Unmarshal(assignW.Body.Bytes(), &assignResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, assignResponse.ErrorCode)

	// Step 4: List organizations for team
	listOrgsReqBody := models.Request{
		Action: models.ActionTeamListOrganizations,
		Data: map[string]interface{}{
			"teamId": teamID,
		},
	}

	listOrgsBody, _ := json.Marshal(listOrgsReqBody)
	listOrgsReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(listOrgsBody))
	listOrgsReq.Header.Set("Content-Type", "application/json")
	listOrgsW := httptest.NewRecorder()

	listOrgsC, _ := gin.CreateTestContext(listOrgsW)
	listOrgsC.Request = listOrgsReq

	handler.TeamListOrganizations(listOrgsC, &listOrgsReqBody)

	assert.Equal(t, http.StatusOK, listOrgsW.Code)

	var listOrgsResponse models.Response
	err = json.Unmarshal(listOrgsW.Body.Bytes(), &listOrgsResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, listOrgsResponse.ErrorCode)
	assert.NotNil(t, listOrgsResponse.Data)

	// Verify organizations list contains our organization
	listOrgsDataMap := listOrgsResponse.Data
	organizations, ok := listOrgsDataMap["organizations"].([]interface{})
	require.True(t, ok)
	assert.Len(t, organizations, 1)

	// Step 5: Unassign team from organization
	unassignReqBody := models.Request{
		Action: models.ActionTeamUnassignOrganization,
		Data: map[string]interface{}{
			"teamId":         teamID,
			"organizationId": orgID,
		},
	}

	unassignBody, _ := json.Marshal(unassignReqBody)
	unassignReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(unassignBody))
	unassignReq.Header.Set("Content-Type", "application/json")
	unassignW := httptest.NewRecorder()

	unassignC, _ := gin.CreateTestContext(unassignW)
	unassignC.Request = unassignReq

	handler.TeamUnassignOrganization(unassignC, &unassignReqBody)

	assert.Equal(t, http.StatusOK, unassignW.Code)

	var unassignResponse models.Response
	err = json.Unmarshal(unassignW.Body.Bytes(), &unassignResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, unassignResponse.ErrorCode)

	// Step 6: Verify organizations list is empty after unassignment
	listOrgsAfterReqBody := models.Request{
		Action: models.ActionTeamListOrganizations,
		Data: map[string]interface{}{
			"teamId": teamID,
		},
	}

	listOrgsAfterBody, _ := json.Marshal(listOrgsAfterReqBody)
	listOrgsAfterReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(listOrgsAfterBody))
	listOrgsAfterReq.Header.Set("Content-Type", "application/json")
	listOrgsAfterW := httptest.NewRecorder()

	listOrgsAfterC, _ := gin.CreateTestContext(listOrgsAfterW)
	listOrgsAfterC.Request = listOrgsAfterReq

	handler.TeamListOrganizations(listOrgsAfterC, &listOrgsAfterReqBody)

	assert.Equal(t, http.StatusOK, listOrgsAfterW.Code)

	var listOrgsAfterResponse models.Response
	err = json.Unmarshal(listOrgsAfterW.Body.Bytes(), &listOrgsAfterResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, listOrgsAfterResponse.ErrorCode)
	assert.NotNil(t, listOrgsAfterResponse.Data)

	// Verify organizations list is empty
	listOrgsAfterDataMap := listOrgsAfterResponse.Data
	organizationsAfter, ok := listOrgsAfterDataMap["organizations"].([]interface{})
	require.True(t, ok)
	assert.Len(t, organizationsAfter, 0)

	// Step 7: Remove team
	removeTeamReqBody := models.Request{
		Action: models.ActionTeamRemove,
		Data: map[string]interface{}{
			"id": teamID,
		},
	}

	removeTeamBody, _ := json.Marshal(removeTeamReqBody)
	removeTeamReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(removeTeamBody))
	removeTeamReq.Header.Set("Content-Type", "application/json")
	removeTeamW := httptest.NewRecorder()

	removeTeamC, _ := gin.CreateTestContext(removeTeamW)
	removeTeamC.Request = removeTeamReq

	handler.TeamRemove(removeTeamC, &removeTeamReqBody)

	assert.Equal(t, http.StatusOK, removeTeamW.Code)

	var removeTeamResponse models.Response
	err = json.Unmarshal(removeTeamW.Body.Bytes(), &removeTeamResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, removeTeamResponse.ErrorCode)

	// Step 8: Remove organization
	removeOrgReqBody := models.Request{
		Action: models.ActionOrganizationRemove,
		Data: map[string]interface{}{
			"id": orgID,
		},
	}

	removeOrgBody, _ := json.Marshal(removeOrgReqBody)
	removeOrgReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(removeOrgBody))
	removeOrgReq.Header.Set("Content-Type", "application/json")
	removeOrgW := httptest.NewRecorder()

	removeOrgC, _ := gin.CreateTestContext(removeOrgW)
	removeOrgC.Request = removeOrgReq

	handler.OrganizationRemove(removeOrgC, &removeOrgReqBody)

	assert.Equal(t, http.StatusOK, removeOrgW.Code)

	var removeOrgResponse models.Response
	err = json.Unmarshal(removeOrgW.Body.Bytes(), &removeOrgResponse)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, removeOrgResponse.ErrorCode)
}