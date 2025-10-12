package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
	"helixtrack.ru/core/internal/websocket"
)

// Handler manages all HTTP handlers
type Handler struct {
	db          database.Database
	authService services.AuthService
	permService services.PermissionService
	version     string
	publisher   websocket.EventPublisher
}

// NewHandler creates a new handler instance
func NewHandler(db database.Database, authService services.AuthService, permService services.PermissionService, version string) *Handler {
	return &Handler{
		db:          db,
		authService: authService,
		permService: permService,
		version:     version,
		publisher:   websocket.NewNoOpPublisher(), // Default to no-op publisher
	}
}

// SetEventPublisher sets the event publisher for the handler
func (h *Handler) SetEventPublisher(publisher websocket.EventPublisher) {
	h.publisher = publisher
}

// DoAction handles the unified /do endpoint with action-based routing
func (h *Handler) DoAction(c *gin.Context) {
	// Get the already-parsed request from context (set by server.go)
	reqInterface, exists := c.Get("request")
	if !exists {
		logger.Error("Request not found in context")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid request format",
			"",
		))
		return
	}

	req, ok := reqInterface.(*models.Request)
	if !ok {
		logger.Error("Invalid request type in context")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid request format",
			"",
		))
		return
	}

	logger.Info("Processing action",
		zap.String("action", req.Action),
		zap.String("object", req.Object),
	)

	// Route to appropriate handler based on action
	switch req.Action {
	// System actions
	case models.ActionVersion:
		h.handleVersion(c, req)
	case models.ActionJWTCapable:
		h.handleJWTCapable(c, req)
	case models.ActionDBCapable:
		h.handleDBCapable(c, req)
	case models.ActionHealth:
		h.handleHealth(c, req)
	case models.ActionAuthenticate:
		h.handleAuthenticate(c, req)

	// Generic CRUD actions
	case models.ActionCreate:
		h.handleCreate(c, req)
	case models.ActionModify:
		h.handleModify(c, req)
	case models.ActionRemove:
		h.handleRemove(c, req)
	case models.ActionRead:
		h.handleRead(c, req)
	case models.ActionList:
		h.handleList(c, req)

	// Priority actions
	case models.ActionPriorityCreate:
		h.handlePriorityCreate(c, req)
	case models.ActionPriorityRead:
		h.handlePriorityRead(c, req)
	case models.ActionPriorityList:
		h.handlePriorityList(c, req)
	case models.ActionPriorityModify:
		h.handlePriorityModify(c, req)
	case models.ActionPriorityRemove:
		h.handlePriorityRemove(c, req)

	// Resolution actions
	case models.ActionResolutionCreate:
		h.handleResolutionCreate(c, req)
	case models.ActionResolutionRead:
		h.handleResolutionRead(c, req)
	case models.ActionResolutionList:
		h.handleResolutionList(c, req)
	case models.ActionResolutionModify:
		h.handleResolutionModify(c, req)
	case models.ActionResolutionRemove:
		h.handleResolutionRemove(c, req)

	// Watcher actions
	case models.ActionWatcherAdd:
		h.handleWatcherAdd(c, req)
	case models.ActionWatcherRemove:
		h.handleWatcherRemove(c, req)
	case models.ActionWatcherList:
		h.handleWatcherList(c, req)

	// Version actions
	case models.ActionVersionCreate:
		h.handleVersionCreate(c, req)
	case models.ActionVersionRead:
		h.handleVersionRead(c, req)
	case models.ActionVersionList:
		h.handleVersionList(c, req)
	case models.ActionVersionModify:
		h.handleVersionModify(c, req)
	case models.ActionVersionRemove:
		h.handleVersionRemove(c, req)
	case models.ActionVersionRelease:
		h.handleVersionRelease(c, req)
	case models.ActionVersionArchive:
		h.handleVersionArchive(c, req)
	case models.ActionVersionAddAffected:
		h.handleVersionAddAffected(c, req)
	case models.ActionVersionRemoveAffected:
		h.handleVersionRemoveAffected(c, req)
	case models.ActionVersionListAffected:
		h.handleVersionListAffected(c, req)
	case models.ActionVersionAddFix:
		h.handleVersionAddFix(c, req)
	case models.ActionVersionRemoveFix:
		h.handleVersionRemoveFix(c, req)
	case models.ActionVersionListFix:
		h.handleVersionListFix(c, req)

	// Filter actions
	case models.ActionFilterSave:
		h.handleFilterSave(c, req)
	case models.ActionFilterLoad:
		h.handleFilterLoad(c, req)
	case models.ActionFilterList:
		h.handleFilterList(c, req)
	case models.ActionFilterShare:
		h.handleFilterShare(c, req)
	case models.ActionFilterModify:
		h.handleFilterModify(c, req)
	case models.ActionFilterRemove:
		h.handleFilterRemove(c, req)

	// Custom field actions
	case models.ActionCustomFieldCreate:
		h.handleCustomFieldCreate(c, req)
	case models.ActionCustomFieldRead:
		h.handleCustomFieldRead(c, req)
	case models.ActionCustomFieldList:
		h.handleCustomFieldList(c, req)
	case models.ActionCustomFieldModify:
		h.handleCustomFieldModify(c, req)
	case models.ActionCustomFieldRemove:
		h.handleCustomFieldRemove(c, req)

	// Custom field option actions
	case models.ActionCustomFieldOptionCreate:
		h.handleCustomFieldOptionCreate(c, req)
	case models.ActionCustomFieldOptionModify:
		h.handleCustomFieldOptionModify(c, req)
	case models.ActionCustomFieldOptionRemove:
		h.handleCustomFieldOptionRemove(c, req)
	case models.ActionCustomFieldOptionList:
		h.handleCustomFieldOptionList(c, req)

	// Custom field value actions
	case models.ActionCustomFieldValueSet:
		h.handleCustomFieldValueSet(c, req)
	case models.ActionCustomFieldValueGet:
		h.handleCustomFieldValueGet(c, req)
	case models.ActionCustomFieldValueList:
		h.handleCustomFieldValueList(c, req)
	case models.ActionCustomFieldValueRemove:
		h.handleCustomFieldValueRemove(c, req)

	// Board actions
	case models.ActionBoardCreate:
		h.handleBoardCreate(c, req)
	case models.ActionBoardRead:
		h.handleBoardRead(c, req)
	case models.ActionBoardList:
		h.handleBoardList(c, req)
	case models.ActionBoardModify:
		h.handleBoardModify(c, req)
	case models.ActionBoardRemove:
		h.handleBoardRemove(c, req)

	// Board ticket assignment
	case models.ActionBoardAddTicket:
		h.handleBoardAddTicket(c, req)
	case models.ActionBoardRemoveTicket:
		h.handleBoardRemoveTicket(c, req)
	case models.ActionBoardListTickets:
		h.handleBoardListTickets(c, req)

	// Board metadata
	case models.ActionBoardSetMetadata:
		h.handleBoardSetMetadata(c, req)
	case models.ActionBoardGetMetadata:
		h.handleBoardGetMetadata(c, req)
	case models.ActionBoardListMetadata:
		h.handleBoardListMetadata(c, req)
	case models.ActionBoardRemoveMetadata:
		h.handleBoardRemoveMetadata(c, req)

	// Cycle actions
	case models.ActionCycleCreate:
		h.handleCycleCreate(c, req)
	case models.ActionCycleRead:
		h.handleCycleRead(c, req)
	case models.ActionCycleList:
		h.handleCycleList(c, req)
	case models.ActionCycleModify:
		h.handleCycleModify(c, req)
	case models.ActionCycleRemove:
		h.handleCycleRemove(c, req)

	// Cycle-project mapping
	case models.ActionCycleAssignProject:
		h.handleCycleAssignProject(c, req)
	case models.ActionCycleUnassignProject:
		h.handleCycleUnassignProject(c, req)
	case models.ActionCycleListProjects:
		h.handleCycleListProjects(c, req)

	// Cycle-ticket mapping
	case models.ActionCycleAddTicket:
		h.handleCycleAddTicket(c, req)
	case models.ActionCycleRemoveTicket:
		h.handleCycleRemoveTicket(c, req)
	case models.ActionCycleListTickets:
		h.handleCycleListTickets(c, req)

	// Workflow actions
	case models.ActionWorkflowCreate:
		h.handleWorkflowCreate(c, req)
	case models.ActionWorkflowRead:
		h.handleWorkflowRead(c, req)
	case models.ActionWorkflowList:
		h.handleWorkflowList(c, req)
	case models.ActionWorkflowModify:
		h.handleWorkflowModify(c, req)
	case models.ActionWorkflowRemove:
		h.handleWorkflowRemove(c, req)

	// Workflow step actions
	case models.ActionWorkflowStepCreate:
		h.handleWorkflowStepCreate(c, req)
	case models.ActionWorkflowStepRead:
		h.handleWorkflowStepRead(c, req)
	case models.ActionWorkflowStepList:
		h.handleWorkflowStepList(c, req)
	case models.ActionWorkflowStepModify:
		h.handleWorkflowStepModify(c, req)
	case models.ActionWorkflowStepRemove:
		h.handleWorkflowStepRemove(c, req)

	// Ticket status actions
	case models.ActionTicketStatusCreate:
		h.handleTicketStatusCreate(c, req)
	case models.ActionTicketStatusRead:
		h.handleTicketStatusRead(c, req)
	case models.ActionTicketStatusList:
		h.handleTicketStatusList(c, req)
	case models.ActionTicketStatusModify:
		h.handleTicketStatusModify(c, req)
	case models.ActionTicketStatusRemove:
		h.handleTicketStatusRemove(c, req)

	// Ticket type actions
	case models.ActionTicketTypeCreate:
		h.handleTicketTypeCreate(c, req)
	case models.ActionTicketTypeRead:
		h.handleTicketTypeRead(c, req)
	case models.ActionTicketTypeList:
		h.handleTicketTypeList(c, req)
	case models.ActionTicketTypeModify:
		h.handleTicketTypeModify(c, req)
	case models.ActionTicketTypeRemove:
		h.handleTicketTypeRemove(c, req)
	case models.ActionTicketTypeAssign:
		h.handleTicketTypeAssign(c, req)
	case models.ActionTicketTypeUnassign:
		h.handleTicketTypeUnassign(c, req)
	case models.ActionTicketTypeListByProject:
		h.handleTicketTypeListByProject(c, req)

	// Account actions (Multi-tenancy support)
	case models.ActionAccountCreate:
		h.AccountCreate(c, req)
	case models.ActionAccountRead:
		h.AccountRead(c, req)
	case models.ActionAccountList:
		h.AccountList(c, req)
	case models.ActionAccountModify:
		h.AccountModify(c, req)
	case models.ActionAccountRemove:
		h.AccountRemove(c, req)

	// Organization actions
	case models.ActionOrganizationCreate:
		h.OrganizationCreate(c, req)
	case models.ActionOrganizationRead:
		h.OrganizationRead(c, req)
	case models.ActionOrganizationList:
		h.OrganizationList(c, req)
	case models.ActionOrganizationModify:
		h.OrganizationModify(c, req)
	case models.ActionOrganizationRemove:
		h.OrganizationRemove(c, req)
	case models.ActionOrganizationAssignAccount:
		h.OrganizationAssignAccount(c, req)
	case models.ActionOrganizationListAccounts:
		h.OrganizationListAccounts(c, req)

	// Team actions
	case models.ActionTeamCreate:
		h.TeamCreate(c, req)
	case models.ActionTeamRead:
		h.TeamRead(c, req)
	case models.ActionTeamList:
		h.TeamList(c, req)
	case models.ActionTeamModify:
		h.TeamModify(c, req)
	case models.ActionTeamRemove:
		h.TeamRemove(c, req)
	case models.ActionTeamAssignOrganization:
		h.TeamAssignOrganization(c, req)
	case models.ActionTeamUnassignOrganization:
		h.TeamUnassignOrganization(c, req)
	case models.ActionTeamListOrganizations:
		h.TeamListOrganizations(c, req)
	case models.ActionTeamAssignProject:
		h.TeamAssignProject(c, req)
	case models.ActionTeamUnassignProject:
		h.TeamUnassignProject(c, req)
	case models.ActionTeamListProjects:
		h.TeamListProjects(c, req)

	// User-Organization mapping
	case models.ActionUserAssignOrganization:
		h.UserAssignOrganization(c, req)
	case models.ActionUserListOrganizations:
		h.UserListOrganizations(c, req)
	case models.ActionOrganizationListUsers:
		h.OrganizationListUsers(c, req)

	// User-Team mapping
	case models.ActionUserAssignTeam:
		h.UserAssignTeam(c, req)
	case models.ActionUserListTeams:
		h.UserListTeams(c, req)
	case models.ActionTeamListUsers:
		h.TeamListUsers(c, req)

	// Component actions
	case models.ActionComponentCreate:
		h.handleComponentCreate(c, req)
	case models.ActionComponentRead:
		h.handleComponentRead(c, req)
	case models.ActionComponentList:
		h.handleComponentList(c, req)
	case models.ActionComponentModify:
		h.handleComponentModify(c, req)
	case models.ActionComponentRemove:
		h.handleComponentRemove(c, req)

	// Component-ticket mapping
	case models.ActionComponentAddTicket:
		h.handleComponentAddTicket(c, req)
	case models.ActionComponentRemoveTicket:
		h.handleComponentRemoveTicket(c, req)
	case models.ActionComponentListTickets:
		h.handleComponentListTickets(c, req)

	// Component metadata
	case models.ActionComponentSetMetadata:
		h.handleComponentSetMetadata(c, req)
	case models.ActionComponentGetMetadata:
		h.handleComponentGetMetadata(c, req)
	case models.ActionComponentListMetadata:
		h.handleComponentListMetadata(c, req)
	case models.ActionComponentRemoveMetadata:
		h.handleComponentRemoveMetadata(c, req)

	// Label actions
	case models.ActionLabelCreate:
		h.handleLabelCreate(c, req)
	case models.ActionLabelRead:
		h.handleLabelRead(c, req)
	case models.ActionLabelList:
		h.handleLabelList(c, req)
	case models.ActionLabelModify:
		h.handleLabelModify(c, req)
	case models.ActionLabelRemove:
		h.handleLabelRemove(c, req)

	// Label category actions
	case models.ActionLabelCategoryCreate:
		h.handleLabelCategoryCreate(c, req)
	case models.ActionLabelCategoryRead:
		h.handleLabelCategoryRead(c, req)
	case models.ActionLabelCategoryList:
		h.handleLabelCategoryList(c, req)
	case models.ActionLabelCategoryModify:
		h.handleLabelCategoryModify(c, req)
	case models.ActionLabelCategoryRemove:
		h.handleLabelCategoryRemove(c, req)

	// Label-ticket mapping
	case models.ActionLabelAddTicket:
		h.handleLabelAddTicket(c, req)
	case models.ActionLabelRemoveTicket:
		h.handleLabelRemoveTicket(c, req)
	case models.ActionLabelListTickets:
		h.handleLabelListTickets(c, req)

	// Label-category mapping
	case models.ActionLabelAssignCategory:
		h.handleLabelAssignCategory(c, req)
	case models.ActionLabelUnassignCategory:
		h.handleLabelUnassignCategory(c, req)
	case models.ActionLabelListCategories:
		h.handleLabelListCategories(c, req)

	// Asset actions
	case models.ActionAssetCreate:
		h.handleAssetCreate(c, req)
	case models.ActionAssetRead:
		h.handleAssetRead(c, req)
	case models.ActionAssetList:
		h.handleAssetList(c, req)
	case models.ActionAssetModify:
		h.handleAssetModify(c, req)
	case models.ActionAssetRemove:
		h.handleAssetRemove(c, req)

	// Asset-ticket mapping
	case models.ActionAssetAddTicket:
		h.handleAssetAddTicket(c, req)
	case models.ActionAssetRemoveTicket:
		h.handleAssetRemoveTicket(c, req)
	case models.ActionAssetListTickets:
		h.handleAssetListTickets(c, req)

	// Asset-comment mapping
	case models.ActionAssetAddComment:
		h.handleAssetAddComment(c, req)
	case models.ActionAssetRemoveComment:
		h.handleAssetRemoveComment(c, req)
	case models.ActionAssetListComments:
		h.handleAssetListComments(c, req)

	// Asset-project mapping
	case models.ActionAssetAddProject:
		h.handleAssetAddProject(c, req)
	case models.ActionAssetRemoveProject:
		h.handleAssetRemoveProject(c, req)
	case models.ActionAssetListProjects:
		h.handleAssetListProjects(c, req)

	// Repository actions
	case models.ActionRepositoryCreate:
		h.handleRepositoryCreate(c, req)
	case models.ActionRepositoryRead:
		h.handleRepositoryRead(c, req)
	case models.ActionRepositoryList:
		h.handleRepositoryList(c, req)
	case models.ActionRepositoryModify:
		h.handleRepositoryModify(c, req)
	case models.ActionRepositoryRemove:
		h.handleRepositoryRemove(c, req)

	// Repository type actions
	case models.ActionRepositoryTypeCreate:
		h.handleRepositoryTypeCreate(c, req)
	case models.ActionRepositoryTypeRead:
		h.handleRepositoryTypeRead(c, req)
	case models.ActionRepositoryTypeList:
		h.handleRepositoryTypeList(c, req)
	case models.ActionRepositoryTypeModify:
		h.handleRepositoryTypeModify(c, req)
	case models.ActionRepositoryTypeRemove:
		h.handleRepositoryTypeRemove(c, req)

	// Repository-project mapping
	case models.ActionRepositoryAssignProject:
		h.handleRepositoryAssignProject(c, req)
	case models.ActionRepositoryUnassignProject:
		h.handleRepositoryUnassignProject(c, req)
	case models.ActionRepositoryListProjects:
		h.handleRepositoryListProjects(c, req)

	// Repository-commit-ticket mapping
	case models.ActionRepositoryAddCommit:
		h.handleRepositoryAddCommit(c, req)
	case models.ActionRepositoryRemoveCommit:
		h.handleRepositoryRemoveCommit(c, req)
	case models.ActionRepositoryListCommits:
		h.handleRepositoryListCommits(c, req)
	case models.ActionRepositoryGetCommit:
		h.handleRepositoryGetCommit(c, req)

	// Ticket relationship type actions
	case models.ActionTicketRelationshipTypeCreate:
		h.handleTicketRelationshipTypeCreate(c, req)
	case models.ActionTicketRelationshipTypeRead:
		h.handleTicketRelationshipTypeRead(c, req)
	case models.ActionTicketRelationshipTypeList:
		h.handleTicketRelationshipTypeList(c, req)
	case models.ActionTicketRelationshipTypeModify:
		h.handleTicketRelationshipTypeModify(c, req)
	case models.ActionTicketRelationshipTypeRemove:
		h.handleTicketRelationshipTypeRemove(c, req)

	// Ticket relationship actions
	case models.ActionTicketRelationshipCreate:
		h.handleTicketRelationshipCreate(c, req)
	case models.ActionTicketRelationshipRemove:
		h.handleTicketRelationshipRemove(c, req)
	case models.ActionTicketRelationshipList:
		h.handleTicketRelationshipList(c, req)

	// Vote actions (Phase 3)
	case models.ActionVoteAdd:
		h.handleVoteAdd(c, req)
	case models.ActionVoteRemove:
		h.handleVoteRemove(c, req)
	case models.ActionVoteCount:
		h.handleVoteCount(c, req)
	case models.ActionVoteList:
		h.handleVoteList(c, req)
	case models.ActionVoteCheck:
		h.handleVoteCheck(c, req)

	// Project Category actions (Phase 3)
	case models.ActionProjectCategoryCreate:
		h.handleProjectCategoryCreate(c, req)
	case models.ActionProjectCategoryRead:
		h.handleProjectCategoryRead(c, req)
	case models.ActionProjectCategoryList:
		h.handleProjectCategoryList(c, req)
	case models.ActionProjectCategoryModify:
		h.handleProjectCategoryModify(c, req)
	case models.ActionProjectCategoryRemove:
		h.handleProjectCategoryRemove(c, req)
	case models.ActionProjectCategoryAssign:
		h.handleProjectCategoryAssign(c, req)

	// Work Log actions (Phase 2)
	case models.ActionWorkLogAdd:
		h.handleWorkLogAdd(c, req)
	case models.ActionWorkLogModify:
		h.handleWorkLogModify(c, req)
	case models.ActionWorkLogRemove:
		h.handleWorkLogRemove(c, req)
	case models.ActionWorkLogList:
		h.handleWorkLogList(c, req)
	case models.ActionWorkLogListByTicket:
		h.handleWorkLogListByTicket(c, req)
	case models.ActionWorkLogListByUser:
		h.handleWorkLogListByUser(c, req)
	case models.ActionWorkLogGetTotalTime:
		h.handleWorkLogGetTotalTime(c, req)

	// Epic actions (Phase 2)
	case models.ActionEpicCreate:
		h.handleEpicCreate(c, req)
	case models.ActionEpicRead:
		h.handleEpicRead(c, req)
	case models.ActionEpicList:
		h.handleEpicList(c, req)
	case models.ActionEpicModify:
		h.handleEpicModify(c, req)
	case models.ActionEpicRemove:
		h.handleEpicRemove(c, req)
	case models.ActionEpicAddStory:
		h.handleEpicAddStory(c, req)
	case models.ActionEpicRemoveStory:
		h.handleEpicRemoveStory(c, req)
	case models.ActionEpicListStories:
		h.handleEpicListStories(c, req)

	// Subtask actions (Phase 2)
	case models.ActionSubtaskCreate:
		h.handleSubtaskCreate(c, req)
	case models.ActionSubtaskList:
		h.handleSubtaskList(c, req)
	case models.ActionSubtaskMoveToParent:
		h.handleSubtaskMoveToParent(c, req)
	case models.ActionSubtaskConvertToIssue:
		h.handleSubtaskConvertToIssue(c, req)
	case models.ActionSubtaskListByParent:
		h.handleSubtaskListByParent(c, req)

	// Project Role actions (Phase 2)
	case models.ActionProjectRoleCreate:
		h.handleProjectRoleCreate(c, req)
	case models.ActionProjectRoleRead:
		h.handleProjectRoleRead(c, req)
	case models.ActionProjectRoleList:
		h.handleProjectRoleList(c, req)
	case models.ActionProjectRoleModify:
		h.handleProjectRoleModify(c, req)
	case models.ActionProjectRoleRemove:
		h.handleProjectRoleRemove(c, req)
	case models.ActionProjectRoleAssignUser:
		h.handleProjectRoleAssignUser(c, req)
	case models.ActionProjectRoleUnassignUser:
		h.handleProjectRoleUnassignUser(c, req)
	case models.ActionProjectRoleListUsers:
		h.handleProjectRoleListUsers(c, req)

	// Security Level actions (Phase 2)
	case models.ActionSecurityLevelCreate:
		h.handleSecurityLevelCreate(c, req)
	case models.ActionSecurityLevelRead:
		h.handleSecurityLevelRead(c, req)
	case models.ActionSecurityLevelList:
		h.handleSecurityLevelList(c, req)
	case models.ActionSecurityLevelModify:
		h.handleSecurityLevelModify(c, req)
	case models.ActionSecurityLevelRemove:
		h.handleSecurityLevelRemove(c, req)
	case models.ActionSecurityLevelGrant:
		h.handleSecurityLevelGrant(c, req)
	case models.ActionSecurityLevelRevoke:
		h.handleSecurityLevelRevoke(c, req)
	case models.ActionSecurityLevelCheck:
		h.handleSecurityLevelCheck(c, req)

	// Dashboard actions (Phase 2)
	case models.ActionDashboardCreate:
		h.handleDashboardCreate(c, req)
	case models.ActionDashboardRead:
		h.handleDashboardRead(c, req)
	case models.ActionDashboardList:
		h.handleDashboardList(c, req)
	case models.ActionDashboardModify:
		h.handleDashboardModify(c, req)
	case models.ActionDashboardRemove:
		h.handleDashboardRemove(c, req)
	case models.ActionDashboardShare:
		h.handleDashboardShare(c, req)
	case models.ActionDashboardUnshare:
		h.handleDashboardUnshare(c, req)
	case models.ActionDashboardAddWidget:
		h.handleDashboardAddWidget(c, req)
	case models.ActionDashboardRemoveWidget:
		h.handleDashboardRemoveWidget(c, req)
	case models.ActionDashboardModifyWidget:
		h.handleDashboardModifyWidget(c, req)
	case models.ActionDashboardListWidgets:
		h.handleDashboardListWidgets(c, req)
	case models.ActionDashboardSetLayout:
		h.handleDashboardSetLayout(c, req)

	// Board Config actions (Phase 2)
	case models.ActionBoardConfigureColumns:
		h.handleBoardConfigureColumns(c, req)
	case models.ActionBoardAddColumn:
		h.handleBoardAddColumn(c, req)
	case models.ActionBoardRemoveColumn:
		h.handleBoardRemoveColumn(c, req)
	case models.ActionBoardModifyColumn:
		h.handleBoardModifyColumn(c, req)
	case models.ActionBoardListColumns:
		h.handleBoardListColumns(c, req)
	case models.ActionBoardAddSwimlane:
		h.handleBoardAddSwimlane(c, req)
	case models.ActionBoardRemoveSwimlane:
		h.handleBoardRemoveSwimlane(c, req)
	case models.ActionBoardListSwimlanes:
		h.handleBoardListSwimlanes(c, req)
	case models.ActionBoardAddQuickFilter:
		h.handleBoardAddQuickFilter(c, req)
	case models.ActionBoardRemoveQuickFilter:
		h.handleBoardRemoveQuickFilter(c, req)
	case models.ActionBoardListQuickFilters:
		h.handleBoardListQuickFilters(c, req)
	case models.ActionBoardSetType:
		h.handleBoardSetType(c, req)

	// Notification actions (Phase 2)
	case models.ActionNotificationSchemeCreate:
		h.handleNotificationSchemeCreate(c, req)
	case models.ActionNotificationSchemeRead:
		h.handleNotificationSchemeRead(c, req)
	case models.ActionNotificationSchemeList:
		h.handleNotificationSchemeList(c, req)
	case models.ActionNotificationSchemeModify:
		h.handleNotificationSchemeModify(c, req)
	case models.ActionNotificationSchemeRemove:
		h.handleNotificationSchemeRemove(c, req)
	case models.ActionNotificationSchemeAddRule:
		h.handleNotificationSchemeAddRule(c, req)
	case models.ActionNotificationSchemeRemoveRule:
		h.handleNotificationSchemeRemoveRule(c, req)
	case models.ActionNotificationSchemeListRules:
		h.handleNotificationSchemeListRules(c, req)
	case models.ActionNotificationEventList:
		h.handleNotificationEventList(c, req)
	case models.ActionNotificationSend:
		h.handleNotificationSend(c, req)

	// Activity Stream actions (Phase 3)
	case models.ActionActivityStreamGet:
		h.handleActivityStreamGet(c, req)
	case models.ActionActivityStreamGetByProject:
		h.handleActivityStreamGetByProject(c, req)
	case models.ActionActivityStreamGetByUser:
		h.handleActivityStreamGetByUser(c, req)
	case models.ActionActivityStreamGetByTicket:
		h.handleActivityStreamGetByTicket(c, req)
	case models.ActionActivityStreamFilter:
		h.handleActivityStreamFilter(c, req)

	// Comment Mention actions (Phase 3)
	case models.ActionCommentMention:
		h.handleCommentMention(c, req)
	case models.ActionCommentUnmention:
		h.handleCommentUnmention(c, req)
	case models.ActionCommentListMentions:
		h.handleCommentListMentions(c, req)
	case models.ActionCommentGetMentions:
		h.handleCommentGetMentions(c, req)
	case models.ActionCommentParseMentions:
		h.handleCommentParseMentions(c, req)

	default:
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidAction,
			"Unknown action: "+req.Action,
			"",
		))
	}
}

