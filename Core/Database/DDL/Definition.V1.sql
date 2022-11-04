/*
    Version: 1
*/

/*
    Notes:

    - The main project board: https://github.com/orgs/red-elf/projects/2/views/1
    - Identifiers in the system are UUID strings.
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
    - Additional tables are defined to provide the meta-data to entities of the system.
    - To follow the order of entities definition in the system follow the 'DROP TABLE' directives.
*/

DROP TABLE IF EXISTS system_info;
DROP TABLE IF EXISTS organization;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS team_organization_mapping;
DROP TABLE IF EXISTS team_project_mapping;
DROP TABLE IF EXISTS user_organization_mapping;
DROP TABLE IF EXISTS user_team_mapping;
DROP TABLE IF EXISTS user_default_mapping;
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS project_organization_mapping;
DROP TABLE IF EXISTS ticket;
DROP TABLE IF EXISTS ticket_meta_data;
DROP TABLE IF EXISTS ticket_project_mapping;
DROP TABLE IF EXISTS ticket_cycle_mapping;
DROP TABLE IF EXISTS ticket_board_mapping;
DROP TABLE IF EXISTS ticket_type;
DROP TABLE IF EXISTS ticket_status;
DROP TABLE IF EXISTS ticket_relationship_type;
DROP TABLE IF EXISTS ticket_relationship;
DROP TABLE IF EXISTS ticket_type_project_mapping;
DROP TABLE IF EXISTS board;
DROP TABLE IF EXISTS board_meta_data;
DROP TABLE IF EXISTS workflow;
DROP TABLE IF EXISTS workflow_step;
DROP TABLE IF EXISTS cycle;
DROP TABLE IF EXISTS cycle_project_mapping;
DROP TABLE IF EXISTS asset;
DROP TABLE IF EXISTS asset_ticket_mapping;
DROP TABLE IF EXISTS asset_comment_mapping;
DROP TABLE IF EXISTS asset_project_mapping;
DROP TABLE IF EXISTS asset_team_mapping;
DROP TABLE IF EXISTS label;
DROP TABLE IF EXISTS label_category;
DROP TABLE IF EXISTS label_label_category_mapping;
DROP TABLE IF EXISTS label_ticket_mapping;
DROP TABLE IF EXISTS label_asset_mapping;
DROP TABLE IF EXISTS label_team_mapping;
DROP TABLE IF EXISTS label_project_mapping;
DROP TABLE IF EXISTS comment;
DROP TABLE IF EXISTS comment_ticket_mapping;
DROP TABLE IF EXISTS repository;
DROP TABLE IF EXISTS repository_type;
DROP TABLE IF EXISTS repository_project_mapping;
DROP TABLE IF EXISTS repository_commit_ticket_mapping;
DROP TABLE IF EXISTS component;
DROP TABLE IF EXISTS component_meta_data;
DROP TABLE IF EXISTS component_ticket_mapping;
DROP TABLE IF EXISTS permission;
DROP TABLE IF EXISTS permission_user_mapping;
DROP TABLE IF EXISTS permission_team_mapping;
DROP TABLE IF EXISTS permission_context;
DROP TABLE IF EXISTS audit;
DROP TABLE IF EXISTS audit_meta_data;
DROP TABLE IF EXISTS report;
DROP TABLE IF EXISTS report_meta_data;
DROP TABLE IF EXISTS extension;
DROP TABLE IF EXISTS extension_meta_data;
DROP TABLE IF EXISTS configuration_data_extension_mapping;

