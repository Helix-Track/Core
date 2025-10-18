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

	// =======================================================================
	// CHAT EXTENSION: REAL-TIME MESSAGING
	// =======================================================================

	// User presence actions
	ActionPresenceUpdate      = "presenceUpdate"      // Update user presence status
	ActionPresenceGet         = "presenceGet"         // Get user presence
	ActionPresenceList        = "presenceList"        // List presence for multiple users
	ActionPresenceGetByStatus = "presenceGetByStatus" // Get users by presence status

	// Chat room management actions
	ActionChatRoomCreate        = "chatRoomCreate"        // Create chat room
	ActionChatRoomRead          = "chatRoomRead"          // Get chat room details
	ActionChatRoomList          = "chatRoomList"          // List accessible chat rooms
	ActionChatRoomModify        = "chatRoomModify"        // Modify chat room
	ActionChatRoomRemove        = "chatRoomRemove"        // Delete chat room
	ActionChatRoomArchive       = "chatRoomArchive"       // Archive chat room
	ActionChatRoomUnarchive     = "chatRoomUnarchive"     // Unarchive chat room
	ActionChatRoomGetByEntity   = "chatRoomGetByEntity"   // Get chat room for entity (project, ticket, etc.)
	ActionChatRoomListByType    = "chatRoomListByType"    // List rooms by type
	ActionChatRoomSearch        = "chatRoomSearch"        // Search chat rooms

	// Chat participant actions
	ActionChatParticipantAdd       = "chatParticipantAdd"       // Add participant to room
	ActionChatParticipantRemove    = "chatParticipantRemove"    // Remove participant from room
	ActionChatParticipantList      = "chatParticipantList"      // List participants in room
	ActionChatParticipantSetRole   = "chatParticipantSetRole"   // Change participant role
	ActionChatParticipantMute      = "chatParticipantMute"      // Mute chat room for user
	ActionChatParticipantUnmute    = "chatParticipantUnmute"    // Unmute chat room for user
	ActionChatParticipantLeave     = "chatParticipantLeave"     // Leave chat room
	ActionChatParticipantGetRooms  = "chatParticipantGetRooms"  // Get all rooms for user

	// Message management actions
	ActionMessageCreate       = "messageCreate"       // Send message
	ActionMessageRead         = "messageRead"         // Get message details
	ActionMessageList         = "messageList"         // List messages in room
	ActionMessageModify       = "messageModify"       // Edit message
	ActionMessageRemove       = "messageRemove"       // Delete message
	ActionMessagePin          = "messagePin"          // Pin message
	ActionMessageUnpin        = "messageUnpin"        // Unpin message
	ActionMessageGetPinned    = "messageGetPinned"    // Get pinned messages
	ActionMessageReply        = "messageReply"        // Reply to message (threading)
	ActionMessageQuote        = "messageQuote"        // Quote message
	ActionMessageGetThread    = "messageGetThread"    // Get message thread (replies)
	ActionMessageSearch       = "messageSearch"       // Full-text search messages
	ActionMessageGetRecent    = "messageGetRecent"    // Get recent messages
	ActionMessageMarkAsRead   = "messageMarkAsRead"   // Mark message as read
	ActionMessageGetUnread    = "messageGetUnread"    // Get unread messages count

	// Typing indicator actions
	ActionTypingStart  = "typingStart"  // User starts typing
	ActionTypingStop   = "typingStop"   // User stops typing
	ActionTypingGetAll = "typingGetAll" // Get all typing users in room

	// Read receipt actions
	ActionReadReceiptCreate = "readReceiptCreate" // Create read receipt
	ActionReadReceiptList   = "readReceiptList"   // List read receipts for message
	ActionReadReceiptGet    = "readReceiptGet"    // Get read status for user

	// Message attachment actions
	ActionAttachmentUpload = "attachmentUpload" // Upload file attachment
	ActionAttachmentList   = "attachmentList"   // List attachments for message
	ActionAttachmentRemove = "attachmentRemove" // Remove attachment
	ActionAttachmentGet    = "attachmentGet"    // Get attachment details

	// Message reaction actions
	ActionReactionAdd    = "reactionAdd"    // Add emoji reaction
	ActionReactionRemove = "reactionRemove" // Remove emoji reaction
	ActionReactionList   = "reactionList"   // List reactions for message
	ActionReactionGet    = "reactionGet"    // Get reactions grouped by emoji

	// External chat integration actions
	ActionChatIntegrationCreate = "chatIntegrationCreate" // Create external integration
	ActionChatIntegrationRead   = "chatIntegrationRead"   // Get integration details
	ActionChatIntegrationList   = "chatIntegrationList"   // List integrations for room
	ActionChatIntegrationModify = "chatIntegrationModify" // Modify integration
	ActionChatIntegrationRemove = "chatIntegrationRemove" // Remove integration
	ActionChatIntegrationSync   = "chatIntegrationSync"   // Sync with external provider

	// ========================================================================
	// DOCUMENTS EXTENSION V2 ACTIONS (90 actions)
	// Requires: Documents Extension V2 + Core V5
	// ========================================================================

	// Core document actions (22)
	ActionDocumentCreate          = "documentCreate"          // Create new document
	ActionDocumentRead            = "documentRead"            // Get document by ID
	ActionDocumentList            = "documentList"            // List documents (filtered)
	ActionDocumentModify          = "documentModify"          // Update document
	ActionDocumentUpdate          = "documentModify"          // Alias for Modify (backwards compatibility)
	ActionDocumentRemove          = "documentRemove"          // Delete document (soft)
	ActionDocumentDelete          = "documentRemove"          // Alias for Remove (backwards compatibility)
	ActionDocumentRestore         = "documentRestore"         // Restore deleted document
	ActionDocumentArchive         = "documentArchive"         // Archive document
	ActionDocumentUnarchive       = "documentUnarchive"       // Unarchive document
	ActionDocumentDuplicate       = "documentDuplicate"       // Duplicate document
	ActionDocumentMove            = "documentMove"            // Move to different space
	ActionDocumentGetHierarchy    = "documentGetHierarchy"    // Get document tree
	ActionDocumentSearch          = "documentSearch"          // Full-text search
	ActionDocumentGetRelated      = "documentGetRelated"      // Get related documents
	ActionDocumentSetParent       = "documentSetParent"       // Set parent document
	ActionDocumentGetChildren     = "documentGetChildren"     // Get child documents
	ActionDocumentGetBreadcrumb   = "documentGetBreadcrumb"   // Get breadcrumb trail
	ActionDocumentGenerateTOC     = "documentGenerateTOC"     // Generate table of contents
	ActionDocumentGetMetadata     = "documentGetMetadata"     // Get document metadata
	ActionDocumentPublish            = "documentPublish"            // Publish document
	ActionDocumentUnpublish          = "documentUnpublish"          // Unpublish document
	ActionDocumentContentGet         = "documentContentGet"         // Get document content
	ActionDocumentContentUpdate      = "documentContentUpdate"      // Update document content
	ActionDocumentContentGetVersion  = "documentContentGetVersion"  // Get specific version content
	ActionDocumentContentGetLatest   = "documentContentGetLatest"   // Get latest version content

	// Document versioning actions (16)
	ActionDocumentVersionCreate     = "documentVersionCreate"     // Create new version
	ActionDocumentVersionList       = "documentVersionList"       // List all versions
	ActionDocumentVersionGet        = "documentVersionGet"        // Get specific version
	ActionDocumentVersionCompare    = "documentVersionCompare"    // Compare two versions
	ActionDocumentVersionRestore    = "documentVersionRestore"    // Rollback to version
	ActionDocumentVersionLabel         = "documentVersionLabel"         // Add label to version
	ActionDocumentVersionLabelCreate   = "documentVersionLabel"         // Alias for Label (backwards compatibility)
	ActionDocumentVersionLabelList     = "documentVersionLabelList"     // List labels for version
	ActionDocumentVersionComment       = "documentVersionComment"       // Add comment to version
	ActionDocumentVersionCommentCreate = "documentVersionComment"       // Alias for Comment (backwards compatibility)
	ActionDocumentVersionCommentList   = "documentVersionCommentList"   // List comments for version
	ActionDocumentVersionTag           = "documentVersionTag"           // Tag a version
	ActionDocumentVersionTagCreate     = "documentVersionTag"           // Alias for Tag (backwards compatibility)
	ActionDocumentVersionTagList       = "documentVersionTagList"       // List tags for version
	ActionDocumentVersionMention       = "documentVersionMention"       // Mention users in version
	ActionDocumentVersionMentionCreate = "documentVersionMention"       // Alias for Mention (backwards compatibility)
	ActionDocumentVersionMentionList   = "documentVersionMentionList"   // List mentions for version
	ActionDocumentVersionGetDiff       = "documentVersionGetDiff"       // Get diff between versions
	ActionDocumentVersionDiffGet       = "documentVersionGetDiff"       // Alias for GetDiff (backwards compatibility)
	ActionDocumentVersionDiffCreate    = "documentVersionDiffCreate"    // Create/store version diff
	ActionDocumentVersionGetHistory  = "documentVersionGetHistory"  // Get full version history
	ActionDocumentVersionSetMajor   = "documentVersionSetMajor"   // Mark as major version
	ActionDocumentVersionSetMinor   = "documentVersionSetMinor"   // Mark as minor version
	ActionDocumentVersionGetLabels  = "documentVersionGetLabels"  // Get version labels
	ActionDocumentVersionGetComments = "documentVersionGetComments" // Get version comments
	ActionDocumentVersionGetTags    = "documentVersionGetTags"    // Get version tags

	// Document collaboration actions (12) - uses core entities
	ActionDocumentCommentAdd           = "documentCommentAdd"           // Add comment (uses core comment)
	ActionDocumentCommentReply         = "documentCommentReply"         // Reply to comment
	ActionDocumentCommentEdit          = "documentCommentEdit"          // Edit comment
	ActionDocumentCommentRemove        = "documentCommentRemove"        // Delete comment
	ActionDocumentCommentList            = "documentCommentList"            // List all comments
	ActionDocumentInlineCommentAdd       = "documentInlineCommentAdd"       // Add inline comment
	ActionDocumentInlineCommentCreate    = "documentInlineCommentAdd"       // Alias for Add (backwards compatibility)
	ActionDocumentInlineCommentResolve   = "documentInlineCommentResolve"   // Resolve inline comment
	ActionDocumentInlineCommentList      = "documentInlineCommentList"      // List inline comments
	ActionDocumentMention                = "documentMention"                // Mention user in document
	ActionDocumentReact                  = "documentReact"                  // Add reaction/like (uses core vote)
	ActionDocumentVoteAdd                = "documentReact"                  // Alias for React (backwards compatibility)
	ActionDocumentVoteRemove             = "documentVoteRemove"             // Remove vote/reaction
	ActionDocumentVoteCount              = "documentVoteCount"              // Get vote count
	ActionDocumentGetReactions           = "documentGetReactions"           // Get all reactions
	ActionDocumentWatch                  = "documentWatch"                  // Start watching document
	ActionDocumentWatcherAdd             = "documentWatch"                  // Alias for Watch (backwards compatibility)
	ActionDocumentUnwatch                = "documentUnwatch"                // Stop watching document
	ActionDocumentWatcherRemove          = "documentUnwatch"                // Alias for Unwatch (backwards compatibility)
	ActionDocumentWatcherList            = "documentWatcherList"            // List watchers

	// Document organization actions (10) - tags + core labels
	ActionDocumentLabelAdd    = "documentLabelAdd"    // Add label (uses core label)
	ActionDocumentLabelRemove = "documentLabelRemove" // Remove label
	ActionDocumentLabelList   = "documentLabelList"   // List document labels
	ActionDocumentTagAdd                = "documentTagAdd"      // Add tag to document
	ActionDocumentTagCreate             = "documentTagCreate"   // Create new tag
	ActionDocumentTagAddToDocument      = "documentTagAdd"      // Alias for Add (backwards compatibility)
	ActionDocumentTagGet                = "documentTagGet"      // Get tag details
	ActionDocumentTagRemove             = "documentTagRemove"   // Remove tag
	ActionDocumentTagRemoveFromDocument = "documentTagRemove"   // Alias for Remove (backwards compatibility)
	ActionDocumentTagList               = "documentTagList"     // List document tags
	ActionDocumentCategoryAssign        = "documentCategoryAssign" // Assign document to category
	ActionDocumentCategoryList          = "documentCategoryList"   // List document categories
	ActionDocumentSpaceCreate = "documentSpaceCreate" // Create document space
	ActionDocumentSpaceRead   = "documentSpaceRead"   // Read/get space
	ActionDocumentSpaceList   = "documentSpaceList"   // List spaces
	ActionDocumentSpaceModify = "documentSpaceModify" // Modify space
	ActionDocumentSpaceUpdate = "documentSpaceModify" // Alias for Modify (backwards compatibility)
	ActionDocumentSpaceRemove = "documentSpaceRemove" // Remove space
	ActionDocumentSpaceDelete = "documentSpaceRemove" // Alias for Remove (backwards compatibility)

	// Document export actions (8)
	ActionDocumentExportPDF             = "documentExportPDF"             // Export to PDF
	ActionDocumentExportWord            = "documentExportWord"            // Export to Word (DOCX)
	ActionDocumentExportHTML            = "documentExportHTML"            // Export to HTML
	ActionDocumentExportXML             = "documentExportXML"             // Export to XML
	ActionDocumentExportMarkdown        = "documentExportMarkdown"        // Export to Markdown
	ActionDocumentExportPlainText       = "documentExportPlainText"       // Export to plain text
	ActionDocumentExportText            = "documentExportPlainText"       // Alias for Plain Text export
	ActionDocumentBulkExport            = "documentBulkExport"            // Bulk export documents
	ActionDocumentExportWithAttachments = "documentExportWithAttachments" // Export with attachments
	ActionDocumentExportStatus          = "documentExportStatus"          // Get export job status
	ActionDocumentExportDownload        = "documentExportDownload"        // Download exported file
	ActionDocumentExportCancel          = "documentExportCancel"          // Cancel export job
	ActionDocumentExportList            = "documentExportList"            // List export jobs

	// Document entity connection actions (10)
	ActionDocumentLinkToTicket       = "documentLinkToTicket"       // Link to ticket
	ActionDocumentLinkToProject      = "documentLinkToProject"      // Link to project
	ActionDocumentLinkToUser         = "documentLinkToUser"         // Link to user
	ActionDocumentLinkToLabel        = "documentLinkToLabel"        // Link to label
	ActionDocumentLinkToAny          = "documentLinkToAny"          // Link to any entity
	ActionDocumentEntityLinkCreate   = "documentLinkToAny"          // Alias for generic link creation
	ActionDocumentUnlink             = "documentUnlink"             // Remove link
	ActionDocumentEntityLinkDelete   = "documentUnlink"             // Alias for generic link deletion
	ActionDocumentEntityLinkRemove   = "documentUnlink"             // Another alias for removing link
	ActionDocumentGetLinks           = "documentGetLinks"           // Get all links
	ActionDocumentEntityLinkList       = "documentGetLinks"              // Alias for GetLinks
	ActionDocumentGetLinkedBy          = "documentGetLinkedBy"           // Get entities linking to doc
	ActionDocumentRelationshipCreate   = "documentRelationshipCreate"    // Create document relationship
	ActionDocumentRelationshipList     = "documentRelationshipList"      // List document relationships
	ActionDocumentRelationshipRemove   = "documentRelationshipRemove"    // Remove document relationship
	ActionDocumentEntityDocumentsList  = "documentEntityDocumentsList"   // List entity documents
	ActionDocumentProjectWikiGet       = "documentProjectWikiGet"        // Get project wiki

	// Document template & blueprint actions (11)
	ActionDocumentTemplateCreate     = "documentTemplateCreate"     // Create template
	ActionDocumentTemplateRead       = "documentTemplateGet"        // Read template (alias)
	ActionDocumentTemplateList       = "documentTemplateList"       // List templates
	ActionDocumentTemplateGet        = "documentTemplateGet"        // Get template
	ActionDocumentTemplateModify     = "documentTemplateModify"     // Modify template
	ActionDocumentTemplateUpdate     = "documentTemplateModify"     // Alias for Modify (backwards compatibility)
	ActionDocumentTemplateRemove     = "documentTemplateRemove"     // Remove template
	ActionDocumentTemplateDelete     = "documentTemplateRemove"     // Alias for Remove (backwards compatibility)
	ActionDocumentCreateFromTemplate = "documentCreateFromTemplate" // Create from template
	ActionDocumentBlueprintCreate    = "documentBlueprintCreate"    // Create blueprint
	ActionDocumentBlueprintList      = "documentBlueprintList"      // List blueprints
	ActionDocumentBlueprintGet       = "documentBlueprintGet"       // Get blueprint

	// Document analytics actions (9)
	ActionDocumentGetViews         = "documentGetViews"         // Get view count/history
	ActionDocumentGetPopular       = "documentGetPopular"       // Get popular documents
	ActionDocumentPopularGet       = "documentGetPopular"       // Alias for GetPopular
	ActionDocumentGetActivity      = "documentGetActivity"      // Get activity stream
	ActionDocumentTrackView        = "documentTrackView"        // Track document view
	ActionDocumentViewRecord       = "documentTrackView"        // Alias for TrackView
	ActionDocumentGetStatistics    = "documentGetStatistics"    // Get document statistics
	ActionDocumentAnalyticsGet     = "documentAnalyticsGet"     // Get analytics data
	ActionDocumentAnalyticsUpdate  = "documentAnalyticsUpdate"  // Update analytics
	ActionDocumentViewHistoryGet   = "documentViewHistoryGet"   // Get view history

	// Document attachment actions (7)
	ActionDocumentAttachmentAdd    = "documentAttachmentAdd"    // Add attachment
	ActionDocumentAttachmentUpload = "documentAttachmentAdd"    // Alias for Add (backwards compatibility)
	ActionDocumentAttachmentRemove = "documentAttachmentRemove" // Remove attachment
	ActionDocumentAttachmentDelete = "documentAttachmentRemove" // Alias for Remove (backwards compatibility)
	ActionDocumentAttachmentList   = "documentAttachmentList"   // List attachments
	ActionDocumentAttachmentGet    = "documentAttachmentGet"    // Get attachment
	ActionDocumentAttachmentUpdate = "documentAttachmentUpdate" // Update attachment
)