// handleVersion returns the API version
func (h *Handler) handleVersion(c *gin.Context, req *models.Request) {
	response := models.NewSuccessResponse(map[string]interface{}{
		"version": h.version,
		"api":     "1.0.0",
	})
	c.JSON(http.StatusOK, response)
}

// handleJWTCapable returns whether JWT authentication is available
func (h *Handler) handleJWTCapable(c *gin.Context, req *models.Request) {
	capable := h.authService != nil && h.authService.IsEnabled()
	response := models.NewSuccessResponse(map[string]interface{}{
		"jwtCapable": capable,
		"enabled":    capable,
	})
	c.JSON(http.StatusOK, response)
}

// handleDBCapable returns whether database is available
func (h *Handler) handleDBCapable(c *gin.Context, req *models.Request) {
	capable := h.db != nil
	dbType := ""
	if h.db != nil {
		dbType = h.db.GetType()
		// Try to ping the database
		if err := h.db.Ping(c.Request.Context()); err != nil {
			logger.Error("Database ping failed", zap.Error(err))
			capable = false
		}
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"dbCapable": capable,
		"type":      dbType,
	})
	c.JSON(http.StatusOK, response)
}

// handleHealth returns the health status of the service
func (h *Handler) handleHealth(c *gin.Context, req *models.Request) {
	healthy := true
	checks := make(map[string]interface{})

	// Check database
	if h.db != nil {
		if err := h.db.Ping(c.Request.Context()); err != nil {
			checks["database"] = "unhealthy"
			healthy = false
		} else {
			checks["database"] = "healthy"
		}
	}

	// Check auth service
	if h.authService != nil && h.authService.IsEnabled() {
		checks["authService"] = "enabled"
	} else {
		checks["authService"] = "disabled"
	}

	// Check permission service
	if h.permService != nil && h.permService.IsEnabled() {
		checks["permissionService"] = "enabled"
	} else {
		checks["permissionService"] = "disabled"
	}

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"status": status,
		"checks": checks,
	})

	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
		response.ErrorCode = models.ErrorCodeServiceUnavailable
		response.ErrorMessage = "Service is unhealthy"
	}

	c.JSON(statusCode, response)
}