DROP INDEX IF EXISTS system_info_get_by_created;
DROP INDEX IF EXISTS system_info_get_by_description;
DROP INDEX IF EXISTS system_info_get_by_created_and_description;
DROP INDEX IF EXISTS projects_get_by_title;
DROP INDEX IF EXISTS projects_get_by_description;
DROP INDEX IF EXISTS projects_get_by_title_and_description;
DROP INDEX IF EXISTS projects_get_by_created;
DROP INDEX IF EXISTS projects_get_by_modified;
DROP INDEX IF EXISTS projects_get_by_created_and_modified;
DROP INDEX IF EXISTS projects_get_by_deleted;
DROP INDEX IF EXISTS projects_get_by_identifier;
DROP INDEX IF EXISTS projects_get_by_workflow_id;
DROP INDEX IF EXISTS ticket_types_get_by_title;
DROP INDEX IF EXISTS ticket_types_get_by_description;
DROP INDEX IF EXISTS ticket_types_get_by_title_and_description;
DROP INDEX IF EXISTS ticket_types_get_by_created;
DROP INDEX IF EXISTS ticket_types_get_by_modified;
DROP INDEX IF EXISTS ticket_types_get_by_deleted;
DROP INDEX IF EXISTS ticket_types_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_statuses_get_by_title;
DROP INDEX IF EXISTS ticket_statuses_get_by_description;
DROP INDEX IF EXISTS ticket_statuses_get_by_title_and_description;
DROP INDEX IF EXISTS ticket_statuses_get_by_deleted;
DROP INDEX IF EXISTS ticket_statuses_get_by_created;
DROP INDEX IF EXISTS ticket_statuses_get_by_modified;
DROP INDEX IF EXISTS ticket_statuses_get_by_created_and_modified;
DROP INDEX IF EXISTS tickets_get_by_ticket_number;
DROP INDEX IF EXISTS tickets_get_by_ticket_type_id;
DROP INDEX IF EXISTS tickets_get_by_ticket_status_id;
DROP INDEX IF EXISTS tickets_get_by_project_id;
DROP INDEX IF EXISTS tickets_get_by_user_id;
DROP INDEX IF EXISTS tickets_get_by_creator;
DROP INDEX IF EXISTS tickets_get_by_project_id_and_user_id;
DROP INDEX IF EXISTS tickets_get_by_project_id_and_creator;
DROP INDEX IF EXISTS tickets_get_by_estimation;
DROP INDEX IF EXISTS tickets_get_by_story_points;
DROP INDEX IF EXISTS tickets_get_by_created;
DROP INDEX IF EXISTS tickets_get_by_modified;
DROP INDEX IF EXISTS tickets_get_by_deleted;
DROP INDEX IF EXISTS tickets_get_by_created_and_modified;
DROP INDEX IF EXISTS tickets_get_by_title;
DROP INDEX IF EXISTS tickets_get_by_description;
DROP INDEX IF EXISTS tickets_get_by_title_and_description;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_title;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_description;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_title_and_description;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_created;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_deleted;
DROP INDEX IF EXISTS ticket_relationship_types_get_by_created_and_modified;
DROP INDEX IF EXISTS boards_get_by_title;
DROP INDEX IF EXISTS boards_get_by_description;
DROP INDEX IF EXISTS boards_get_by_title_and_description;
DROP INDEX IF EXISTS boards_get_by_created;
DROP INDEX IF EXISTS boards_get_by_modified;
DROP INDEX IF EXISTS boards_get_by_deleted;
DROP INDEX IF EXISTS boards_get_by_created_and_modified;
DROP INDEX IF EXISTS workflows_get_by_title;
DROP INDEX IF EXISTS workflows_get_by_description;
DROP INDEX IF EXISTS workflows_get_by_title_and_description;
DROP INDEX IF EXISTS workflows_get_by_created;
DROP INDEX IF EXISTS workflows_get_by_modified;
DROP INDEX IF EXISTS workflows_get_by_deleted;
DROP INDEX IF EXISTS workflows_get_by_created_and_modified;
DROP INDEX IF EXISTS assets_get_by_url;
DROP INDEX IF EXISTS assets_get_by_description;
DROP INDEX IF EXISTS assets_get_by_created;
DROP INDEX IF EXISTS assets_get_by_deleted;
DROP INDEX IF EXISTS assets_get_by_modified;
DROP INDEX IF EXISTS assets_get_by_created_and_modified;
DROP INDEX IF EXISTS labels_get_by_title;
DROP INDEX IF EXISTS labels_get_by_description;
DROP INDEX IF EXISTS labels_get_by_title_and_description;
DROP INDEX IF EXISTS labels_get_by_created;
DROP INDEX IF EXISTS labels_get_by_deleted;
DROP INDEX IF EXISTS labels_get_by_modified;
DROP INDEX IF EXISTS labels_get_by_created_and_modified;
DROP INDEX IF EXISTS label_categories_get_by_title;
DROP INDEX IF EXISTS label_categories_get_by_description;
DROP INDEX IF EXISTS label_categories_get_by_title_and_description;
DROP INDEX IF EXISTS label_categories_get_by_created;
DROP INDEX IF EXISTS label_categories_get_by_deleted;
DROP INDEX IF EXISTS label_categories_get_by_modified;
DROP INDEX IF EXISTS label_categories_get_by_created_and_modified;
DROP INDEX IF EXISTS repositories_get_by_repository;
DROP INDEX IF EXISTS repositories_get_by_description;
DROP INDEX IF EXISTS repositories_get_by_repository_and_description;
DROP INDEX IF EXISTS repositories_get_by_deleted;
DROP INDEX IF EXISTS repositories_get_by_repository_type_id;
DROP INDEX IF EXISTS repositories_get_by_created;
DROP INDEX IF EXISTS repositories_get_by_modified;
DROP INDEX IF EXISTS repositories_get_by_created_and_modified;
DROP INDEX IF EXISTS repository_types_get_by_title;
DROP INDEX IF EXISTS repository_types_get_by_description;
DROP INDEX IF EXISTS repository_types_get_by_title_and_description;
DROP INDEX IF EXISTS repository_types_get_by_deleted;
DROP INDEX IF EXISTS repository_types_get_by_created;
DROP INDEX IF EXISTS repository_types_get_by_modified;
DROP INDEX IF EXISTS repository_types_get_by_created_and_modified;
DROP INDEX IF EXISTS components_get_by_title;
DROP INDEX IF EXISTS components_get_by_description;
DROP INDEX IF EXISTS components_get_by_title_description;
DROP INDEX IF EXISTS components_get_by_created;
DROP INDEX IF EXISTS components_get_by_deleted;
DROP INDEX IF EXISTS components_get_by_modified;
DROP INDEX IF EXISTS components_get_by_created_modified;
DROP INDEX IF EXISTS organizations_get_by_title;
DROP INDEX IF EXISTS organizations_get_by_description;
DROP INDEX IF EXISTS organizations_get_by_title_and_description;
DROP INDEX IF EXISTS organizations_get_by_created;
DROP INDEX IF EXISTS organizations_get_by_deleted;
DROP INDEX IF EXISTS organizations_get_by_modified;
DROP INDEX IF EXISTS organizations_get_by_created_and_modified;
DROP INDEX IF EXISTS teams_get_by_title;
DROP INDEX IF EXISTS teams_get_by_description;
DROP INDEX IF EXISTS teams_get_by_title_and_description;
DROP INDEX IF EXISTS teams_get_by_created;
DROP INDEX IF EXISTS teams_get_by_modified;
DROP INDEX IF EXISTS teams_get_by_deleted;
DROP INDEX IF EXISTS teams_get_by_created_and_modified;
DROP INDEX IF EXISTS permissions_get_by_title;
DROP INDEX IF EXISTS permissions_get_by_description;
DROP INDEX IF EXISTS permissions_get_by_title_and_description;
DROP INDEX IF EXISTS permissions_get_by_deleted;
DROP INDEX IF EXISTS permissions_get_by_created;
DROP INDEX IF EXISTS permissions_get_by_modified;
DROP INDEX IF EXISTS permissions_get_by_created_and_modified;
DROP INDEX IF EXISTS comments_get_by_comment;
DROP INDEX IF EXISTS comments_get_by_created;
DROP INDEX IF EXISTS comments_get_by_modified;
DROP INDEX IF EXISTS comments_get_by_deleted;
DROP INDEX IF EXISTS comments_get_by_created_and_modified;
DROP INDEX IF EXISTS permission_contexts_get_by_title;
DROP INDEX IF EXISTS permission_contexts_get_by_description;
DROP INDEX IF EXISTS permission_contexts_get_by_title_and_description;
DROP INDEX IF EXISTS permission_contexts_get_by_created;
DROP INDEX IF EXISTS permission_contexts_get_by_modified;
DROP INDEX IF EXISTS permission_contexts_get_by_deleted;
DROP INDEX IF EXISTS permission_contexts_get_by_created_and_modified;
DROP INDEX IF EXISTS workflow_steps_get_by_title;
DROP INDEX IF EXISTS workflow_steps_get_by_description;
DROP INDEX IF EXISTS workflow_steps_get_by_title_and_description;
DROP INDEX IF EXISTS workflow_steps_get_by_workflow_id;
DROP INDEX IF EXISTS workflow_steps_get_by_workflow_step_id;
DROP INDEX IF EXISTS workflow_steps_get_by_ticket_status_id;
DROP INDEX IF EXISTS workflow_steps_get_by_workflow_id_and_ticket_status_id;
DROP INDEX IF EXISTS workflow_steps_get_by_workflow_id_and_workflow_step_id;
DROP INDEX IF EXISTS workflow_steps_get_by_workflow_id_and_workflow_step_id_and_ticket_status_id;
DROP INDEX IF EXISTS workflow_steps_get_by_created;
DROP INDEX IF EXISTS workflow_steps_get_by_deleted;
DROP INDEX IF EXISTS workflow_steps_get_by_modified;
DROP INDEX IF EXISTS workflow_steps_get_by_created_and_modified;
DROP INDEX IF EXISTS reports_get_by_title;
DROP INDEX IF EXISTS reports_get_by_description;
DROP INDEX IF EXISTS reports_get_by_title_and_description;
DROP INDEX IF EXISTS reports_get_by_created;
DROP INDEX IF EXISTS reports_get_by_deleted;
DROP INDEX IF EXISTS reports_get_by_modified;
DROP INDEX IF EXISTS reports_get_by_created_and_modified;
DROP INDEX IF EXISTS cycles_get_by_title;
DROP INDEX IF EXISTS cycles_get_by_description;
DROP INDEX IF EXISTS cycles_get_by_title_and_description;
DROP INDEX IF EXISTS cycles_get_by_cycle_id;
DROP INDEX IF EXISTS cycles_get_by_type;
DROP INDEX IF EXISTS cycles_get_by_cycle_id_and_type;
DROP INDEX IF EXISTS cycles_get_by_created;
DROP INDEX IF EXISTS cycles_get_by_deleted;
DROP INDEX IF EXISTS cycles_get_by_modified;
DROP INDEX IF EXISTS cycles_get_by_created_and_modified;
DROP INDEX IF EXISTS extensions_get_by_title;
DROP INDEX IF EXISTS extensions_get_by_description;
DROP INDEX IF EXISTS extensions_get_by_title_and_description;
DROP INDEX IF EXISTS extensions_get_by_extension_key;
DROP INDEX IF EXISTS extensions_get_by_created;
DROP INDEX IF EXISTS extensions_get_by_deleted;
DROP INDEX IF EXISTS extensions_get_by_enabled;
DROP INDEX IF EXISTS extensions_get_by_modified;
DROP INDEX IF EXISTS extensions_get_by_created_and_modified;
DROP INDEX IF EXISTS audit_get_by_created;
DROP INDEX IF EXISTS audit_get_by_entity;
DROP INDEX IF EXISTS audit_get_by_operation;
DROP INDEX IF EXISTS audit_get_by_entity_and_operation;
DROP INDEX IF EXISTS project_organization_mappings_get_by_project_id;
DROP INDEX IF EXISTS project_organization_mappings_get_by_organization_id;
DROP INDEX IF EXISTS project_organization_mappings_get_by_project_id_and_organization_id;
DROP INDEX IF EXISTS project_organization_mappings_get_by_created;
DROP INDEX IF EXISTS project_organization_mappings_get_by_deleted;
DROP INDEX IF EXISTS project_organization_mappings_get_by_modified;
DROP INDEX IF EXISTS project_organization_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_ticket_type_id;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_ticket_type_id_and_project_id;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_created;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_modified;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS ticket_type_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS audit_meta_data_get_by_audit_id;
DROP INDEX IF EXISTS audit_meta_data_get_by_property;
DROP INDEX IF EXISTS audit_meta_data_get_by_audit_id_and_property;
DROP INDEX IF EXISTS audit_meta_data_get_by_value;
DROP INDEX IF EXISTS audit_meta_data_get_by_created;
DROP INDEX IF EXISTS audit_meta_data_get_by_modified;
DROP INDEX IF EXISTS audit_meta_data_get_by_created_and_modified;
DROP INDEX IF EXISTS reports_meta_data_get_by_report_id;
DROP INDEX IF EXISTS reports_meta_data_get_by_property;
DROP INDEX IF EXISTS reports_meta_data_get_by_report_id_and_property;
DROP INDEX IF EXISTS reports_meta_data_get_by_value;
DROP INDEX IF EXISTS reports_meta_data_get_by_report_id_and_value;
DROP INDEX IF EXISTS reports_meta_data_get_by_report_id_and_property_and_value;
DROP INDEX IF EXISTS reports_meta_data_get_by_created;
DROP INDEX IF EXISTS reports_meta_data_get_by_modified;
DROP INDEX IF EXISTS reports_meta_data_get_by_created_and_modified;
DROP INDEX IF EXISTS boards_meta_data_get_by_board_id;
DROP INDEX IF EXISTS boards_meta_data_get_by_property;
DROP INDEX IF EXISTS boards_meta_data_get_by_value;
DROP INDEX IF EXISTS boards_meta_data_get_by_board_id_and_property;
DROP INDEX IF EXISTS boards_meta_data_get_by_board_id_and_value;
DROP INDEX IF EXISTS boards_meta_data_get_by_board_id_and_property_and_value;
DROP INDEX IF EXISTS boards_meta_data_get_by_created;
DROP INDEX IF EXISTS boards_meta_data_get_by_modified;
DROP INDEX IF EXISTS boards_meta_data_get_by_created_and_modified;
DROP INDEX IF EXISTS tickets_meta_data_get_by_ticket_id;
DROP INDEX IF EXISTS tickets_meta_data_get_by_property;
DROP INDEX IF EXISTS tickets_meta_data_get_by_value;
DROP INDEX IF EXISTS tickets_meta_data_get_by_ticket_id_and_property;
DROP INDEX IF EXISTS tickets_meta_data_get_by_ticket_id_and_value;
DROP INDEX IF EXISTS tickets_meta_data_get_by_ticket_id_and_property_and_value;
DROP INDEX IF EXISTS tickets_meta_data_get_by_property_and_value;
DROP INDEX IF EXISTS tickets_meta_data_get_by_deleted;
DROP INDEX IF EXISTS tickets_meta_data_get_by_created;
DROP INDEX IF EXISTS tickets_meta_data_get_by_modified;
DROP INDEX IF EXISTS tickets_meta_data_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_relationships_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_child_ticket_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_child_ticket_id_and_child_ticket_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_ticket_relationship_type_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_ticket_id_and_ticket_relationship_type_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_ticket_id_and_child_ticket_id_and_ticket_relationship_type_id;
DROP INDEX IF EXISTS ticket_relationships_get_by_deleted;
DROP INDEX IF EXISTS ticket_relationships_get_by_created;
DROP INDEX IF EXISTS ticket_relationships_get_by_modified;
DROP INDEX IF EXISTS ticket_relationships_get_by_created_and_modified;
DROP INDEX IF EXISTS team_organization_mappings_get_by_team_id;
DROP INDEX IF EXISTS team_organization_mappings_get_by_organization_id;
DROP INDEX IF EXISTS team_organization_mappings_get_by_deleted;
DROP INDEX IF EXISTS team_organization_mappings_get_by_created;
DROP INDEX IF EXISTS team_organization_mappings_get_by_modified;
DROP INDEX IF EXISTS team_organization_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS team_project_mappings_get_by_team_id;
DROP INDEX IF EXISTS team_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS team_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS team_project_mappings_get_by_created;
DROP INDEX IF EXISTS team_project_mappings_get_by_modified;
DROP INDEX IF EXISTS team_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS repository_project_mappings_get_by_repository_id;
DROP INDEX IF EXISTS repository_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS repository_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS repository_project_mappings_get_by_created;
DROP INDEX IF EXISTS repository_project_mappings_get_by_modified;
DROP INDEX IF EXISTS repository_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_repository_id;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_commit_hash;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_ticket_id_commit_hash;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id_commit_hash;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_deleted;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_created;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_modified;
DROP INDEX IF EXISTS repository_commit_ticket_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_component_id;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_deleted;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_created;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_modified;
DROP INDEX IF EXISTS component_ticket_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS components_meta_data_get_by_component_id;
DROP INDEX IF EXISTS components_meta_data_get_by_property;
DROP INDEX IF EXISTS components_meta_data_get_by_component_id_and_property;
DROP INDEX IF EXISTS components_meta_data_get_by_value;
DROP INDEX IF EXISTS components_meta_data_get_by_component_id_and_value;
DROP INDEX IF EXISTS components_meta_data_get_by_property_and_value;
DROP INDEX IF EXISTS components_meta_data_get_by_component_id_and_property_and_value;
DROP INDEX IF EXISTS components_meta_data_get_by_deleted;
DROP INDEX IF EXISTS components_meta_data_get_by_created;
DROP INDEX IF EXISTS components_meta_data_get_by_modified;
DROP INDEX IF EXISTS components_meta_data_get_by_created_and_modified;
DROP INDEX IF EXISTS asset_project_mappings_get_by_asset_id;
DROP INDEX IF EXISTS asset_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS asset_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS asset_project_mappings_get_by_created;
DROP INDEX IF EXISTS asset_project_mappings_get_by_modified;
DROP INDEX IF EXISTS asset_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS asset_team_mappings_get_by_asset_id;
DROP INDEX IF EXISTS asset_team_mappings_get_by_team_id;
DROP INDEX IF EXISTS asset_team_mappings_get_by_deleted;
DROP INDEX IF EXISTS asset_team_mappings_get_by_created;
DROP INDEX IF EXISTS asset_team_mappings_get_by_modified;
DROP INDEX IF EXISTS asset_team_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_asset_id;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_deleted;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_created;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_modified;
DROP INDEX IF EXISTS asset_ticket_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_asset_id;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_comment_id;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_deleted;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_created;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_modified;
DROP INDEX IF EXISTS asset_comment_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_label_category_id;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_deleted;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_created;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_modified;
DROP INDEX IF EXISTS label_label_category_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS label_project_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS label_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS label_project_mappings_get_by_created;
DROP INDEX IF EXISTS label_project_mappings_get_by_modified;
DROP INDEX IF EXISTS label_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS label_team_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_team_mappings_get_by_team_id;
DROP INDEX IF EXISTS label_team_mappings_get_by_deleted;
DROP INDEX IF EXISTS label_team_mappings_get_by_created;
DROP INDEX IF EXISTS label_team_mappings_get_by_modified;
DROP INDEX IF EXISTS label_team_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_team_id;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_deleted;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_created;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_modified;
DROP INDEX IF EXISTS label_ticket_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS label_asset_mappings_get_by_label_id;
DROP INDEX IF EXISTS label_asset_mappings_get_by_team_id;
DROP INDEX IF EXISTS label_asset_mappings_get_by_deleted;
DROP INDEX IF EXISTS label_asset_mappings_get_by_created;
DROP INDEX IF EXISTS label_asset_mappings_get_by_modified;
DROP INDEX IF EXISTS label_asset_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_comment_id;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_deleted;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_created;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_modified;
DROP INDEX IF EXISTS comment_ticket_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_created;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_modified;
DROP INDEX IF EXISTS ticket_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_project_id;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_cycle_id;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_deleted;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_created;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_modified;
DROP INDEX IF EXISTS cycle_project_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_cycle_id;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_deleted;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_created;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_modified;
DROP INDEX IF EXISTS ticket_cycle_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_ticket_id;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_bord_id;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_deleted;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_created;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_modified;
DROP INDEX IF EXISTS ticket_board_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS users_default_mappings_get_by_user_id;
DROP INDEX IF EXISTS users_default_mappings_get_by_username;
DROP INDEX IF EXISTS users_default_mappings_get_by_username_and_secret;
DROP INDEX IF EXISTS users_default_mappings_get_by_deleted;
DROP INDEX IF EXISTS users_default_mappings_get_by_created;
DROP INDEX IF EXISTS users_default_mappings_get_by_modified;
DROP INDEX IF EXISTS users_default_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS user_organization_mappings_get_by_user_id;
DROP INDEX IF EXISTS user_organization_mappings_get_by_organization_id;
DROP INDEX IF EXISTS user_organization_mappings_get_by_deleted;
DROP INDEX IF EXISTS user_organization_mappings_get_by_created;
DROP INDEX IF EXISTS user_organization_mappings_get_by_modified;
DROP INDEX IF EXISTS user_organization_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS user_team_mappings_get_by_user_id;
DROP INDEX IF EXISTS user_team_mappings_get_by_team_id;
DROP INDEX IF EXISTS user_team_mappings_get_by_deleted;
DROP INDEX IF EXISTS user_team_mappings_get_by_created;
DROP INDEX IF EXISTS user_team_mappings_get_by_modified;
DROP INDEX IF EXISTS user_team_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS permission_user_mappings_get_by_user_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_permission_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_permission_context_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_user_id_and_permission_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_user_id_and_permission_context_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_permission_id_and_permission_context_id;
DROP INDEX IF EXISTS permission_user_mappings_get_by_deleted;
DROP INDEX IF EXISTS permission_user_mappings_get_by_created;
DROP INDEX IF EXISTS permission_user_mappings_get_by_modified;
DROP INDEX IF EXISTS permission_user_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS permission_team_mappings_get_by_team_id;
DROP INDEX IF EXISTS permission_team_mappings_get_by_permission_id;
DROP INDEX IF EXISTS permission_team_mappings_get_by_team_id_and_permission_id;
DROP INDEX IF EXISTS permission_team_mappings_get_by_permission_context_id;
DROP INDEX IF EXISTS permission_team_mappings_get_by_team_id_and_permission_context_id;
DROP INDEX IF EXISTS permission_team_mappings_get_by_deleted;
DROP INDEX IF EXISTS permission_team_mappings_get_by_created;
DROP INDEX IF EXISTS permission_team_mappings_get_by_modified;
DROP INDEX IF EXISTS permission_team_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_extension_id;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_property;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_value;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_property_and_value;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_extension_id_and_property;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_extension_id_and_property_and_value;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_enabled;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_deleted;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_created;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_modified;
DROP INDEX IF EXISTS configuration_data_extension_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS extensions_meta_data_get_by_extension_id;
DROP INDEX IF EXISTS extensions_meta_data_get_by_property;
DROP INDEX IF EXISTS extensions_meta_data_get_by_value;
DROP INDEX IF EXISTS extensions_meta_data_get_by_property_and_value;
DROP INDEX IF EXISTS extensions_meta_data_get_by_extension_id_and_property_and_value;
DROP INDEX IF EXISTS extensions_meta_data_get_by_extension_id_and_property;
DROP INDEX IF EXISTS extensions_meta_data_get_by_deleted;
DROP INDEX IF EXISTS extensions_meta_data_get_by_created;
DROP INDEX IF EXISTS extensions_meta_data_get_by_modified;
DROP INDEX IF EXISTS extensions_meta_data_get_by_created_and_modified;

