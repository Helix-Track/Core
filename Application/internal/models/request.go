package models

// Request represents the unified API request format for the /do endpoint
type Request struct {
	Action string                 `json:"action" binding:"required"` // authenticate, version, jwtCapable, dbCapable, health, create, modify, remove
	JWT    string                 `json:"jwt"`                       // Required for authenticated actions
	Locale string                 `json:"locale"`                    // Optional locale for localized responses
	Object string                 `json:"object"`                    // Required for CRUD operations (create, modify, remove)
	Data   map[string]interface{} `json:"data"`                      // Additional data for the action
}

// IsAuthenticationRequired returns true if the action requires JWT authentication
func (r *Request) IsAuthenticationRequired() bool {
	switch r.Action {
	case ActionVersion, ActionJWTCapable, ActionDBCapable, ActionHealth, ActionAuthenticate:
		return false
	default:
		return true
	}
}

// IsCRUDOperation returns true if the action is a CRUD operation requiring an object
func (r *Request) IsCRUDOperation() bool {
	switch r.Action {
	case ActionCreate, ActionModify, ActionRemove:
		return true
	default:
		return false
	}
}

// Action constants
const (
	// System actions
	ActionAuthenticate = "authenticate"
	ActionVersion      = "version"
	ActionJWTCapable   = "jwtCapable"
	ActionDBCapable    = "dbCapable"
	ActionHealth       = "health"

	// Generic CRUD actions
	ActionCreate = "create"
	ActionModify = "modify"
	ActionRemove = "remove"
	ActionRead   = "read"
	ActionList   = "list"

	// Priority actions
	ActionPriorityCreate = "priorityCreate"
	ActionPriorityRead   = "priorityRead"
	ActionPriorityList   = "priorityList"
	ActionPriorityModify = "priorityModify"
	ActionPriorityRemove = "priorityRemove"

	// Resolution actions
	ActionResolutionCreate = "resolutionCreate"
	ActionResolutionRead   = "resolutionRead"
	ActionResolutionList   = "resolutionList"
	ActionResolutionModify = "resolutionModify"
	ActionResolutionRemove = "resolutionRemove"

	// Version actions
	ActionVersionCreate  = "versionCreate"
	ActionVersionRead    = "versionRead"
	ActionVersionList    = "versionList"
	ActionVersionModify  = "versionModify"
	ActionVersionRemove  = "versionRemove"
	ActionVersionRelease = "versionRelease" // Mark version as released
	ActionVersionArchive = "versionArchive" // Archive a version

	// Version mapping actions
	ActionVersionAddAffected    = "versionAddAffected"    // Add affected version to ticket
	ActionVersionRemoveAffected = "versionRemoveAffected" // Remove affected version from ticket
	ActionVersionListAffected   = "versionListAffected"   // List affected versions for ticket
	ActionVersionAddFix         = "versionAddFix"         // Add fix version to ticket
	ActionVersionRemoveFix      = "versionRemoveFix"      // Remove fix version from ticket
	ActionVersionListFix        = "versionListFix"        // List fix versions for ticket

	// Watcher actions
	ActionWatcherAdd    = "watcherAdd"    // Start watching a ticket
	ActionWatcherRemove = "watcherRemove" // Stop watching a ticket
	ActionWatcherList   = "watcherList"   // List watchers for a ticket

	// Filter actions
	ActionFilterSave   = "filterSave"   // Create or update a saved filter
	ActionFilterLoad   = "filterLoad"   // Load a saved filter
	ActionFilterList   = "filterList"   // List user's filters
	ActionFilterShare  = "filterShare"  // Share a filter
	ActionFilterModify = "filterModify" // Modify a filter
	ActionFilterRemove = "filterRemove" // Delete a filter

	// Custom field actions
	ActionCustomFieldCreate = "customFieldCreate"
	ActionCustomFieldRead   = "customFieldRead"
	ActionCustomFieldList   = "customFieldList"
	ActionCustomFieldModify = "customFieldModify"
	ActionCustomFieldRemove = "customFieldRemove"

	// Custom field option actions
	ActionCustomFieldOptionCreate = "customFieldOptionCreate"
	ActionCustomFieldOptionModify = "customFieldOptionModify"
	ActionCustomFieldOptionRemove = "customFieldOptionRemove"
	ActionCustomFieldOptionList   = "customFieldOptionList"

	// Custom field value actions (for tickets)
	ActionCustomFieldValueSet    = "customFieldValueSet"    // Set custom field value for a ticket
	ActionCustomFieldValueGet    = "customFieldValueGet"    // Get custom field value for a ticket
	ActionCustomFieldValueList   = "customFieldValueList"   // List all custom field values for a ticket
	ActionCustomFieldValueRemove = "customFieldValueRemove" // Remove custom field value from a ticket

	// Workflow actions
	ActionWorkflowCreate = "workflowCreate"
	ActionWorkflowRead   = "workflowRead"
	ActionWorkflowList   = "workflowList"
	ActionWorkflowModify = "workflowModify"
	ActionWorkflowRemove = "workflowRemove"

	// Workflow step actions
	ActionWorkflowStepCreate = "workflowStepCreate"
	ActionWorkflowStepRead   = "workflowStepRead"
	ActionWorkflowStepList   = "workflowStepList"
	ActionWorkflowStepModify = "workflowStepModify"
	ActionWorkflowStepRemove = "workflowStepRemove"

	// Ticket status actions
	ActionTicketStatusCreate = "ticketStatusCreate"
	ActionTicketStatusRead   = "ticketStatusRead"
	ActionTicketStatusList   = "ticketStatusList"
	ActionTicketStatusModify = "ticketStatusModify"
	ActionTicketStatusRemove = "ticketStatusRemove"

	// Ticket type actions
	ActionTicketTypeCreate        = "ticketTypeCreate"
	ActionTicketTypeRead          = "ticketTypeRead"
	ActionTicketTypeList          = "ticketTypeList"
	ActionTicketTypeModify        = "ticketTypeModify"
	ActionTicketTypeRemove        = "ticketTypeRemove"
	ActionTicketTypeAssign        = "ticketTypeAssign"        // Assign type to project
	ActionTicketTypeUnassign      = "ticketTypeUnassign"      // Unassign type from project
	ActionTicketTypeListByProject = "ticketTypeListByProject" // List types assigned to a project

	// Board actions
	ActionBoardCreate = "boardCreate"
	ActionBoardRead   = "boardRead"
	ActionBoardList   = "boardList"
	ActionBoardModify = "boardModify"
	ActionBoardRemove = "boardRemove"

	// Board ticket assignment
	ActionBoardAddTicket    = "boardAddTicket"
	ActionBoardRemoveTicket = "boardRemoveTicket"
	ActionBoardListTickets  = "boardListTickets"

	// Board metadata
	ActionBoardSetMetadata    = "boardSetMetadata"
	ActionBoardGetMetadata    = "boardGetMetadata"
	ActionBoardListMetadata   = "boardListMetadata"
	ActionBoardRemoveMetadata = "boardRemoveMetadata"

	// Cycle actions (Sprint/Milestone/Release management)
	ActionCycleCreate = "cycleCreate"
	ActionCycleRead   = "cycleRead"
	ActionCycleList   = "cycleList"
	ActionCycleModify = "cycleModify"
	ActionCycleRemove = "cycleRemove"

	// Cycle-project mapping
	ActionCycleAssignProject   = "cycleAssignProject"   // Assign cycle to project
	ActionCycleUnassignProject = "cycleUnassignProject" // Unassign cycle from project
	ActionCycleListProjects    = "cycleListProjects"    // List projects assigned to cycle

	// Cycle-ticket mapping
	ActionCycleAddTicket    = "cycleAddTicket"    // Add ticket to cycle
	ActionCycleRemoveTicket = "cycleRemoveTicket" // Remove ticket from cycle
	ActionCycleListTickets  = "cycleListTickets"  // List tickets in cycle

	// Account actions (Multi-tenancy support)
	ActionAccountCreate = "accountCreate"
	ActionAccountRead   = "accountRead"
	ActionAccountList   = "accountList"
	ActionAccountModify = "accountModify"
	ActionAccountRemove = "accountRemove"

	// Organization actions
	ActionOrganizationCreate        = "organizationCreate"
	ActionOrganizationRead          = "organizationRead"
	ActionOrganizationList          = "organizationList"
	ActionOrganizationModify        = "organizationModify"
	ActionOrganizationRemove        = "organizationRemove"
	ActionOrganizationAssignAccount = "organizationAssignAccount" // Assign organization to account
	ActionOrganizationListAccounts  = "organizationListAccounts"  // List accounts for organization

	// Team actions
	ActionTeamCreate              = "teamCreate"
	ActionTeamRead                = "teamRead"
	ActionTeamList                = "teamList"
	ActionTeamModify              = "teamModify"
	ActionTeamRemove              = "teamRemove"
	ActionTeamAssignOrganization  = "teamAssignOrganization"  // Assign team to organization
	ActionTeamUnassignOrganization = "teamUnassignOrganization" // Unassign team from organization
	ActionTeamListOrganizations   = "teamListOrganizations"   // List organizations for team
	ActionTeamAssignProject       = "teamAssignProject"       // Assign team to project
	ActionTeamUnassignProject     = "teamUnassignProject"     // Unassign team from project
	ActionTeamListProjects        = "teamListProjects"        // List projects for team

	// User-Organization mapping
	ActionUserAssignOrganization = "userAssignOrganization" // Assign user to organization
	ActionUserListOrganizations  = "userListOrganizations"  // List organizations for user
	ActionOrganizationListUsers  = "organizationListUsers"  // List users in organization

	// User-Team mapping
	ActionUserAssignTeam = "userAssignTeam" // Assign user to team
	ActionUserListTeams  = "userListTeams"  // List teams for user
	ActionTeamListUsers  = "teamListUsers"  // List users in team

	// Component actions
	ActionComponentCreate = "componentCreate"
	ActionComponentRead   = "componentRead"
	ActionComponentList   = "componentList"
	ActionComponentModify = "componentModify"
	ActionComponentRemove = "componentRemove"

	// Component-ticket mapping
	ActionComponentAddTicket    = "componentAddTicket"    // Add component to ticket
	ActionComponentRemoveTicket = "componentRemoveTicket" // Remove component from ticket
	ActionComponentListTickets  = "componentListTickets"  // List tickets for component

	// Component metadata
	ActionComponentSetMetadata    = "componentSetMetadata"    // Set component metadata
	ActionComponentGetMetadata    = "componentGetMetadata"    // Get component metadata
	ActionComponentListMetadata   = "componentListMetadata"   // List all metadata for component
	ActionComponentRemoveMetadata = "componentRemoveMetadata" // Remove component metadata

	// Label actions
	ActionLabelCreate = "labelCreate"
	ActionLabelRead   = "labelRead"
	ActionLabelList   = "labelList"
	ActionLabelModify = "labelModify"
	ActionLabelRemove = "labelRemove"

	// Label category actions
	ActionLabelCategoryCreate = "labelCategoryCreate"
	ActionLabelCategoryRead   = "labelCategoryRead"
	ActionLabelCategoryList   = "labelCategoryList"
	ActionLabelCategoryModify = "labelCategoryModify"
	ActionLabelCategoryRemove = "labelCategoryRemove"

	// Label-ticket mapping
	ActionLabelAddTicket    = "labelAddTicket"    // Add label to ticket
	ActionLabelRemoveTicket = "labelRemoveTicket" // Remove label from ticket
	ActionLabelListTickets  = "labelListTickets"  // List tickets for label

	// Label-category mapping
	ActionLabelAssignCategory   = "labelAssignCategory"   // Assign label to category
	ActionLabelUnassignCategory = "labelUnassignCategory" // Unassign label from category
	ActionLabelListCategories   = "labelListCategories"   // List categories for label

	// Asset actions
	ActionAssetCreate = "assetCreate"
	ActionAssetRead   = "assetRead"
	ActionAssetList   = "assetList"
	ActionAssetModify = "assetModify"
	ActionAssetRemove = "assetRemove"

	// Asset-ticket mapping
	ActionAssetAddTicket    = "assetAddTicket"    // Add asset to ticket
	ActionAssetRemoveTicket = "assetRemoveTicket" // Remove asset from ticket
	ActionAssetListTickets  = "assetListTickets"  // List tickets for asset

	// Asset-comment mapping
	ActionAssetAddComment    = "assetAddComment"    // Add asset to comment
	ActionAssetRemoveComment = "assetRemoveComment" // Remove asset from comment
	ActionAssetListComments  = "assetListComments"  // List comments for asset

	// Asset-project mapping
	ActionAssetAddProject    = "assetAddProject"    // Add asset to project
	ActionAssetRemoveProject = "assetRemoveProject" // Remove asset from project
	ActionAssetListProjects  = "assetListProjects"  // List projects for asset

	// Permission actions
	ActionPermissionCreate = "permissionCreate"
	ActionPermissionRead   = "permissionRead"
	ActionPermissionList   = "permissionList"
	ActionPermissionModify = "permissionModify"
	ActionPermissionRemove = "permissionRemove"

	// Permission context actions
	ActionPermissionContextCreate = "permissionContextCreate"
	ActionPermissionContextRead   = "permissionContextRead"
	ActionPermissionContextList   = "permissionContextList"
	ActionPermissionContextModify = "permissionContextModify"
	ActionPermissionContextRemove = "permissionContextRemove"

	// Permission-user mapping actions
	ActionPermissionAssignUser   = "permissionAssignUser"   // Assign permission to user
	ActionPermissionUnassignUser = "permissionUnassignUser" // Unassign permission from user

	// Permission-team mapping actions
	ActionPermissionAssignTeam   = "permissionAssignTeam"   // Assign permission to team
	ActionPermissionUnassignTeam = "permissionUnassignTeam" // Unassign permission from team

	// Permission query action
	ActionPermissionCheck = "permissionCheck" // Check if user has permission

	// Audit actions
	ActionAuditCreate  = "auditCreate"  // Create audit entry
	ActionAuditRead    = "auditRead"    // Read audit entry
	ActionAuditList    = "auditList"    // List audit entries
	ActionAuditQuery   = "auditQuery"   // Query audit entries with filters
	ActionAuditAddMeta = "auditAddMeta" // Add metadata to audit entry

	// Report actions
	ActionReportCreate  = "reportCreate"  // Create report
	ActionReportRead    = "reportRead"    // Read report
	ActionReportList    = "reportList"    // List reports
	ActionReportModify  = "reportModify"  // Modify report
	ActionReportRemove  = "reportRemove"  // Remove report
	ActionReportExecute = "reportExecute" // Execute/run report

	// Report metadata actions
	ActionReportSetMetadata    = "reportSetMetadata"    // Set report metadata
	ActionReportGetMetadata    = "reportGetMetadata"    // Get report metadata
	ActionReportRemoveMetadata = "reportRemoveMetadata" // Remove report metadata

	// Extension actions
	ActionExtensionCreate  = "extensionCreate"  // Create/register extension
	ActionExtensionRead    = "extensionRead"    // Read extension
	ActionExtensionList    = "extensionList"    // List extensions
	ActionExtensionModify  = "extensionModify"  // Modify extension
	ActionExtensionRemove  = "extensionRemove"  // Remove extension
	ActionExtensionEnable  = "extensionEnable"  // Enable extension
	ActionExtensionDisable = "extensionDisable" // Disable extension

	// Extension metadata actions
	ActionExtensionSetMetadata = "extensionSetMetadata" // Set extension metadata

	// Repository actions
	ActionRepositoryCreate = "repositoryCreate"
	ActionRepositoryRead   = "repositoryRead"
	ActionRepositoryList   = "repositoryList"
	ActionRepositoryModify = "repositoryModify"
	ActionRepositoryRemove = "repositoryRemove"

	// Repository type actions
	ActionRepositoryTypeCreate = "repositoryTypeCreate"
	ActionRepositoryTypeRead   = "repositoryTypeRead"
	ActionRepositoryTypeList   = "repositoryTypeList"
	ActionRepositoryTypeModify = "repositoryTypeModify"
	ActionRepositoryTypeRemove = "repositoryTypeRemove"

	// Repository-project mapping
	ActionRepositoryAssignProject   = "repositoryAssignProject"   // Assign repository to project
	ActionRepositoryUnassignProject = "repositoryUnassignProject" // Unassign repository from project
	ActionRepositoryListProjects    = "repositoryListProjects"    // List projects for repository

	// Repository-commit-ticket mapping
	ActionRepositoryAddCommit    = "repositoryAddCommit"    // Add commit to ticket
	ActionRepositoryRemoveCommit = "repositoryRemoveCommit" // Remove commit from ticket
	ActionRepositoryListCommits  = "repositoryListCommits"  // List commits for ticket
	ActionRepositoryGetCommit    = "repositoryGetCommit"    // Get commit details

	// Ticket relationship type actions
	ActionTicketRelationshipTypeCreate = "ticketRelationshipTypeCreate"
	ActionTicketRelationshipTypeRead   = "ticketRelationshipTypeRead"
	ActionTicketRelationshipTypeList   = "ticketRelationshipTypeList"
	ActionTicketRelationshipTypeModify = "ticketRelationshipTypeModify"
	ActionTicketRelationshipTypeRemove = "ticketRelationshipTypeRemove"

	// Ticket relationship actions
	ActionTicketRelationshipCreate = "ticketRelationshipCreate" // Create relationship between tickets
	ActionTicketRelationshipRemove = "ticketRelationshipRemove" // Remove relationship between tickets
	ActionTicketRelationshipList   = "ticketRelationshipList"   // List relationships for a ticket

	// =======================================================================
	// PHASE 2: AGILE ENHANCEMENTS
	// =======================================================================

	// Epic actions
	ActionEpicCreate      = "epicCreate"      // Create epic ticket
	ActionEpicRead        = "epicRead"        // Read epic
	ActionEpicList        = "epicList"        // List all epics
	ActionEpicModify      = "epicModify"      // Update epic
	ActionEpicRemove      = "epicRemove"      // Delete epic
	ActionEpicAddStory    = "epicAddStory"    // Add story to epic
	ActionEpicRemoveStory = "epicRemoveStory" // Remove story from epic
	ActionEpicListStories = "epicListStories" // List stories in epic

	// Subtask actions
	ActionSubtaskCreate        = "subtaskCreate"        // Create subtask
	ActionSubtaskList          = "subtaskList"          // List subtasks
	ActionSubtaskMoveToParent  = "subtaskMoveToParent"  // Change parent
	ActionSubtaskConvertToIssue = "subtaskConvertToIssue" // Convert to regular issue
	ActionSubtaskListByParent  = "subtaskListByParent"  // List all subtasks of parent

	// Work log actions
	ActionWorkLogAdd         = "workLogAdd"         // Add work log
	ActionWorkLogModify      = "workLogModify"      // Update work log
	ActionWorkLogRemove      = "workLogRemove"      // Delete work log
	ActionWorkLogList        = "workLogList"        // List work logs
	ActionWorkLogListByTicket = "workLogListByTicket" // List work logs for ticket
	ActionWorkLogListByUser  = "workLogListByUser"  // List work logs by user
	ActionWorkLogGetTotalTime = "workLogGetTotalTime" // Get total time spent

	// Project role actions
	ActionProjectRoleCreate      = "projectRoleCreate"      // Create project role
	ActionProjectRoleRead        = "projectRoleRead"        // Read project role
	ActionProjectRoleList        = "projectRoleList"        // List project roles
	ActionProjectRoleModify      = "projectRoleModify"      // Update project role
	ActionProjectRoleRemove      = "projectRoleRemove"      // Delete project role
	ActionProjectRoleAssignUser  = "projectRoleAssignUser"  // Assign user to role
	ActionProjectRoleUnassignUser = "projectRoleUnassignUser" // Remove user from role
	ActionProjectRoleListUsers   = "projectRoleListUsers"   // List users in role

	// Security level actions
	ActionSecurityLevelCreate = "securityLevelCreate" // Create security level
	ActionSecurityLevelRead   = "securityLevelRead"   // Read security level
	ActionSecurityLevelList   = "securityLevelList"   // List security levels
	ActionSecurityLevelModify = "securityLevelModify" // Update security level
	ActionSecurityLevelRemove = "securityLevelRemove" // Delete security level
	ActionSecurityLevelGrant  = "securityLevelGrant"  // Grant access to user/team/role
	ActionSecurityLevelRevoke = "securityLevelRevoke" // Revoke access
	ActionSecurityLevelCheck  = "securityLevelCheck"  // Check if user has access

	// Dashboard actions
	ActionDashboardCreate       = "dashboardCreate"       // Create dashboard
	ActionDashboardRead         = "dashboardRead"         // Read dashboard
	ActionDashboardList         = "dashboardList"         // List dashboards
	ActionDashboardModify       = "dashboardModify"       // Update dashboard
	ActionDashboardRemove       = "dashboardRemove"       // Delete dashboard
	ActionDashboardShare        = "dashboardShare"        // Share dashboard
	ActionDashboardUnshare      = "dashboardUnshare"      // Unshare dashboard
	ActionDashboardAddWidget    = "dashboardAddWidget"    // Add widget
	ActionDashboardRemoveWidget = "dashboardRemoveWidget" // Remove widget
	ActionDashboardModifyWidget = "dashboardModifyWidget" // Update widget
	ActionDashboardListWidgets  = "dashboardListWidgets"  // List widgets
	ActionDashboardSetLayout    = "dashboardSetLayout"    // Update layout

	// Advanced board configuration actions
	ActionBoardConfigureColumns = "boardConfigureColumns" // Configure columns
	ActionBoardAddColumn        = "boardAddColumn"        // Add column
	ActionBoardRemoveColumn     = "boardRemoveColumn"     // Remove column
	ActionBoardModifyColumn     = "boardModifyColumn"     // Update column
	ActionBoardListColumns      = "boardListColumns"      // List columns
	ActionBoardAddSwimlane      = "boardAddSwimlane"      // Add swimlane
	ActionBoardRemoveSwimlane   = "boardRemoveSwimlane"   // Remove swimlane
	ActionBoardListSwimlanes    = "boardListSwimlanes"    // List swimlanes
	ActionBoardAddQuickFilter   = "boardAddQuickFilter"   // Add quick filter
	ActionBoardRemoveQuickFilter = "boardRemoveQuickFilter" // Remove quick filter
	ActionBoardListQuickFilters = "boardListQuickFilters" // List quick filters
	ActionBoardSetType          = "boardSetType"          // Set board type (scrum/kanban)

	// =======================================================================
	// PHASE 3: COLLABORATION FEATURES
	// =======================================================================

	// Vote actions
	ActionVoteAdd    = "voteAdd"    // Add vote
	ActionVoteRemove = "voteRemove" // Remove vote
	ActionVoteCount  = "voteCount"  // Get vote count
	ActionVoteList   = "voteList"   // List voters
	ActionVoteCheck  = "voteCheck"  // Check if user voted

	// Project category actions
	ActionProjectCategoryCreate = "projectCategoryCreate" // Create category
	ActionProjectCategoryRead   = "projectCategoryRead"   // Read category
	ActionProjectCategoryList   = "projectCategoryList"   // List categories
	ActionProjectCategoryModify = "projectCategoryModify" // Update category
	ActionProjectCategoryRemove = "projectCategoryRemove" // Delete category
	ActionProjectCategoryAssign = "projectCategoryAssign" // Assign to project

	// Notification scheme actions
	ActionNotificationSchemeCreate     = "notificationSchemeCreate"     // Create scheme
	ActionNotificationSchemeRead       = "notificationSchemeRead"       // Read scheme
	ActionNotificationSchemeList       = "notificationSchemeList"       // List schemes
	ActionNotificationSchemeModify     = "notificationSchemeModify"     // Update scheme
	ActionNotificationSchemeRemove     = "notificationSchemeRemove"     // Delete scheme
	ActionNotificationSchemeAddRule    = "notificationSchemeAddRule"    // Add rule
	ActionNotificationSchemeRemoveRule = "notificationSchemeRemoveRule" // Remove rule
	ActionNotificationSchemeListRules  = "notificationSchemeListRules"  // List rules
	ActionNotificationEventList        = "notificationEventList"        // List event types
	ActionNotificationSend             = "notificationSend"             // Send notification (manual trigger)

	// Activity stream actions (enhancements to audit)
	ActionActivityStreamGet          = "activityStreamGet"          // Get activity stream
	ActionActivityStreamGetByProject = "activityStreamGetByProject" // Get project activity
	ActionActivityStreamGetByUser    = "activityStreamGetByUser"    // Get user activity
	ActionActivityStreamGetByTicket  = "activityStreamGetByTicket"  // Get ticket activity
	ActionActivityStreamFilter       = "activityStreamFilter"       // Filter by activity type

	// Comment mention actions
	ActionCommentMention       = "commentMention"       // Add mention to comment
	ActionCommentUnmention     = "commentUnmention"     // Remove mention
	ActionCommentListMentions  = "commentListMentions"  // List mentioned users
	ActionCommentGetMentions   = "commentGetMentions"   // Get mentions for user
	ActionCommentParseMentions = "commentParseMentions" // Parse @mentions from text
)