// handleAuthenticate handles authentication requests
func (h *Handler) handleAuthenticate(c *gin.Context, req *models.Request) {
	username, ok := req.Data["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing username",
			"",
		))
		return
	}

	password, ok := req.Data["password"].(string)
	if !ok || password == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing password",
			"",
		))
		return
	}

	// Try external auth service first if enabled
	if h.authService != nil && h.authService.IsEnabled() {
		claims, err := h.authService.Authenticate(c.Request.Context(), username, password)
		if err != nil {
			logger.Error("Authentication failed", zap.Error(err), zap.String("username", username))
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrorCodeUnauthorized,
				"Authentication failed",
				"",
			))
			return
		}

		response := models.NewSuccessResponse(map[string]interface{}{
			"username": claims.Username,
			"role":     claims.Role,
			"name":     claims.Name,
		})
		c.JSON(http.StatusOK, response)
		return
	}

	// Fall back to local authentication (for testing)
	// Get user from database
	query := `
		SELECT id, username, password_hash, email, name, role, created_at, updated_at
		FROM users
		WHERE username = ? AND deleted = 0
	`

	var user models.User
	var createdAt, updatedAt int64

	err := h.db.QueryRow(context.Background(), query, username).Scan(
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
		logger.Error("User not found", zap.Error(err), zap.String("username", username))
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
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		logger.Error("Invalid password", zap.String("username", username))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid username or password",
			"",
		))
		return
	}

	// Generate JWT token
	jwtService := services.NewJWTService("", "", 24)
	token, err := jwtService.GenerateToken(user.Username, user.Email, user.Name, user.Role)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to generate authentication token",
			"",
		))
		return
	}

	// Return success response with token
	response := models.NewSuccessResponse(map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
	})
	c.JSON(http.StatusOK, response)
}