/*
  Identifies the version of the database (system).
  After each SQL script execution the version will be increased and execution description provided.
*/
CREATE TABLE system_info
(

    id          INTEGER PRIMARY KEY UNIQUE,
    description TEXT    NOT NULL,
    created     INTEGER NOT NULL
);

CREATE INDEX system_info_get_by_created ON system_info (created);
CREATE INDEX system_info_get_by_description ON system_info (description);
CREATE INDEX system_info_get_by_created_and_description ON system_info (created, description);

/*
    The system entities:
*/

/*
    The basic project definition.

    Notes:
        - The 'workflow_id' represents the assigned workflow. Workflow is mandatory for the project.
        - The 'identifier' represents the human readable identifier for the project up to 4 characters,
            for example: MSF, KSS, etc.
*/
CREATE TABLE project
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    identifier  TEXT    NOT NULL UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    workflow_id TEXT    NOT NULL,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX projects_get_by_identifier ON project (identifier);
CREATE INDEX projects_get_by_title ON project (title);
CREATE INDEX projects_get_by_description ON project (description);
CREATE INDEX projects_get_by_title_and_description ON project (title, description);
CREATE INDEX projects_get_by_workflow_id ON project (workflow_id);
CREATE INDEX projects_get_by_created ON project (created);
CREATE INDEX projects_get_by_modified ON project (modified);
CREATE INDEX projects_get_by_deleted ON project (deleted);
CREATE INDEX projects_get_by_created_and_modified ON project (created, modified);

/*
    Ticket type definitions.
*/
CREATE TABLE ticket_type
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX ticket_types_get_by_title ON ticket_type (title);
CREATE INDEX ticket_types_get_by_description ON ticket_type (description);
CREATE INDEX ticket_types_get_by_title_and_description ON ticket_type (title, description);
CREATE INDEX ticket_types_get_by_created ON ticket_type (created);
CREATE INDEX ticket_types_get_by_modified ON ticket_type (modified);
CREATE INDEX ticket_types_get_by_deleted ON ticket_type (deleted);
CREATE INDEX ticket_types_get_by_created_and_modified ON ticket_type (created, modified);

/*
    Ticket statuses.
    For example:
        - To-do
        - Selected for development
        - In progress
        - Completed, etc.
*/
CREATE TABLE ticket_status
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX ticket_statuses_get_by_title ON ticket_status (title);
CREATE INDEX ticket_statuses_get_by_description ON ticket_status (description);
CREATE INDEX ticket_statuses_get_by_title_and_description ON ticket_status (title, description);
CREATE INDEX ticket_statuses_get_by_deleted ON ticket_status (deleted);
CREATE INDEX ticket_statuses_get_by_created ON ticket_status (created);
CREATE INDEX ticket_statuses_get_by_modified ON ticket_status (modified);
CREATE INDEX ticket_statuses_get_by_created_and_modified ON ticket_status (created, modified);

/*
    Tickets.
    Tickets belong to the project.
    Each ticket has its ticket type anf children if supported.
    The 'estimation' is the estimation value in man days.
    The 'story_points' represent the complexity of the ticket in story points (for example: 1, 3, 5, 8, 13).

    Notes:
        - The 'user_id' is the current owner of the ticket.
            It can be bull - the ticket is unassigned.
        - The 'creator' is the user id of the ticket creator.
        - The ticket number is human readable identifier of the ticket - the whole number.
            The ticket number is unique per project.
            In combination with 'project's 'identifier' field it can give the whole ticket numbers (identifiers),
            for example: MSF-112, BBP-222, etc.
*/
CREATE TABLE ticket
(

    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_number    INTEGER NOT NULL,
    position         INTEGER NOT NULL,
    title            TEXT,
    description      TEXT,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    ticket_type_id   TEXT    NOT NULL,
    ticket_status_id TEXT    NOT NULL,
    project_id       TEXT    NOT NULL,
    user_id          TEXT,
    estimation       REAL    NOT NULL,
    story_points     INTEGER NOT NULL,
    creator          TEXT    NOT NULL,
    deleted          BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_number, project_id) ON CONFLICT ABORT
);

CREATE INDEX tickets_get_by_ticket_number ON ticket (ticket_number);
CREATE INDEX tickets_get_by_ticket_type_id ON ticket (ticket_type_id);
CREATE INDEX tickets_get_by_ticket_status_id ON ticket (ticket_status_id);
CREATE INDEX tickets_get_by_project_id ON ticket (project_id);
CREATE INDEX tickets_get_by_user_id ON ticket (user_id);
CREATE INDEX tickets_get_by_creator ON ticket (creator);
CREATE INDEX tickets_get_by_project_id_and_user_id ON ticket (project_id, user_id);
CREATE INDEX tickets_get_by_project_id_and_creator ON ticket (project_id, creator);
CREATE INDEX tickets_get_by_estimation ON ticket (estimation);
CREATE INDEX tickets_get_by_story_points ON ticket (story_points);
CREATE INDEX tickets_get_by_created ON ticket (created);
CREATE INDEX tickets_get_by_modified ON ticket (modified);
CREATE INDEX tickets_get_by_deleted ON ticket (deleted);
CREATE INDEX tickets_get_by_created_and_modified ON ticket (created, modified);
CREATE INDEX tickets_get_by_title ON ticket (title);
CREATE INDEX tickets_get_by_description ON ticket (description);
CREATE INDEX tickets_get_by_title_and_description ON ticket (title, description);

/*
    Ticket relationship types.
    For example:
        - Blocked by
        - Blocks
        - Cloned by
        - Clones, etc.
*/
CREATE TABLE ticket_relationship_type
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX ticket_relationship_types_get_by_title ON ticket_relationship_type (title);
CREATE INDEX ticket_relationship_types_get_by_description ON ticket_relationship_type (description);
CREATE INDEX ticket_relationship_types_get_by_title_and_description ON ticket_relationship_type (title, description);
CREATE INDEX ticket_relationship_types_get_by_created ON ticket_relationship_type (created);
CREATE INDEX ticket_relationship_types_get_by_deleted ON ticket_relationship_type (deleted);
CREATE INDEX ticket_relationship_types_get_by_created_and_modified ON ticket_relationship_type (created, modified);

/*
    Ticket boards.
    Tickets belong to the board.
    Ticket may belong or may not belong to certain board. It is not mandatory.

    Boards examples:
        - Backlog
        - Main board
*/
CREATE TABLE board
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX boards_get_by_title ON board (title);
CREATE INDEX boards_get_by_description ON board (description);
CREATE INDEX boards_get_by_title_and_description ON board (title, description);
CREATE INDEX boards_get_by_created ON board (created);
CREATE INDEX boards_get_by_modified ON board (modified);
CREATE INDEX boards_get_by_deleted ON board (deleted);
CREATE INDEX boards_get_by_created_and_modified ON board (created, modified);

/*
    Workflows.
    The workflow represents a ordered set of steps (statuses) for the tickets that are connected to each other.
*/
CREATE TABLE workflow
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX workflows_get_by_title ON workflow (title);
CREATE INDEX workflows_get_by_description ON workflow (description);
CREATE INDEX workflows_get_by_title_and_description ON workflow (title, description);
CREATE INDEX workflows_get_by_created ON workflow (created);
CREATE INDEX workflows_get_by_modified ON workflow (modified);
CREATE INDEX workflows_get_by_deleted ON workflow (deleted);
CREATE INDEX workflows_get_by_created_and_modified ON workflow (created, modified);

/*
    Images, attachments, etc.
    Defined by the identifier and the resource url.
*/
CREATE TABLE asset
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    url         TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX assets_get_by_url ON asset (url);
CREATE INDEX assets_get_by_description ON asset (description);
CREATE INDEX assets_get_by_created ON asset (created);
CREATE INDEX assets_get_by_deleted ON asset (deleted);
CREATE INDEX assets_get_by_modified ON asset (modified);
CREATE INDEX assets_get_by_created_and_modified ON asset (created, modified);

/*
    Labels.
    Label can be associated with the almost everything:
        - Project
        - Team
        - Ticket
        - Asset
*/
CREATE TABLE label
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX labels_get_by_title ON label (title);
CREATE INDEX labels_get_by_description ON label (description);
CREATE INDEX labels_get_by_title_and_description ON label (title, description);
CREATE INDEX labels_get_by_created ON label (created);
CREATE INDEX labels_get_by_deleted ON label (deleted);
CREATE INDEX labels_get_by_modified ON label (modified);
CREATE INDEX labels_get_by_created_and_modified ON label (created, modified);

/*
    Labels can be divided into categories (which is optional).
*/
CREATE TABLE label_category
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX label_categories_get_by_title ON label_category (title);
CREATE INDEX label_categories_get_by_description ON label_category (description);
CREATE INDEX label_categories_get_by_title_and_description ON label_category (title, description);
CREATE INDEX label_categories_get_by_created ON label_category (created);
CREATE INDEX label_categories_get_by_deleted ON label_category (deleted);
CREATE INDEX label_categories_get_by_modified ON label_category (modified);
CREATE INDEX label_categories_get_by_created_and_modified ON label_category (created, modified);