// handleCreate handles create operations
func (h *Handler) handleCreate(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	// Get username from middleware
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionCreate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleCreateProject(c, req)
	case "ticket":
		h.handleCreateTicket(c, req)
	case "comment":
		h.handleCreateComment(c, req)
	default:
		logger.Info("Create operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}

// handleModify handles modify operations
func (h *Handler) handleModify(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionUpdate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleModifyProject(c, req)
	case "ticket":
		h.handleModifyTicket(c, req)
	case "comment":
		h.handleModifyComment(c, req)
	default:
		logger.Info("Modify operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}

// handleRemove handles remove operations
func (h *Handler) handleRemove(c *gin.Context, req *models.Request) {
	if req.Object == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingObject,
			"Missing object type",
			"",
		))
		return
	}

	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, req.Object, models.PermissionDelete)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	// If permission service is disabled, check user role from JWT claims
	if !h.permService.IsEnabled() {
		if claims, exists := middleware.GetClaims(c); exists {
			// Viewer role cannot delete
			if username == "viewer" || claims.Role == "viewer" {
				allowed = false
			}
		}
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission - forbidden",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleRemoveProject(c, req)
	case "ticket":
		h.handleRemoveTicket(c, req)
	case "comment":
		h.handleRemoveComment(c, req)
	default:
		logger.Info("Remove operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}

// handleRead handles read operations
func (h *Handler) handleRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleReadProject(c, req)
	case "ticket":
		h.handleReadTicket(c, req)
	case "comment":
		h.handleReadComment(c, req)
	default:
		logger.Info("Read operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}

// handleList handles list operations
func (h *Handler) handleList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Route to specific handler based on object type
	switch req.Object {
	case "project":
		h.handleListProjects(c, req)
	case "ticket":
		h.handleListTickets(c, req)
	case "comment":
		h.handleListComments(c, req)
	default:
		logger.Info("List operation for unsupported object",
			zap.String("object", req.Object),
			zap.String("username", username),
		)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidObject,
			"Unsupported object type: "+req.Object,
			"",
		))
	}
}