/*
      The code repositories - Identified by the identifier and the repository URL.
      Default repository type is Git repository.
*/
CREATE TABLE repository
(

    id                 TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository         TEXT    NOT NULL UNIQUE,
    description        TEXT,
    repository_type_id TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    modified           INTEGER NOT NULL,
    deleted            BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX repositories_get_by_repository ON repository (repository);
CREATE INDEX repositories_get_by_description ON repository (description);
CREATE INDEX repositories_get_by_repository_and_description ON repository (repository, description);
CREATE INDEX repositories_get_by_deleted ON repository (deleted);
CREATE INDEX repositories_get_by_repository_type_id ON repository (repository_type_id);
CREATE INDEX repositories_get_by_created ON repository (created);
CREATE INDEX repositories_get_by_modified ON repository (modified);
CREATE INDEX repositories_get_by_created_and_modified ON repository (created, modified);

/*
    'Git', 'CVS', 'SVN', 'Mercurial',
  'Perforce', 'Monotone', 'Bazaar',
  'TFS', 'VSTS', 'IBM Rational ClearCase',
  'Revision Control System', 'VSS',
  'CA Harvest Software Change Manager',
  'PVCS', 'darcs'
*/
CREATE TABLE repository_type
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX repository_types_get_by_title ON repository_type (title);
CREATE INDEX repository_types_get_by_description ON repository_type (description);
CREATE INDEX repository_types_get_by_title_and_description ON repository_type (title, description);
CREATE INDEX repository_types_get_by_deleted ON repository_type (deleted);
CREATE INDEX repository_types_get_by_created ON repository_type (created);
CREATE INDEX repository_types_get_by_modified ON repository_type (modified);
CREATE INDEX repository_types_get_by_created_and_modified ON repository_type (created, modified);

/*
    Components.
    Components are associated with the tickets.
    For example:
        - Backend
        - Android Client
        - Core Engine
        - Webapp, etc.
*/
CREATE TABLE component
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX components_get_by_title ON component (title);
CREATE INDEX components_get_by_description ON component (description);
CREATE INDEX components_get_by_title_description ON component (title, description);
CREATE INDEX components_get_by_created ON component (created);
CREATE INDEX components_get_by_deleted ON component (deleted);
CREATE INDEX components_get_by_modified ON component (modified);
CREATE INDEX components_get_by_created_modified ON component (created, modified);

/*
    The organization definition. Organization is the owner of the project.
*/
CREATE TABLE organization
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX organizations_get_by_title ON organization (title);
CREATE INDEX organizations_get_by_description ON organization (description);
CREATE INDEX organizations_get_by_title_and_description ON organization (title, description);
CREATE INDEX organizations_get_by_created ON organization (created);
CREATE INDEX organizations_get_by_deleted ON organization (deleted);
CREATE INDEX organizations_get_by_modified ON organization (modified);
CREATE INDEX organizations_get_by_created_and_modified ON organization (created, modified);

/*
    The team definition. Organization is the owner of the team.
*/
CREATE TABLE team
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX teams_get_by_title ON team (title);
CREATE INDEX teams_get_by_description ON team (description);
CREATE INDEX teams_get_by_title_and_description ON team (title, description);
CREATE INDEX teams_get_by_created ON team (created);
CREATE INDEX teams_get_by_modified ON team (modified);
CREATE INDEX teams_get_by_deleted ON team (deleted);
CREATE INDEX teams_get_by_created_and_modified ON team (created, modified);

/*
    Permission definitions.
    Permissions are (for example):

        CREATE
        UPDATE
        DELETE
        etc.
*/
CREATE TABLE permission
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX permissions_get_by_title ON permission (title);
CREATE INDEX permissions_get_by_description ON permission (description);
CREATE INDEX permissions_get_by_title_and_description ON permission (title, description);
CREATE INDEX permissions_get_by_deleted ON permission (deleted);
CREATE INDEX permissions_get_by_created ON permission (created);
CREATE INDEX permissions_get_by_modified ON permission (modified);
CREATE INDEX permissions_get_by_created_and_modified ON permission (created, modified);

/*
    Comments.
    Users can comment on:
        - Tickets
        - Tbd.
*/
CREATE TABLE comment
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX comments_get_by_comment ON comment (comment);
CREATE INDEX comments_get_by_created ON comment (created);
CREATE INDEX comments_get_by_modified ON comment (modified);
CREATE INDEX comments_get_by_deleted ON comment (deleted);
CREATE INDEX comments_get_by_created_and_modified ON comment (created, modified);

/*
    Permission contexts.
    Each permission must assigned to the permission owner must have a valid context.
    Permission contexts are (for example):

        organization.project
        organization.team
*/
CREATE TABLE permission_context
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX permission_contexts_get_by_title ON permission_context (title);
CREATE INDEX permission_contexts_get_by_description ON permission_context (description);
CREATE INDEX permission_contexts_get_by_title_and_description ON permission_context (title, description);
CREATE INDEX permission_contexts_get_by_created ON permission_context (created);
CREATE INDEX permission_contexts_get_by_modified ON permission_context (modified);
CREATE INDEX permission_contexts_get_by_deleted ON permission_context (deleted);
CREATE INDEX permission_contexts_get_by_created_and_modified ON permission_context (created, modified);

/*
    Workflow steps.
    Steps for the workflow that are linked to each other.

    Notes:
        - The 'workflow_step_id' is the parent step. The root steps (for example: 'to-do') have no parents.
        - The 'ticket_status_id' represents the status (connection with it) that will be assigned to the ticket once
            the ticket gets to the workflow step.
*/
CREATE TABLE workflow_step
(

    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title            TEXT    NOT NULL UNIQUE,
    description      TEXT,
    workflow_id      TEXT    NOT NULL,
    workflow_step_id TEXT,
    ticket_status_id TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX workflow_steps_get_by_title ON workflow_step (title);
CREATE INDEX workflow_steps_get_by_description ON workflow_step (description);
CREATE INDEX workflow_steps_get_by_title_and_description ON workflow_step (title, description);
CREATE INDEX workflow_steps_get_by_workflow_id ON workflow_step (workflow_id);
CREATE INDEX workflow_steps_get_by_workflow_step_id ON workflow_step (workflow_step_id);
CREATE INDEX workflow_steps_get_by_ticket_status_id ON workflow_step (ticket_status_id);
CREATE INDEX workflow_steps_get_by_workflow_id_and_ticket_status_id ON workflow_step (workflow_id, ticket_status_id);
CREATE INDEX workflow_steps_get_by_workflow_id_and_workflow_step_id ON workflow_step (workflow_id, workflow_step_id);

CREATE INDEX workflow_steps_get_by_workflow_id_and_workflow_step_id_and_ticket_status_id ON workflow_step
    (workflow_id, workflow_step_id, ticket_status_id);

CREATE INDEX workflow_steps_get_by_created ON workflow_step (created);
CREATE INDEX workflow_steps_get_by_deleted ON workflow_step (deleted);
CREATE INDEX workflow_steps_get_by_modified ON workflow_step (modified);
CREATE INDEX workflow_steps_get_by_created_and_modified ON workflow_step (created, modified);

/*
    Reports, such as:
        - Time tracking reports
        - Progress status(es), etc.
*/
CREATE TABLE report
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX reports_get_by_title ON report (title);
CREATE INDEX reports_get_by_description ON report (description);
CREATE INDEX reports_get_by_title_and_description ON report (title, description);
CREATE INDEX reports_get_by_created ON report (created);
CREATE INDEX reports_get_by_deleted ON report (deleted);
CREATE INDEX reports_get_by_modified ON report (modified);
CREATE INDEX reports_get_by_created_and_modified ON report (created, modified);

/*
    Contains the information about all work cycles in the system.
    Cycle belongs to the project. To only one project.
    Ticket belongs to the cycle. Ticket can belong to the multiple cycles.

    Work cycle types:
        - Release (top cycle category, not mandatory)
        - Milestone (middle cycle category, not mandatory)
        - Sprint (smaller cycle category, not mandatory)

    Milestones may or may not belong to the release.
    Sprints may or may not belong to milestones or releases.
    Releases may or may not have the version associated.
    Each cycle may have different meta data associated.

    Each cycle has the value:
        - Release       = 1000
        - Milestone     = 100
        - Sprint        = 10

    To illustrate its relationship.
    Based on this future custom cycle types could be supported.

    Cycle can belong to only one parent.
    Parent's type integer value mus be > than the type integer value of current cycle (this).
*/
CREATE TABLE cycle
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    /*
      Parent cycle id.
     */
    cycle_id    TEXT    NOT NULL UNIQUE,
    /*
        CHECK ( type IN (1000, 100, 10))
    */
    type        INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX cycles_get_by_title ON cycle (title);
CREATE INDEX cycles_get_by_description ON cycle (description);
CREATE INDEX cycles_get_by_title_and_description ON cycle (title, description);
CREATE INDEX cycles_get_by_cycle_id ON cycle (cycle_id);
CREATE INDEX cycles_get_by_type ON cycle (type);
CREATE INDEX cycles_get_by_cycle_id_and_type ON cycle (cycle_id, type);
CREATE INDEX cycles_get_by_created ON cycle (created);
CREATE INDEX cycles_get_by_deleted ON cycle (deleted);
CREATE INDEX cycles_get_by_modified ON cycle (modified);
CREATE INDEX cycles_get_by_created_and_modified ON cycle (created, modified);

/*
  The 3rd party extensions.
  Each extension is identified by the 'extension_key' which is properly verified by the system.
  Extension can be enabled or disabled - the 'enabled' field.
*/
CREATE TABLE extension
(

    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    title         TEXT,
    description   TEXT,
    extension_key TEXT    NOT NULL UNIQUE,
    enabled       BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    deleted       BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX extensions_get_by_title ON extension (title);
CREATE INDEX extensions_get_by_description ON extension (description);
CREATE INDEX extensions_get_by_title_and_description ON extension (title, description);
CREATE INDEX extensions_get_by_extension_key ON extension (extension_key);
CREATE INDEX extensions_get_by_created ON extension (created);
CREATE INDEX extensions_get_by_deleted ON extension (deleted);
CREATE INDEX extensions_get_by_enabled ON extension (enabled);
CREATE INDEX extensions_get_by_modified ON extension (modified);
CREATE INDEX extensions_get_by_created_and_modified ON extension (created, modified);

/*
    Audit trail.
*/
CREATE TABLE audit
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created   INTEGER NOT NULL,
    entity    TEXT,
    operation TEXT
);

CREATE INDEX audit_get_by_created ON audit (created);
CREATE INDEX audit_get_by_entity ON audit (entity);
CREATE INDEX audit_get_by_operation ON audit (operation);
CREATE INDEX audit_get_by_entity_and_operation ON audit (entity, operation);

/*
    Mappings:
*/

/*
    Project belongs to the organization. Multiple projects can belong to one organization.
*/
CREATE TABLE project_organization_mapping
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (project_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX project_organization_mappings_get_by_project_id ON project_organization_mapping (project_id);
CREATE INDEX project_organization_mappings_get_by_organization_id ON project_organization_mapping (organization_id);

CREATE INDEX project_organization_mappings_get_by_project_id_and_organization_id ON
    project_organization_mapping (project_id, organization_id);

CREATE INDEX project_organization_mappings_get_by_created ON project_organization_mapping (created);
CREATE INDEX project_organization_mappings_get_by_deleted ON project_organization_mapping (deleted);
CREATE INDEX project_organization_mappings_get_by_modified ON project_organization_mapping (modified);

CREATE INDEX project_organization_mappings_get_by_created_and_modified ON
    project_organization_mapping (created, modified);

/*
    Each project has the ticket types that it supports.
*/
CREATE TABLE ticket_type_project_mapping
(

    id             TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id TEXT    NOT NULL,
    project_id     TEXT    NOT NULL,
    created        INTEGER NOT NULL,
    modified       INTEGER NOT NULL,
    deleted        BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_type_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_type_project_mappings_get_by_ticket_type_id ON ticket_type_project_mapping (ticket_type_id);
CREATE INDEX ticket_type_project_mappings_get_by_project_id ON ticket_type_project_mapping (project_id);

CREATE INDEX ticket_type_project_mappings_get_by_ticket_type_id_and_project_id
    ON ticket_type_project_mapping (ticket_type_id, project_id);

CREATE INDEX ticket_type_project_mappings_get_by_created ON ticket_type_project_mapping (created);
CREATE INDEX ticket_type_project_mappings_get_by_modified ON ticket_type_project_mapping (modified);
CREATE INDEX ticket_type_project_mappings_get_by_deleted ON ticket_type_project_mapping (deleted);

CREATE INDEX ticket_type_project_mappings_get_by_created_and_modified
    ON ticket_type_project_mapping (created, modified);

/*
    Audit trail meta-data.
*/
CREATE TABLE audit_meta_data
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    audit_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL
);

CREATE INDEX audit_meta_data_get_by_audit_id ON audit_meta_data (audit_id);
CREATE INDEX audit_meta_data_get_by_property ON audit_meta_data (property);
CREATE INDEX audit_meta_data_get_by_audit_id_and_property ON audit_meta_data (audit_id, property);
CREATE INDEX audit_meta_data_get_by_value ON audit_meta_data (value);
CREATE INDEX audit_meta_data_get_by_created ON audit_meta_data (created);
CREATE INDEX audit_meta_data_get_by_modified ON audit_meta_data (modified);
CREATE INDEX audit_meta_data_get_by_created_and_modified ON audit_meta_data (created, modified);

/*
   Reports met-data: used to populate reports with the information.
*/
CREATE TABLE report_meta_data
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    report_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL
);

CREATE INDEX reports_meta_data_get_by_report_id ON report_meta_data (report_id);
CREATE INDEX reports_meta_data_get_by_property ON report_meta_data (property);
CREATE INDEX reports_meta_data_get_by_report_id_and_property ON report_meta_data (report_id, property);
CREATE INDEX reports_meta_data_get_by_value ON report_meta_data (value);
CREATE INDEX reports_meta_data_get_by_report_id_and_value ON report_meta_data (report_id, value);
CREATE INDEX reports_meta_data_get_by_report_id_and_property_and_value ON report_meta_data (report_id, property, value);
CREATE INDEX reports_meta_data_get_by_created ON report_meta_data (created);
CREATE INDEX reports_meta_data_get_by_modified ON report_meta_data (modified);
CREATE INDEX reports_meta_data_get_by_created_and_modified ON report_meta_data (created, modified);

/*
   Boards meta-data: additional data that can be associated with certain board.
*/
CREATE TABLE board_meta_data
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL
);

CREATE INDEX boards_meta_data_get_by_board_id ON board_meta_data (board_id);
CREATE INDEX boards_meta_data_get_by_property ON board_meta_data (property);
CREATE INDEX boards_meta_data_get_by_value ON board_meta_data (value);
CREATE INDEX boards_meta_data_get_by_board_id_and_property ON board_meta_data (board_id, property);
CREATE INDEX boards_meta_data_get_by_board_id_and_value ON board_meta_data (board_id, value);
CREATE INDEX boards_meta_data_get_by_board_id_and_property_and_value ON board_meta_data (board_id, property, value);
CREATE INDEX boards_meta_data_get_by_created ON board_meta_data (created);
CREATE INDEX boards_meta_data_get_by_modified ON board_meta_data (modified);
CREATE INDEX boards_meta_data_get_by_created_and_modified ON board_meta_data (created, modified);

/*
    Tickets meta-data.
*/
CREATE TABLE ticket_meta_data
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX tickets_meta_data_get_by_ticket_id ON ticket_meta_data (ticket_id);
CREATE INDEX tickets_meta_data_get_by_property ON ticket_meta_data (property);
CREATE INDEX tickets_meta_data_get_by_value ON ticket_meta_data (value);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_property ON ticket_meta_data (ticket_id, property);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_value ON ticket_meta_data (ticket_id, value);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_property_and_value ON ticket_meta_data (ticket_id, property, value);
CREATE INDEX tickets_meta_data_get_by_property_and_value ON ticket_meta_data (property, value);
CREATE INDEX tickets_meta_data_get_by_deleted ON ticket_meta_data (deleted);
CREATE INDEX tickets_meta_data_get_by_created ON ticket_meta_data (created);
CREATE INDEX tickets_meta_data_get_by_modified ON ticket_meta_data (modified);
CREATE INDEX tickets_meta_data_get_by_created_and_modified ON ticket_meta_data (created, modified);

/*
    All relationships between the tickets.
*/
CREATE TABLE ticket_relationship
(

    id                          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_relationship_type_id TEXT    NOT NULL,
    ticket_id                   TEXT    NOT NULL,
    child_ticket_id             TEXT    NOT NULL,
    created                     INTEGER NOT NULL,
    modified                    INTEGER NOT NULL,
    deleted                     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, child_ticket_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_relationships_get_by_ticket_id ON ticket_relationship (ticket_id);
CREATE INDEX ticket_relationships_get_by_child_ticket_id ON ticket_relationship (child_ticket_id);

CREATE INDEX ticket_relationships_get_by_child_ticket_id_and_child_ticket_id
    ON ticket_relationship (ticket_id, child_ticket_id);

CREATE INDEX ticket_relationships_get_by_ticket_relationship_type_id
    ON ticket_relationship (ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationship (ticket_id, ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_ticket_id_and_child_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationship (ticket_id, child_ticket_id, ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_deleted ON ticket_relationship (deleted);
CREATE INDEX ticket_relationships_get_by_created ON ticket_relationship (created);
CREATE INDEX ticket_relationships_get_by_modified ON ticket_relationship (modified);
CREATE INDEX ticket_relationships_get_by_created_and_modified ON ticket_relationship (created, modified);

/*
    Team belongs to the organization. Multiple teams can belong to one organization.
*/
CREATE TABLE team_organization_mapping
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (team_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX team_organization_mappings_get_by_team_id ON team_organization_mapping (team_id);
CREATE INDEX team_organization_mappings_get_by_organization_id ON team_organization_mapping (organization_id);
CREATE INDEX team_organization_mappings_get_by_deleted ON team_organization_mapping (deleted);
CREATE INDEX team_organization_mappings_get_by_created ON team_organization_mapping (created);
CREATE INDEX team_organization_mappings_get_by_modified ON team_organization_mapping (modified);
CREATE INDEX team_organization_mappings_get_by_created_and_modified ON team_organization_mapping (created, modified);

/*
    Team belongs to one or more projects. Multiple teams can work on multiple projects.
*/
CREATE TABLE team_project_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id    TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (team_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX team_project_mappings_get_by_team_id ON team_project_mapping (team_id);
CREATE INDEX team_project_mappings_get_by_project_id ON team_project_mapping (project_id);
CREATE INDEX team_project_mappings_get_by_deleted ON team_project_mapping (deleted);
CREATE INDEX team_project_mappings_get_by_created ON team_project_mapping (created);
CREATE INDEX team_project_mappings_get_by_modified ON team_project_mapping (modified);
CREATE INDEX team_project_mappings_get_by_created_and_modified ON team_project_mapping (created, modified);

/*
     Repository belongs to project. Multiple repositories can belong to multiple projects.
     So, two projects can actually have the same repository.
*/
CREATE TABLE repository_project_mapping
(

    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    project_id    TEXT    NOT NULL,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (repository_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX repository_project_mappings_get_by_repository_id ON repository_project_mapping (repository_id);
CREATE INDEX repository_project_mappings_get_by_project_id ON repository_project_mapping (project_id);
CREATE INDEX repository_project_mappings_get_by_deleted ON repository_project_mapping (deleted);
CREATE INDEX repository_project_mappings_get_by_created ON repository_project_mapping (created);
CREATE INDEX repository_project_mappings_get_by_modified ON repository_project_mapping (modified);
CREATE INDEX repository_project_mappings_get_by_created_and_modified ON repository_project_mapping (created, modified);

/*
     Mapping all commits to the corresponding tickets
*/
CREATE TABLE repository_commit_ticket_mapping
(

    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    ticket_id     TEXT    NOT NULL,
    commit_hash   TEXT    NOT NULL UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id
    ON repository_commit_ticket_mapping (repository_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id ON repository_commit_ticket_mapping (ticket_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id
    ON repository_commit_ticket_mapping (repository_id, ticket_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_commit_hash ON repository_commit_ticket_mapping (commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id_commit_hash
    ON repository_commit_ticket_mapping (ticket_id, commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id_commit_hash
    ON repository_commit_ticket_mapping (repository_id, ticket_id, commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_deleted ON repository_commit_ticket_mapping (deleted);
CREATE INDEX repository_commit_ticket_mappings_get_by_created ON repository_commit_ticket_mapping (created);
CREATE INDEX repository_commit_ticket_mappings_get_by_modified ON repository_commit_ticket_mapping (modified);

CREATE INDEX repository_commit_ticket_mappings_get_by_created_and_modified
    ON repository_commit_ticket_mapping (created, modified);

/*
    Components to the tickets mappings.
    Component can be mapped to the multiple tickets.
*/
CREATE TABLE component_ticket_mapping
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    ticket_id    TEXT    NOT NULL,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (component_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX component_ticket_mappings_get_by_ticket_id ON component_ticket_mapping (ticket_id);
CREATE INDEX component_ticket_mappings_get_by_component_id ON component_ticket_mapping (component_id);
CREATE INDEX component_ticket_mappings_get_by_deleted ON component_ticket_mapping (deleted);
CREATE INDEX component_ticket_mappings_get_by_created ON component_ticket_mapping (created);
CREATE INDEX component_ticket_mappings_get_by_modified ON component_ticket_mapping (modified);
CREATE INDEX component_ticket_mappings_get_by_created_and_modified ON component_ticket_mapping (created, modified);

/*
    Components meta-data:
    Associate the various information with different components.
*/
CREATE TABLE component_meta_data
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX components_meta_data_get_by_component_id ON component_meta_data (component_id);
CREATE INDEX components_meta_data_get_by_property ON component_meta_data (property);
CREATE INDEX components_meta_data_get_by_component_id_and_property ON component_meta_data (component_id, property);
CREATE INDEX components_meta_data_get_by_value ON component_meta_data (value);
CREATE INDEX components_meta_data_get_by_component_id_and_value ON component_meta_data (component_id, value);
CREATE INDEX components_meta_data_get_by_property_and_value ON component_meta_data (property, value);

CREATE INDEX components_meta_data_get_by_component_id_and_property_and_value
    ON component_meta_data (component_id, property, value);

CREATE INDEX components_meta_data_get_by_deleted ON component_meta_data (deleted);
CREATE INDEX components_meta_data_get_by_created ON component_meta_data (created);
CREATE INDEX components_meta_data_get_by_modified ON component_meta_data (modified);
CREATE INDEX components_meta_data_get_by_created_and_modified ON component_meta_data (created, modified);

/*
    Assets can belong to the multiple projects.
    One example of the image used in the context of the project is the project's avatar.
    Projects may have various other assets associated to itself.
    Various documentation for example.
*/
CREATE TABLE asset_project_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX asset_project_mappings_get_by_asset_id ON asset_project_mapping (asset_id);
CREATE INDEX asset_project_mappings_get_by_project_id ON asset_project_mapping (project_id);
CREATE INDEX asset_project_mappings_get_by_deleted ON asset_project_mapping (deleted);
CREATE INDEX asset_project_mappings_get_by_created ON asset_project_mapping (created);
CREATE INDEX asset_project_mappings_get_by_modified ON asset_project_mapping (modified);
CREATE INDEX asset_project_mappings_get_by_created_and_modified ON asset_project_mapping (created, modified);

/*
    Assets can belong to the multiple teams.
    The image used in the context of the team is the team's avatar, for example.
    Teams may have other additions associated to itself. Various documents for example,
*/
CREATE TABLE asset_team_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX asset_team_mappings_get_by_asset_id ON asset_team_mapping (asset_id);
CREATE INDEX asset_team_mappings_get_by_team_id ON asset_team_mapping (team_id);
CREATE INDEX asset_team_mappings_get_by_deleted ON asset_team_mapping (deleted);
CREATE INDEX asset_team_mappings_get_by_created ON asset_team_mapping (created);
CREATE INDEX asset_team_mappings_get_by_modified ON asset_team_mapping (modified);
CREATE INDEX asset_team_mappings_get_by_created_and_modified ON asset_team_mapping (created, modified);

/*
    Assets (attachments for example) can belong to the multiple tickets.
*/
CREATE TABLE asset_ticket_mapping
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX asset_ticket_mappings_get_by_asset_id ON asset_ticket_mapping (asset_id);
CREATE INDEX asset_ticket_mappings_get_by_ticket_id ON asset_ticket_mapping (ticket_id);
CREATE INDEX asset_ticket_mappings_get_by_deleted ON asset_ticket_mapping (deleted);
CREATE INDEX asset_ticket_mappings_get_by_created ON asset_ticket_mapping (created);
CREATE INDEX asset_ticket_mappings_get_by_modified ON asset_ticket_mapping (modified);
CREATE INDEX asset_ticket_mappings_get_by_created_and_modified ON asset_ticket_mapping (created, modified);

/*
    Assets (attachments for example) can belong to the multiple comments.
*/
CREATE TABLE asset_comment_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    comment_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, comment_id) ON CONFLICT ABORT
);

CREATE INDEX asset_comment_mappings_get_by_asset_id ON asset_comment_mapping (asset_id);
CREATE INDEX asset_comment_mappings_get_by_comment_id ON asset_comment_mapping (comment_id);
CREATE INDEX asset_comment_mappings_get_by_deleted ON asset_comment_mapping (deleted);
CREATE INDEX asset_comment_mappings_get_by_created ON asset_comment_mapping (created);
CREATE INDEX asset_comment_mappings_get_by_modified ON asset_comment_mapping (modified);
CREATE INDEX asset_comment_mappings_get_by_created_and_modified ON asset_comment_mapping (created, modified);

/*
    Labels can belong to the label category.
    One single asset can belong to multiple categories.
*/
CREATE TABLE label_label_category_mapping
(

    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id          TEXT    NOT NULL,
    label_category_id TEXT    NOT NULL,
    created           INTEGER NOT NULL,
    modified          INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, label_category_id) ON CONFLICT ABORT
);

CREATE INDEX label_label_category_mappings_get_by_label_id ON label_label_category_mapping (label_id);

CREATE INDEX label_label_category_mappings_get_by_label_category_id
    ON label_label_category_mapping (label_category_id);

CREATE INDEX label_label_category_mappings_get_by_deleted ON label_label_category_mapping (deleted);
CREATE INDEX label_label_category_mappings_get_by_created ON label_label_category_mapping (created);
CREATE INDEX label_label_category_mappings_get_by_modified ON label_label_category_mapping (modified);

CREATE INDEX label_label_category_mappings_get_by_created_and_modified
    ON label_label_category_mapping (created, modified);

/*
    Label can be associated with one or more projects.
*/
CREATE TABLE label_project_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX label_project_mappings_get_by_label_id ON label_project_mapping (label_id);
CREATE INDEX label_project_mappings_get_by_project_id ON label_project_mapping (project_id);
CREATE INDEX label_project_mappings_get_by_deleted ON label_project_mapping (deleted);
CREATE INDEX label_project_mappings_get_by_created ON label_project_mapping (created);
CREATE INDEX label_project_mappings_get_by_modified ON label_project_mapping (modified);
CREATE INDEX label_project_mappings_get_by_created_and_modified ON label_project_mapping (created, modified);

/*
    Label can be associated with one or more teams.
*/
CREATE TABLE label_team_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX label_team_mappings_get_by_label_id ON label_team_mapping (label_id);
CREATE INDEX label_team_mappings_get_by_team_id ON label_team_mapping (team_id);
CREATE INDEX label_team_mappings_get_by_deleted ON label_team_mapping (deleted);
CREATE INDEX label_team_mappings_get_by_created ON label_team_mapping (created);
CREATE INDEX label_team_mappings_get_by_modified ON label_team_mapping (modified);
CREATE INDEX label_team_mappings_get_by_created_and_modified ON label_team_mapping (created, modified);

/*
    Label can be associated with one or more tickets.
*/
CREATE TABLE label_ticket_mapping
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX label_ticket_mappings_get_by_label_id ON label_ticket_mapping (label_id);
CREATE INDEX label_ticket_mappings_get_by_team_id ON label_ticket_mapping (ticket_id);
CREATE INDEX label_ticket_mappings_get_by_deleted ON label_ticket_mapping (deleted);
CREATE INDEX label_ticket_mappings_get_by_created ON label_ticket_mapping (created);
CREATE INDEX label_ticket_mappings_get_by_modified ON label_ticket_mapping (modified);
CREATE INDEX label_ticket_mappings_get_by_created_and_modified ON label_ticket_mapping (created, modified);

/*
    Label can be associated with one or more assets.
*/
CREATE TABLE label_asset_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    asset_id TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, asset_id) ON CONFLICT ABORT
);

CREATE INDEX label_asset_mappings_get_by_label_id ON label_asset_mapping (label_id);
CREATE INDEX label_asset_mappings_get_by_team_id ON label_asset_mapping (asset_id);
CREATE INDEX label_asset_mappings_get_by_deleted ON label_asset_mapping (deleted);
CREATE INDEX label_asset_mappings_get_by_created ON label_asset_mapping (created);
CREATE INDEX label_asset_mappings_get_by_modified ON label_asset_mapping (modified);
CREATE INDEX label_asset_mappings_get_by_created_and_modified ON label_asset_mapping (created, modified);

/*
    Comments are usually associated with project tickets:
*/
CREATE TABLE comment_ticket_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id TEXT    NOT NULL,
    ticket_id  TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (comment_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX comment_ticket_mappings_get_by_comment_id ON comment_ticket_mapping (comment_id);
CREATE INDEX comment_ticket_mappings_get_by_ticket_id ON comment_ticket_mapping (ticket_id);
CREATE INDEX comment_ticket_mappings_get_by_deleted ON comment_ticket_mapping (deleted);
CREATE INDEX comment_ticket_mappings_get_by_created ON comment_ticket_mapping (created);
CREATE INDEX comment_ticket_mappings_get_by_modified ON comment_ticket_mapping (modified);
CREATE INDEX comment_ticket_mappings_get_by_created_and_modified ON comment_ticket_mapping (created, modified);

/*
    Tickets belong to the project:
*/
CREATE TABLE ticket_project_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_project_mappings_get_by_project_id ON ticket_project_mapping (project_id);
CREATE INDEX ticket_project_mappings_get_by_ticket_id ON ticket_project_mapping (ticket_id);
CREATE INDEX ticket_project_mappings_get_by_deleted ON ticket_project_mapping (deleted);
CREATE INDEX ticket_project_mappings_get_by_created ON ticket_project_mapping (created);
CREATE INDEX ticket_project_mappings_get_by_modified ON ticket_project_mapping (modified);
CREATE INDEX ticket_project_mappings_get_by_created_and_modified ON ticket_project_mapping (created, modified);

/*
    Cycles belong to the projects.
    Cycle can belong to exactly one project.
*/
CREATE TABLE cycle_project_mapping
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    cycle_id   TEXT    NOT NULL UNIQUE,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (cycle_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX cycle_project_mappings_get_by_project_id ON cycle_project_mapping (project_id);
CREATE INDEX cycle_project_mappings_get_by_cycle_id ON cycle_project_mapping (cycle_id);
CREATE INDEX cycle_project_mappings_get_by_deleted ON cycle_project_mapping (deleted);
CREATE INDEX cycle_project_mappings_get_by_created ON cycle_project_mapping (created);
CREATE INDEX cycle_project_mappings_get_by_modified ON cycle_project_mapping (modified);
CREATE INDEX cycle_project_mappings_get_by_created_and_modified ON cycle_project_mapping (created, modified);

/*
    Tickets can belong to cycles:
*/
CREATE TABLE ticket_cycle_mapping
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    cycle_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, cycle_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_cycle_mappings_get_by_ticket_id ON ticket_cycle_mapping (ticket_id);
CREATE INDEX ticket_cycle_mappings_get_by_cycle_id ON ticket_cycle_mapping (cycle_id);
CREATE INDEX ticket_cycle_mappings_get_by_deleted ON ticket_cycle_mapping (deleted);
CREATE INDEX ticket_cycle_mappings_get_by_created ON ticket_cycle_mapping (created);
CREATE INDEX ticket_cycle_mappings_get_by_modified ON ticket_cycle_mapping (modified);
CREATE INDEX ticket_cycle_mappings_get_by_created_and_modified ON ticket_cycle_mapping (created, modified);

/*
    Tickets can belong to one or more boards:
*/
CREATE TABLE ticket_board_mapping
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    board_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, board_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_board_mappings_get_by_ticket_id ON ticket_board_mapping (ticket_id);
CREATE INDEX ticket_board_mappings_get_by_bord_id ON ticket_board_mapping (board_id);
CREATE INDEX ticket_board_mappings_get_by_deleted ON ticket_board_mapping (deleted);
CREATE INDEX ticket_board_mappings_get_by_created ON ticket_board_mapping (created);
CREATE INDEX ticket_board_mappings_get_by_modified ON ticket_board_mapping (modified);
CREATE INDEX ticket_board_mappings_get_by_created_and_modified ON ticket_board_mapping (created, modified);

/*
    Default mapping for users (default auth.)
    The 'secret' field represnts the salted/hashed password value.
*/
CREATE TABLE user_default_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    secret   TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX users_default_mappings_get_by_user_id ON user_default_mapping (user_id);
CREATE INDEX users_default_mappings_get_by_username ON user_default_mapping (username);
CREATE INDEX users_default_mappings_get_by_username_and_secret ON user_default_mapping (username, secret);
CREATE INDEX users_default_mappings_get_by_deleted ON user_default_mapping (deleted);
CREATE INDEX users_default_mappings_get_by_created ON user_default_mapping (created);
CREATE INDEX users_default_mappings_get_by_modified ON user_default_mapping (modified);
CREATE INDEX users_default_mappings_get_by_created_and_modified ON user_default_mapping (created, modified);

/*
    User access rights:
*/

/*
    User belongs to organizations:
*/
CREATE TABLE user_organization_mapping
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (user_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX user_organization_mappings_get_by_user_id ON user_organization_mapping (user_id);
CREATE INDEX user_organization_mappings_get_by_organization_id ON user_organization_mapping (organization_id);
CREATE INDEX user_organization_mappings_get_by_deleted ON user_organization_mapping (deleted);
CREATE INDEX user_organization_mappings_get_by_created ON user_organization_mapping (created);
CREATE INDEX user_organization_mappings_get_by_modified ON user_organization_mapping (modified);
CREATE INDEX user_organization_mappings_get_by_created_and_modified ON user_organization_mapping (created, modified);

/*
    User belongs to the organization's teams:
*/
CREATE TABLE user_team_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (user_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX user_team_mappings_get_by_user_id ON user_team_mapping (user_id);
CREATE INDEX user_team_mappings_get_by_team_id ON user_team_mapping (team_id);
CREATE INDEX user_team_mappings_get_by_deleted ON user_team_mapping (deleted);
CREATE INDEX user_team_mappings_get_by_created ON user_team_mapping (created);
CREATE INDEX user_team_mappings_get_by_modified ON user_team_mapping (modified);
CREATE INDEX user_team_mappings_get_by_created_and_modified ON user_team_mapping (created, modified);

/*
    User has the permissions.
    Each permission has be associated to the proper permission context.
*/
CREATE TABLE permission_user_mapping
(

    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    user_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (user_id, permission_id, permission_context_id) ON CONFLICT ABORT
);

CREATE INDEX permission_user_mappings_get_by_user_id ON permission_user_mapping (user_id);
CREATE INDEX permission_user_mappings_get_by_permission_id ON permission_user_mapping (permission_id);
CREATE INDEX permission_user_mappings_get_by_permission_context_id ON permission_user_mapping (permission_context_id);

CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_id
    ON permission_user_mapping (user_id, permission_id);

CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_context_id
    ON permission_user_mapping (user_id, permission_context_id);

CREATE INDEX permission_user_mappings_get_by_permission_id_and_permission_context_id
    ON permission_user_mapping (permission_id, permission_context_id);

CREATE INDEX permission_user_mappings_get_by_deleted ON permission_user_mapping (deleted);
CREATE INDEX permission_user_mappings_get_by_created ON permission_user_mapping (created);
CREATE INDEX permission_user_mappings_get_by_modified ON permission_user_mapping (modified);
CREATE INDEX permission_user_mappings_get_by_created_and_modified ON permission_user_mapping (created, modified);

/*
    Team has the permissions.
    Each team permission has be associated to the proper permission context.
    All team members (users) will inherit team's permissions.
*/
CREATE TABLE permission_team_mapping
(

    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    team_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (team_id, permission_id, permission_context_id) ON CONFLICT ABORT
);

CREATE INDEX permission_team_mappings_get_by_team_id ON permission_team_mapping (team_id);
CREATE INDEX permission_team_mappings_get_by_permission_id ON permission_team_mapping (permission_id);

CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_id
    ON permission_team_mapping (team_id, permission_id);

CREATE INDEX permission_team_mappings_get_by_permission_context_id ON permission_team_mapping (permission_context_id);

CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_context_id
    ON permission_team_mapping (team_id, permission_context_id);

CREATE INDEX permission_team_mappings_get_by_deleted ON permission_team_mapping (deleted);
CREATE INDEX permission_team_mappings_get_by_created ON permission_team_mapping (created);
CREATE INDEX permission_team_mappings_get_by_modified ON permission_team_mapping (modified);
CREATE INDEX permission_team_mappings_get_by_created_and_modified ON permission_team_mapping (created, modified);

/*
    The configuration data for the extension.
    Basically it represents the meta-data associated with each extension.
    Each configuration property can be enabled or disabled.
*/
CREATE TABLE configuration_data_extension_mapping
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    enabled      BOOLEAN NOT NULL CHECK (enabled IN (0, 1)),
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX configuration_data_extension_mappings_get_by_extension_id
    ON configuration_data_extension_mapping (extension_id);

CREATE INDEX configuration_data_extension_mappings_get_by_property ON configuration_data_extension_mapping (property);
CREATE INDEX configuration_data_extension_mappings_get_by_value ON configuration_data_extension_mapping (value);

CREATE INDEX configuration_data_extension_mappings_get_by_property_and_value
    ON configuration_data_extension_mapping (property, value);

CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property
    ON configuration_data_extension_mapping (extension_id, property);

CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property_and_value
    ON configuration_data_extension_mapping (extension_id, property, value);

CREATE INDEX configuration_data_extension_mappings_get_by_enabled ON configuration_data_extension_mapping (enabled);
CREATE INDEX configuration_data_extension_mappings_get_by_deleted ON configuration_data_extension_mapping (deleted);
CREATE INDEX configuration_data_extension_mappings_get_by_created ON configuration_data_extension_mapping (created);
CREATE INDEX configuration_data_extension_mappings_get_by_modified ON configuration_data_extension_mapping (modified);

CREATE INDEX configuration_data_extension_mappings_get_by_created_and_modified
    ON configuration_data_extension_mapping (created, modified);

/*
    Extensions meta-data:
    Associate the various information with different extension.
    Meta-data information are the extension specific.
*/
CREATE TABLE extension_meta_data
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX extensions_meta_data_get_by_extension_id ON extension_meta_data (extension_id);
CREATE INDEX extensions_meta_data_get_by_property ON extension_meta_data (property);
CREATE INDEX extensions_meta_data_get_by_value ON extension_meta_data (value);
CREATE INDEX extensions_meta_data_get_by_property_and_value ON extension_meta_data (property, value);

CREATE INDEX extensions_meta_data_get_by_extension_id_and_property_and_value
    ON extension_meta_data (extension_id, property, value);

CREATE INDEX extensions_meta_data_get_by_extension_id_and_property ON extension_meta_data (extension_id, property);
CREATE INDEX extensions_meta_data_get_by_deleted ON extension_meta_data (deleted);
CREATE INDEX extensions_meta_data_get_by_created ON extension_meta_data (created);
CREATE INDEX extensions_meta_data_get_by_modified ON extension_meta_data (modified);
CREATE INDEX extensions_meta_data_get_by_created_and_modified ON extension_meta_data (created, modified);
