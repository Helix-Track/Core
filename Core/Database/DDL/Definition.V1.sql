/*
    Version: 1
*/

/*
    Notes:

    - TODOs: https://github.com/orgs/red-elf/projects/2/views/1
    - Identifiers in the system are UUID strings.
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
*/

DROP TABLE IF EXISTS system_info;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS team_organization_mappings;
DROP TABLE IF EXISTS team_project_mappings;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS user_organization_mappings;
DROP TABLE IF EXISTS user_team_mappings;
DROP TABLE IF EXISTS users_yandex_mappings;
DROP TABLE IF EXISTS users_google_mappings;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS project_organization_mappings;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_project_mappings;
DROP TABLE IF EXISTS ticket_cycle_mappings;
DROP TABLE IF EXISTS ticket_board_mappings;
DROP TABLE IF EXISTS ticket_types;
DROP TABLE IF EXISTS ticket_statuses;
DROP TABLE IF EXISTS ticket_relationship_types;
DROP TABLE IF EXISTS ticket_relationships;
DROP TABLE IF EXISTS ticket_type_project_mappings;
DROP TABLE IF EXISTS boards;
DROP TABLE IF EXISTS boards_meta_data;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS workflow_steps;
DROP TABLE IF EXISTS cycles;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS label_categories;
DROP TABLE IF EXISTS label_label_category_mappings;
DROP TABLE IF EXISTS label_ticket_mappings;
DROP TABLE IF EXISTS label_asset_mappings;
DROP TABLE IF EXISTS label_team_mappings;
DROP TABLE IF EXISTS label_project_mappings;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS repository_project_mappings;
DROP TABLE IF EXISTS repository_commit_ticket_mappings;
DROP TABLE IF EXISTS components;
DROP TABLE IF EXISTS component_ticket_mappings;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS permission_contexts;
DROP TABLE IF EXISTS audit;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS extensions;
DROP TABLE IF EXISTS permission_user_mappings;
DROP TABLE IF EXISTS audit_meta_data;
DROP TABLE IF EXISTS reports_meta_data;
DROP TABLE IF EXISTS tickets_meta_data;
DROP TABLE IF EXISTS components_meta_data;
DROP TABLE IF EXISTS asset_project_mappings;
DROP TABLE IF EXISTS asset_team_mappings;
DROP TABLE IF EXISTS asset_ticket_mappings;
DROP TABLE IF EXISTS asset_comment_mappings;
DROP TABLE IF EXISTS comment_ticket_mappings;
DROP TABLE IF EXISTS cycle_project_mappings;
DROP TABLE IF EXISTS permission_team_mappings;
DROP TABLE IF EXISTS configuration_data_extension_mappings;
DROP TABLE IF EXISTS extensions_meta_data;

DROP INDEX IF EXISTS system_info_get_by_created;
DROP INDEX IF EXISTS system_info_get_by_description;
DROP INDEX IF EXISTS system_info_get_by_created_and_description;
DROP INDEX IF EXISTS users_get_by_created;
DROP INDEX IF EXISTS users_get_by_modified;
DROP INDEX IF EXISTS users_get_by_deleted;
DROP INDEX IF EXISTS users_get_by_created_and_modified;
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
DROP INDEX IF EXISTS repositories_get_by_type;
DROP INDEX IF EXISTS repositories_get_by_created;
DROP INDEX IF EXISTS repositories_get_by_modified;
DROP INDEX IF EXISTS repositories_get_by_created_and_modified;
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
DROP INDEX IF EXISTS users_yandex_mappings_get_by_user_id;
DROP INDEX IF EXISTS users_yandex_mappings_get_by_username;
DROP INDEX IF EXISTS users_yandex_mappings_get_by_deleted;
DROP INDEX IF EXISTS users_yandex_mappings_get_by_created;
DROP INDEX IF EXISTS users_yandex_mappings_get_by_modified;
DROP INDEX IF EXISTS users_yandex_mappings_get_by_created_and_modified;
DROP INDEX IF EXISTS users_google_mappings_get_by_user_id;
DROP INDEX IF EXISTS users_google_mappings_get_by_username;
DROP INDEX IF EXISTS users_google_mappings_get_by_deleted;
DROP INDEX IF EXISTS users_google_mappings_get_by_created;
DROP INDEX IF EXISTS users_google_mappings_get_by_modified;
DROP INDEX IF EXISTS users_google_mappings_get_by_created_and_modified;
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
     System's users.
     User is identified by the unique identifier (id).
     Since there may be different types of users, different kinds of data
     can be mapped (associated) with the user ID.
     For that purpose there are other mappings to the user ID such as Yandex OAuth2 mappings for example.
*/
CREATE TABLE users
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX users_get_by_created ON users (created);
CREATE INDEX users_get_by_modified ON users (modified);
CREATE INDEX users_get_by_deleted ON users (deleted);
CREATE INDEX users_get_by_created_and_modified ON users (created, modified);

/*
    The basic project definition.

    Notes:
        - The 'workflow_id' represents the assigned workflow. Workflow is mandatory for the project.
        - The 'identifier' represents the human readable identifier for the project up to 4 characters,
            for example: MSF, KSS, etc.
*/
CREATE TABLE projects
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

CREATE INDEX projects_get_by_identifier ON projects (identifier);
CREATE INDEX projects_get_by_title ON projects (title);
CREATE INDEX projects_get_by_description ON projects (description);
CREATE INDEX projects_get_by_title_and_description ON projects (title, description);
CREATE INDEX projects_get_by_workflow_id ON projects (workflow_id);
CREATE INDEX projects_get_by_created ON projects (created);
CREATE INDEX projects_get_by_modified ON projects (modified);
CREATE INDEX projects_get_by_deleted ON projects (deleted);
CREATE INDEX projects_get_by_created_and_modified ON projects (created, modified);

/*
    Ticket type definitions.
*/
CREATE TABLE ticket_types
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX ticket_types_get_by_title ON ticket_types (title);
CREATE INDEX ticket_types_get_by_description ON ticket_types (description);
CREATE INDEX ticket_types_get_by_title_and_description ON ticket_types (title, description);
CREATE INDEX ticket_types_get_by_created ON ticket_types (created);
CREATE INDEX ticket_types_get_by_modified ON ticket_types (modified);
CREATE INDEX ticket_types_get_by_deleted ON ticket_types (deleted);
CREATE INDEX ticket_types_get_by_created_and_modified ON ticket_types (created, modified);

/*
    Ticket statuses.
    For example:
        - To-do
        - Selected for development
        - In progress
        - Completed, etc.
*/
CREATE TABLE ticket_statuses
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX ticket_statuses_get_by_title ON ticket_statuses (title);
CREATE INDEX ticket_statuses_get_by_description ON ticket_statuses (description);
CREATE INDEX ticket_statuses_get_by_title_and_description ON ticket_statuses (title, description);
CREATE INDEX ticket_statuses_get_by_deleted ON ticket_statuses (deleted);
CREATE INDEX ticket_statuses_get_by_created ON ticket_statuses (created);
CREATE INDEX ticket_statuses_get_by_modified ON ticket_statuses (modified);
CREATE INDEX ticket_statuses_get_by_created_and_modified ON ticket_statuses (created, modified);

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
CREATE TABLE tickets
(

    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_number    INTEGER NOT NULL DEFAULT 1,
    title            TEXT,
    description      TEXT,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    ticket_type_id   TEXT    NOT NULL,
    ticket_status_id TEXT    NOT NULL,
    project_id       TEXT    NOT NULL,
    user_id          TEXT,
    estimation       REAL    NOT NULL DEFAULT 0,
    story_points     INTEGER NOT NULL DEFAULT 0,
    creator          TEXT    NOT NULL,
    deleted          BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_number, project_id) ON CONFLICT ABORT
);

CREATE INDEX tickets_get_by_ticket_number ON tickets (ticket_number);
CREATE INDEX tickets_get_by_ticket_type_id ON tickets (ticket_type_id);
CREATE INDEX tickets_get_by_ticket_status_id ON tickets (ticket_status_id);
CREATE INDEX tickets_get_by_project_id ON tickets (project_id);
CREATE INDEX tickets_get_by_user_id ON tickets (user_id);
CREATE INDEX tickets_get_by_creator ON tickets (creator);
CREATE INDEX tickets_get_by_project_id_and_user_id ON tickets (project_id, user_id);
CREATE INDEX tickets_get_by_project_id_and_creator ON tickets (project_id, creator);
CREATE INDEX tickets_get_by_estimation ON tickets (estimation);
CREATE INDEX tickets_get_by_story_points ON tickets (story_points);
CREATE INDEX tickets_get_by_created ON tickets (created);
CREATE INDEX tickets_get_by_modified ON tickets (modified);
CREATE INDEX tickets_get_by_deleted ON tickets (deleted);
CREATE INDEX tickets_get_by_created_and_modified ON tickets (created, modified);
CREATE INDEX tickets_get_by_title ON tickets (title);
CREATE INDEX tickets_get_by_description ON tickets (description);
CREATE INDEX tickets_get_by_title_and_description ON tickets (title, description);

/*
    Ticket relationship types.
    For example:
        - Blocked by
        - Blocks
        - Cloned by
        - Clones, etc.
*/
CREATE TABLE ticket_relationship_types
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX ticket_relationship_types_get_by_title ON ticket_relationship_types (title);
CREATE INDEX ticket_relationship_types_get_by_description ON ticket_relationship_types (description);
CREATE INDEX ticket_relationship_types_get_by_title_and_description ON ticket_relationship_types (title, description);
CREATE INDEX ticket_relationship_types_get_by_created ON ticket_relationship_types (created);
CREATE INDEX ticket_relationship_types_get_by_deleted ON ticket_relationship_types (deleted);
CREATE INDEX ticket_relationship_types_get_by_created_and_modified ON ticket_relationship_types (created, modified);

/*
    Ticket boards.
    Tickets belong to the board.
    Ticket may belong or may not belong to certain board. It is not mandatory.

    Boards examples:
        - Backlog
        - Main board
*/
CREATE TABLE boards
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX boards_get_by_title ON boards (title);
CREATE INDEX boards_get_by_description ON boards (description);
CREATE INDEX boards_get_by_title_and_description ON boards (title, description);
CREATE INDEX boards_get_by_created ON boards (created);
CREATE INDEX boards_get_by_modified ON boards (modified);
CREATE INDEX boards_get_by_deleted ON boards (deleted);
CREATE INDEX boards_get_by_created_and_modified ON boards (created, modified);

/*
    Workflows.
    The workflow represents a ordered set of steps (statuses) for the tickets that are connected to each other.
*/
CREATE TABLE workflows
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX workflows_get_by_title ON workflows (title);
CREATE INDEX workflows_get_by_description ON workflows (description);
CREATE INDEX workflows_get_by_title_and_description ON workflows (title, description);
CREATE INDEX workflows_get_by_created ON workflows (created);
CREATE INDEX workflows_get_by_modified ON workflows (modified);
CREATE INDEX workflows_get_by_deleted ON workflows (deleted);
CREATE INDEX workflows_get_by_created_and_modified ON workflows (created, modified);

/*
    Images, attachments, etc.
    Defined by the identifier and the resource url.
*/
CREATE TABLE assets
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    url         TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX assets_get_by_url ON assets (url);
CREATE INDEX assets_get_by_description ON assets (description);
CREATE INDEX assets_get_by_created ON assets (created);
CREATE INDEX assets_get_by_deleted ON assets (deleted);
CREATE INDEX assets_get_by_modified ON assets (modified);
CREATE INDEX assets_get_by_created_and_modified ON assets (created, modified);

/*
    Labels.
    Label can be associated with the almost everything:
        - Project
        - Team
        - Ticket
        - Asset
*/
CREATE TABLE labels
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX labels_get_by_title ON labels (title);
CREATE INDEX labels_get_by_description ON labels (description);
CREATE INDEX labels_get_by_title_and_description ON labels (title, description);
CREATE INDEX labels_get_by_created ON labels (created);
CREATE INDEX labels_get_by_deleted ON labels (deleted);
CREATE INDEX labels_get_by_modified ON labels (modified);
CREATE INDEX labels_get_by_created_and_modified ON labels (created, modified);

/*
    Labels can be divided into categories (which is optional).
*/
CREATE TABLE label_categories
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX label_categories_get_by_title ON label_categories (title);
CREATE INDEX label_categories_get_by_description ON label_categories (description);
CREATE INDEX label_categories_get_by_title_and_description ON label_categories (title, description);
CREATE INDEX label_categories_get_by_created ON label_categories (created);
CREATE INDEX label_categories_get_by_deleted ON label_categories (deleted);
CREATE INDEX label_categories_get_by_modified ON label_categories (modified);
CREATE INDEX label_categories_get_by_created_and_modified ON label_categories (created, modified);

/*
      The code repositories - Identified by the identifier and the repository URL.
      Default repository type is Git repository.
*/
CREATE TABLE repositories
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository  TEXT    NOT NULL UNIQUE,
    description TEXT,

    type        TEXT CHECK ( type IN
                             ('Git', 'CVS', 'SVN', 'Mercurial',
                              'Perforce', 'Monotone', 'Bazaar',
                              'TFS', 'VSTS', 'IBM Rational ClearCase',
                              'Revision Control System', 'VSS',
                              'CA Harvest Software Change Manager',
                              'PVCS', 'darcs')
        )               NOT NULL DEFAULT 'Git',

    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX repositories_get_by_repository ON repositories (repository);
CREATE INDEX repositories_get_by_description ON repositories (description);
CREATE INDEX repositories_get_by_repository_and_description ON repositories (repository, description);
CREATE INDEX repositories_get_by_deleted ON repositories (deleted);
CREATE INDEX repositories_get_by_type ON repositories (type);
CREATE INDEX repositories_get_by_created ON repositories (created);
CREATE INDEX repositories_get_by_modified ON repositories (modified);
CREATE INDEX repositories_get_by_created_and_modified ON repositories (created, modified);

/*
    Components.
    Components are associated with the tickets.
    For example:
        - Backend
        - Android Client
        - Core Engine
        - Webapp, etc.
*/
CREATE TABLE components
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX components_get_by_title ON components (title);
CREATE INDEX components_get_by_description ON components (description);
CREATE INDEX components_get_by_title_description ON components (title, description);
CREATE INDEX components_get_by_created ON components (created);
CREATE INDEX components_get_by_deleted ON components (deleted);
CREATE INDEX components_get_by_modified ON components (modified);
CREATE INDEX components_get_by_created_modified ON components (created, modified);

/*
    The organization definition. Organization is the owner of the project.
*/
CREATE TABLE organizations
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX organizations_get_by_title ON organizations (title);
CREATE INDEX organizations_get_by_description ON organizations (description);
CREATE INDEX organizations_get_by_title_and_description ON organizations (title, description);
CREATE INDEX organizations_get_by_created ON organizations (created);
CREATE INDEX organizations_get_by_deleted ON organizations (deleted);
CREATE INDEX organizations_get_by_modified ON organizations (modified);
CREATE INDEX organizations_get_by_created_and_modified ON organizations (created, modified);

/*
    The team definition. Organization is the owner of the team.
*/
CREATE TABLE teams
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX teams_get_by_title ON teams (title);
CREATE INDEX teams_get_by_description ON teams (description);
CREATE INDEX teams_get_by_title_and_description ON teams (title, description);
CREATE INDEX teams_get_by_created ON teams (created);
CREATE INDEX teams_get_by_modified ON teams (modified);
CREATE INDEX teams_get_by_deleted ON teams (deleted);
CREATE INDEX teams_get_by_created_and_modified ON teams (created, modified);

/*
    Permission definitions.
    Permissions are (for example):

        CREATE
        UPDATE
        DELETE
        etc.
*/
CREATE TABLE permissions
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX permissions_get_by_title ON permissions (title);
CREATE INDEX permissions_get_by_description ON permissions (description);
CREATE INDEX permissions_get_by_title_and_description ON permissions (title, description);
CREATE INDEX permissions_get_by_deleted ON permissions (deleted);
CREATE INDEX permissions_get_by_created ON permissions (created);
CREATE INDEX permissions_get_by_modified ON permissions (modified);
CREATE INDEX permissions_get_by_created_and_modified ON permissions (created, modified);

/*
    Comments.
    Users can comment on:
        - Tickets
        - Tbd.
*/
CREATE TABLE comments
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX comments_get_by_comment ON comments (comment);
CREATE INDEX comments_get_by_created ON comments (created);
CREATE INDEX comments_get_by_modified ON comments (modified);
CREATE INDEX comments_get_by_deleted ON comments (deleted);
CREATE INDEX comments_get_by_created_and_modified ON comments (created, modified);

/*
    Permission contexts.
    Each permission must assigned to the permission owner must have a valid context.
    Permission contexts are (for example):

        organization.project
        organization.team
*/
CREATE TABLE permission_contexts
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX permission_contexts_get_by_title ON permission_contexts (title);
CREATE INDEX permission_contexts_get_by_description ON permission_contexts (description);
CREATE INDEX permission_contexts_get_by_title_and_description ON permission_contexts (title, description);
CREATE INDEX permission_contexts_get_by_created ON permission_contexts (created);
CREATE INDEX permission_contexts_get_by_modified ON permission_contexts (modified);
CREATE INDEX permission_contexts_get_by_deleted ON permission_contexts (deleted);
CREATE INDEX permission_contexts_get_by_created_and_modified ON permission_contexts (created, modified);

/*
    Workflow steps.
    Steps for the workflow that are linked to each other.

    Notes:
        - The 'workflow_step_id' is the parent step. The root steps (for example: 'to-do') have no parents.
        - The 'ticket_status_id' represents the status (connection with it) that will be assigned to the ticket once
            the ticket gets to the workflow step.
*/
CREATE TABLE workflow_steps
(

    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title            TEXT    NOT NULL UNIQUE,
    description      TEXT,
    workflow_id      TEXT    NOT NULL,
    workflow_step_id TEXT,
    ticket_status_id TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX workflow_steps_get_by_title ON workflow_steps (title);
CREATE INDEX workflow_steps_get_by_description ON workflow_steps (description);
CREATE INDEX workflow_steps_get_by_title_and_description ON workflow_steps (title, description);
CREATE INDEX workflow_steps_get_by_workflow_id ON workflow_steps (workflow_id);
CREATE INDEX workflow_steps_get_by_workflow_step_id ON workflow_steps (workflow_step_id);
CREATE INDEX workflow_steps_get_by_ticket_status_id ON workflow_steps (ticket_status_id);
CREATE INDEX workflow_steps_get_by_workflow_id_and_ticket_status_id ON workflow_steps (workflow_id, ticket_status_id);
CREATE INDEX workflow_steps_get_by_workflow_id_and_workflow_step_id ON workflow_steps (workflow_id, workflow_step_id);

CREATE INDEX workflow_steps_get_by_workflow_id_and_workflow_step_id_and_ticket_status_id ON workflow_steps
    (workflow_id, workflow_step_id, ticket_status_id);

CREATE INDEX workflow_steps_get_by_created ON workflow_steps (created);
CREATE INDEX workflow_steps_get_by_deleted ON workflow_steps (deleted);
CREATE INDEX workflow_steps_get_by_modified ON workflow_steps (modified);
CREATE INDEX workflow_steps_get_by_created_and_modified ON workflow_steps (created, modified);

/*
    Reports, such as:
        - Time tracking reports
        - Progress status(es), etc.
*/
CREATE TABLE reports
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX reports_get_by_title ON reports (title);
CREATE INDEX reports_get_by_description ON reports (description);
CREATE INDEX reports_get_by_title_and_description ON reports (title, description);
CREATE INDEX reports_get_by_created ON reports (created);
CREATE INDEX reports_get_by_deleted ON reports (deleted);
CREATE INDEX reports_get_by_modified ON reports (modified);
CREATE INDEX reports_get_by_created_and_modified ON reports (created, modified);

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
CREATE TABLE cycles
(

    id          TEXT                                     NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER                                  NOT NULL,
    modified    INTEGER                                  NOT NULL,
    title       TEXT,
    description TEXT,
    /**
      Prent cycle id.
     */
    cycle_id    TEXT                                     NOT NULL UNIQUE,
    type        INTEGER CHECK ( type IN (1000, 100, 10)) NOT NULL,
    deleted     BOOLEAN                                  NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX cycles_get_by_title ON cycles (title);
CREATE INDEX cycles_get_by_description ON cycles (description);
CREATE INDEX cycles_get_by_title_and_description ON cycles (title, description);
CREATE INDEX cycles_get_by_cycle_id ON cycles (cycle_id);
CREATE INDEX cycles_get_by_type ON cycles (type);
CREATE INDEX cycles_get_by_cycle_id_and_type ON cycles (cycle_id, type);
CREATE INDEX cycles_get_by_created ON cycles (created);
CREATE INDEX cycles_get_by_deleted ON cycles (deleted);
CREATE INDEX cycles_get_by_modified ON cycles (modified);
CREATE INDEX cycles_get_by_created_and_modified ON cycles (created, modified);

/*
  The 3rd party extensions.
  Each extension is identified by the 'extension_key' which is properly verified by the system.
  Extension can be enabled or disabled - the 'enabled' field.
*/
CREATE TABLE extensions
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

CREATE INDEX extensions_get_by_title ON extensions (title);
CREATE INDEX extensions_get_by_description ON extensions (description);
CREATE INDEX extensions_get_by_title_and_description ON extensions (title, description);
CREATE INDEX extensions_get_by_extension_key ON extensions (extension_key);
CREATE INDEX extensions_get_by_created ON extensions (created);
CREATE INDEX extensions_get_by_deleted ON extensions (deleted);
CREATE INDEX extensions_get_by_enabled ON extensions (enabled);
CREATE INDEX extensions_get_by_modified ON extensions (modified);
CREATE INDEX extensions_get_by_created_and_modified ON extensions (created, modified);

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
CREATE TABLE project_organization_mappings
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (project_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX project_organization_mappings_get_by_project_id ON project_organization_mappings (project_id);
CREATE INDEX project_organization_mappings_get_by_organization_id ON project_organization_mappings (organization_id);

CREATE INDEX project_organization_mappings_get_by_project_id_and_organization_id ON
    project_organization_mappings (project_id, organization_id);

CREATE INDEX project_organization_mappings_get_by_created ON project_organization_mappings (created);
CREATE INDEX project_organization_mappings_get_by_deleted ON project_organization_mappings (deleted);
CREATE INDEX project_organization_mappings_get_by_modified ON project_organization_mappings (modified);

CREATE INDEX project_organization_mappings_get_by_created_and_modified ON
    project_organization_mappings (created, modified);

/*
    Each project has the ticket types that it supports.
*/
CREATE TABLE ticket_type_project_mappings
(

    id             TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id TEXT    NOT NULL,
    project_id     TEXT    NOT NULL,
    created        INTEGER NOT NULL,
    modified       INTEGER NOT NULL,
    deleted        BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_type_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_type_project_mappings_get_by_ticket_type_id ON ticket_type_project_mappings (ticket_type_id);
CREATE INDEX ticket_type_project_mappings_get_by_project_id ON ticket_type_project_mappings (project_id);

CREATE INDEX ticket_type_project_mappings_get_by_ticket_type_id_and_project_id
    ON ticket_type_project_mappings (ticket_type_id, project_id);

CREATE INDEX ticket_type_project_mappings_get_by_created ON ticket_type_project_mappings (created);
CREATE INDEX ticket_type_project_mappings_get_by_modified ON ticket_type_project_mappings (modified);
CREATE INDEX ticket_type_project_mappings_get_by_deleted ON ticket_type_project_mappings (deleted);

CREATE INDEX ticket_type_project_mappings_get_by_created_and_modified
    ON ticket_type_project_mappings (created, modified);

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
CREATE TABLE reports_meta_data
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    report_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL
);

CREATE INDEX reports_meta_data_get_by_report_id ON reports_meta_data (report_id);
CREATE INDEX reports_meta_data_get_by_property ON reports_meta_data (property);
CREATE INDEX reports_meta_data_get_by_report_id_and_property ON reports_meta_data (report_id, property);
CREATE INDEX reports_meta_data_get_by_value ON reports_meta_data (value);
CREATE INDEX reports_meta_data_get_by_report_id_and_value ON reports_meta_data (report_id, value);
CREATE INDEX reports_meta_data_get_by_report_id_and_property_and_value ON reports_meta_data (report_id, property, value);
CREATE INDEX reports_meta_data_get_by_created ON reports_meta_data (created);
CREATE INDEX reports_meta_data_get_by_modified ON reports_meta_data (modified);
CREATE INDEX reports_meta_data_get_by_created_and_modified ON reports_meta_data (created, modified);

/*
   Boards meta-data: additional data that can be associated with certain board.
*/
CREATE TABLE boards_meta_data
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL
);

CREATE INDEX boards_meta_data_get_by_board_id ON boards_meta_data (board_id);
CREATE INDEX boards_meta_data_get_by_property ON boards_meta_data (property);
CREATE INDEX boards_meta_data_get_by_value ON boards_meta_data (value);
CREATE INDEX boards_meta_data_get_by_board_id_and_property ON boards_meta_data (board_id, property);
CREATE INDEX boards_meta_data_get_by_board_id_and_value ON boards_meta_data (board_id, value);
CREATE INDEX boards_meta_data_get_by_board_id_and_property_and_value ON boards_meta_data (board_id, property, value);
CREATE INDEX boards_meta_data_get_by_created ON boards_meta_data (created);
CREATE INDEX boards_meta_data_get_by_modified ON boards_meta_data (modified);
CREATE INDEX boards_meta_data_get_by_created_and_modified ON boards_meta_data (created, modified);

/*
    Tickets meta-data.
*/
CREATE TABLE tickets_meta_data
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX tickets_meta_data_get_by_ticket_id ON tickets_meta_data (ticket_id);
CREATE INDEX tickets_meta_data_get_by_property ON tickets_meta_data (property);
CREATE INDEX tickets_meta_data_get_by_value ON tickets_meta_data (value);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_property ON tickets_meta_data (ticket_id, property);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_value ON tickets_meta_data (ticket_id, value);
CREATE INDEX tickets_meta_data_get_by_ticket_id_and_property_and_value ON tickets_meta_data (ticket_id, property, value);
CREATE INDEX tickets_meta_data_get_by_property_and_value ON tickets_meta_data (property, value);
CREATE INDEX tickets_meta_data_get_by_deleted ON tickets_meta_data (deleted);
CREATE INDEX tickets_meta_data_get_by_created ON tickets_meta_data (created);
CREATE INDEX tickets_meta_data_get_by_modified ON tickets_meta_data (modified);
CREATE INDEX tickets_meta_data_get_by_created_and_modified ON tickets_meta_data (created, modified);

/*
    All relationships between the tickets.
*/
CREATE TABLE ticket_relationships
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

CREATE INDEX ticket_relationships_get_by_ticket_id ON ticket_relationships (ticket_id);
CREATE INDEX ticket_relationships_get_by_child_ticket_id ON ticket_relationships (child_ticket_id);

CREATE INDEX ticket_relationships_get_by_child_ticket_id_and_child_ticket_id
    ON ticket_relationships (ticket_id, child_ticket_id);

CREATE INDEX ticket_relationships_get_by_ticket_relationship_type_id
    ON ticket_relationships (ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationships (ticket_id, ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_ticket_id_and_child_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationships (ticket_id, child_ticket_id, ticket_relationship_type_id);

CREATE INDEX ticket_relationships_get_by_deleted ON ticket_relationships (deleted);
CREATE INDEX ticket_relationships_get_by_created ON ticket_relationships (created);
CREATE INDEX ticket_relationships_get_by_modified ON ticket_relationships (modified);
CREATE INDEX ticket_relationships_get_by_created_and_modified ON ticket_relationships (created, modified);

/*
    Team belongs to the organization. Multiple teams can belong to one organization.
*/
CREATE TABLE team_organization_mappings
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (team_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX team_organization_mappings_get_by_team_id ON team_organization_mappings (team_id);
CREATE INDEX team_organization_mappings_get_by_organization_id ON team_organization_mappings (organization_id);
CREATE INDEX team_organization_mappings_get_by_deleted ON team_organization_mappings (deleted);
CREATE INDEX team_organization_mappings_get_by_created ON team_organization_mappings (created);
CREATE INDEX team_organization_mappings_get_by_modified ON team_organization_mappings (modified);
CREATE INDEX team_organization_mappings_get_by_created_and_modified ON team_organization_mappings (created, modified);

/*
    Team belongs to one or more projects. Multiple teams can work on multiple projects.
*/
CREATE TABLE team_project_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id    TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (team_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX team_project_mappings_get_by_team_id ON team_project_mappings (team_id);
CREATE INDEX team_project_mappings_get_by_project_id ON team_project_mappings (project_id);
CREATE INDEX team_project_mappings_get_by_deleted ON team_project_mappings (deleted);
CREATE INDEX team_project_mappings_get_by_created ON team_project_mappings (created);
CREATE INDEX team_project_mappings_get_by_modified ON team_project_mappings (modified);
CREATE INDEX team_project_mappings_get_by_created_and_modified ON team_project_mappings (created, modified);

/*
     Repository belongs to project. Multiple repositories can belong to multiple projects.
     So, two projects can actually have the same repository.
*/
CREATE TABLE repository_project_mappings
(

    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    project_id    TEXT    NOT NULL,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (repository_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX repository_project_mappings_get_by_repository_id ON repository_project_mappings (repository_id);
CREATE INDEX repository_project_mappings_get_by_project_id ON repository_project_mappings (project_id);
CREATE INDEX repository_project_mappings_get_by_deleted ON repository_project_mappings (deleted);
CREATE INDEX repository_project_mappings_get_by_created ON repository_project_mappings (created);
CREATE INDEX repository_project_mappings_get_by_modified ON repository_project_mappings (modified);
CREATE INDEX repository_project_mappings_get_by_created_and_modified ON repository_project_mappings (created, modified);

/*
     Mapping all commits to the corresponding tickets
*/
CREATE TABLE repository_commit_ticket_mappings
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
    ON repository_commit_ticket_mappings (repository_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id ON repository_commit_ticket_mappings (ticket_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id
    ON repository_commit_ticket_mappings (repository_id, ticket_id);

CREATE INDEX repository_commit_ticket_mappings_get_by_commit_hash ON repository_commit_ticket_mappings (commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id_commit_hash
    ON repository_commit_ticket_mappings (ticket_id, commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id_commit_hash
    ON repository_commit_ticket_mappings (repository_id, ticket_id, commit_hash);

CREATE INDEX repository_commit_ticket_mappings_get_by_deleted ON repository_commit_ticket_mappings (deleted);
CREATE INDEX repository_commit_ticket_mappings_get_by_created ON repository_commit_ticket_mappings (created);
CREATE INDEX repository_commit_ticket_mappings_get_by_modified ON repository_commit_ticket_mappings (modified);

CREATE INDEX repository_commit_ticket_mappings_get_by_created_and_modified
    ON repository_commit_ticket_mappings (created, modified);

/*
    Components to the tickets mappings.
    Component can be mapped to the multiple tickets.
*/
CREATE TABLE component_ticket_mappings
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    ticket_id    TEXT    NOT NULL,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (component_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX component_ticket_mappings_get_by_ticket_id ON component_ticket_mappings (ticket_id);
CREATE INDEX component_ticket_mappings_get_by_component_id ON component_ticket_mappings (component_id);
CREATE INDEX component_ticket_mappings_get_by_deleted ON component_ticket_mappings (deleted);
CREATE INDEX component_ticket_mappings_get_by_created ON component_ticket_mappings (created);
CREATE INDEX component_ticket_mappings_get_by_modified ON component_ticket_mappings (modified);
CREATE INDEX component_ticket_mappings_get_by_created_and_modified ON component_ticket_mappings (created, modified);

/*
    Components meta-data:
    Associate the various information with different components.
*/
CREATE TABLE components_meta_data
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX components_meta_data_get_by_component_id ON components_meta_data (component_id);
CREATE INDEX components_meta_data_get_by_property ON components_meta_data (property);
CREATE INDEX components_meta_data_get_by_component_id_and_property ON components_meta_data (component_id, property);
CREATE INDEX components_meta_data_get_by_value ON components_meta_data (value);
CREATE INDEX components_meta_data_get_by_component_id_and_value ON components_meta_data (component_id, value);
CREATE INDEX components_meta_data_get_by_property_and_value ON components_meta_data (property, value);

CREATE INDEX components_meta_data_get_by_component_id_and_property_and_value
    ON components_meta_data (component_id, property, value);

CREATE INDEX components_meta_data_get_by_deleted ON components_meta_data (deleted);
CREATE INDEX components_meta_data_get_by_created ON components_meta_data (created);
CREATE INDEX components_meta_data_get_by_modified ON components_meta_data (modified);
CREATE INDEX components_meta_data_get_by_created_and_modified ON components_meta_data (created, modified);

/*
    Assets can belong to the multiple projects.
    One example of the image used in the context of the project is the project's avatar.
    Projects may have various other assets associated to itself.
    Various documentation for example.
*/
CREATE TABLE asset_project_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX asset_project_mappings_get_by_asset_id ON asset_project_mappings (asset_id);
CREATE INDEX asset_project_mappings_get_by_project_id ON asset_project_mappings (project_id);
CREATE INDEX asset_project_mappings_get_by_deleted ON asset_project_mappings (deleted);
CREATE INDEX asset_project_mappings_get_by_created ON asset_project_mappings (created);
CREATE INDEX asset_project_mappings_get_by_modified ON asset_project_mappings (modified);
CREATE INDEX asset_project_mappings_get_by_created_and_modified ON asset_project_mappings (created, modified);

/*
    Assets can belong to the multiple teams.
    The image used in the context of the team is the team's avatar, for example.
    Teams may have other additions associated to itself. Various documents for example,
*/
CREATE TABLE asset_team_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0,
    UNIQUE (asset_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX asset_team_mappings_get_by_asset_id ON asset_team_mappings (asset_id);
CREATE INDEX asset_team_mappings_get_by_team_id ON asset_team_mappings (team_id);
CREATE INDEX asset_team_mappings_get_by_deleted ON asset_team_mappings (deleted);
CREATE INDEX asset_team_mappings_get_by_created ON asset_team_mappings (created);
CREATE INDEX asset_team_mappings_get_by_modified ON asset_team_mappings (modified);
CREATE INDEX asset_team_mappings_get_by_created_and_modified ON asset_team_mappings (created, modified);

/*
    Assets (attachments for example) can belong to the multiple tickets.
*/
CREATE TABLE asset_ticket_mappings
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX asset_ticket_mappings_get_by_asset_id ON asset_ticket_mappings (asset_id);
CREATE INDEX asset_ticket_mappings_get_by_ticket_id ON asset_ticket_mappings (ticket_id);
CREATE INDEX asset_ticket_mappings_get_by_deleted ON asset_ticket_mappings (deleted);
CREATE INDEX asset_ticket_mappings_get_by_created ON asset_ticket_mappings (created);
CREATE INDEX asset_ticket_mappings_get_by_modified ON asset_ticket_mappings (modified);
CREATE INDEX asset_ticket_mappings_get_by_created_and_modified ON asset_ticket_mappings (created, modified);

/*
    Assets (attachments for example) can belong to the multiple comments.
*/
CREATE TABLE asset_comment_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    comment_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (asset_id, comment_id) ON CONFLICT ABORT
);

CREATE INDEX asset_comment_mappings_get_by_asset_id ON asset_comment_mappings (asset_id);
CREATE INDEX asset_comment_mappings_get_by_comment_id ON asset_comment_mappings (comment_id);
CREATE INDEX asset_comment_mappings_get_by_deleted ON asset_comment_mappings (deleted);
CREATE INDEX asset_comment_mappings_get_by_created ON asset_comment_mappings (created);
CREATE INDEX asset_comment_mappings_get_by_modified ON asset_comment_mappings (modified);
CREATE INDEX asset_comment_mappings_get_by_created_and_modified ON asset_comment_mappings (created, modified);

/*
    Labels can belong to the label category.
    One single asset can belong to multiple categories.
*/
CREATE TABLE label_label_category_mappings
(

    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id          TEXT    NOT NULL,
    label_category_id TEXT    NOT NULL,
    created           INTEGER NOT NULL,
    modified          INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, label_category_id) ON CONFLICT ABORT
);

CREATE INDEX label_label_category_mappings_get_by_label_id ON label_label_category_mappings (label_id);

CREATE INDEX label_label_category_mappings_get_by_label_category_id
    ON label_label_category_mappings (label_category_id);

CREATE INDEX label_label_category_mappings_get_by_deleted ON label_label_category_mappings (deleted);
CREATE INDEX label_label_category_mappings_get_by_created ON label_label_category_mappings (created);
CREATE INDEX label_label_category_mappings_get_by_modified ON label_label_category_mappings (modified);

CREATE INDEX label_label_category_mappings_get_by_created_and_modified
    ON label_label_category_mappings (created, modified);

/*
    Label can be associated with one or more projects.
*/
CREATE TABLE label_project_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX label_project_mappings_get_by_label_id ON label_project_mappings (label_id);
CREATE INDEX label_project_mappings_get_by_project_id ON label_project_mappings (project_id);
CREATE INDEX label_project_mappings_get_by_deleted ON label_project_mappings (deleted);
CREATE INDEX label_project_mappings_get_by_created ON label_project_mappings (created);
CREATE INDEX label_project_mappings_get_by_modified ON label_project_mappings (modified);
CREATE INDEX label_project_mappings_get_by_created_and_modified ON label_project_mappings (created, modified);

/*
    Label can be associated with one or more teams.
*/
CREATE TABLE label_team_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0,
    UNIQUE (label_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX label_team_mappings_get_by_label_id ON label_team_mappings (label_id);
CREATE INDEX label_team_mappings_get_by_team_id ON label_team_mappings (team_id);
CREATE INDEX label_team_mappings_get_by_deleted ON label_team_mappings (deleted);
CREATE INDEX label_team_mappings_get_by_created ON label_team_mappings (created);
CREATE INDEX label_team_mappings_get_by_modified ON label_team_mappings (modified);
CREATE INDEX label_team_mappings_get_by_created_and_modified ON label_team_mappings (created, modified);

/*
    Label can be associated with one or more tickets.
*/
CREATE TABLE label_ticket_mappings
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (label_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX label_ticket_mappings_get_by_label_id ON label_ticket_mappings (label_id);
CREATE INDEX label_ticket_mappings_get_by_team_id ON label_ticket_mappings (ticket_id);
CREATE INDEX label_ticket_mappings_get_by_deleted ON label_ticket_mappings (deleted);
CREATE INDEX label_ticket_mappings_get_by_created ON label_ticket_mappings (created);
CREATE INDEX label_ticket_mappings_get_by_modified ON label_ticket_mappings (modified);
CREATE INDEX label_ticket_mappings_get_by_created_and_modified ON label_ticket_mappings (created, modified);

/*
    Label can be associated with one or more assets.
*/
CREATE TABLE label_asset_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    asset_id TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0,
    UNIQUE (label_id, asset_id) ON CONFLICT ABORT
);

CREATE INDEX label_asset_mappings_get_by_label_id ON label_asset_mappings (label_id);
CREATE INDEX label_asset_mappings_get_by_team_id ON label_asset_mappings (asset_id);
CREATE INDEX label_asset_mappings_get_by_deleted ON label_asset_mappings (deleted);
CREATE INDEX label_asset_mappings_get_by_created ON label_asset_mappings (created);
CREATE INDEX label_asset_mappings_get_by_modified ON label_asset_mappings (modified);
CREATE INDEX label_asset_mappings_get_by_created_and_modified ON label_asset_mappings (created, modified);

/*
    Comments are usually associated with project tickets:
*/
CREATE TABLE comment_ticket_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id TEXT    NOT NULL,
    ticket_id  TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (comment_id, ticket_id) ON CONFLICT ABORT
);

CREATE INDEX comment_ticket_mappings_get_by_comment_id ON comment_ticket_mappings (comment_id);
CREATE INDEX comment_ticket_mappings_get_by_ticket_id ON comment_ticket_mappings (ticket_id);
CREATE INDEX comment_ticket_mappings_get_by_deleted ON comment_ticket_mappings (deleted);
CREATE INDEX comment_ticket_mappings_get_by_created ON comment_ticket_mappings (created);
CREATE INDEX comment_ticket_mappings_get_by_modified ON comment_ticket_mappings (modified);
CREATE INDEX comment_ticket_mappings_get_by_created_and_modified ON comment_ticket_mappings (created, modified);

/*
    Tickets belong to the project:
*/
CREATE TABLE ticket_project_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_project_mappings_get_by_project_id ON ticket_project_mappings (project_id);
CREATE INDEX ticket_project_mappings_get_by_ticket_id ON ticket_project_mappings (ticket_id);
CREATE INDEX ticket_project_mappings_get_by_deleted ON ticket_project_mappings (deleted);
CREATE INDEX ticket_project_mappings_get_by_created ON ticket_project_mappings (created);
CREATE INDEX ticket_project_mappings_get_by_modified ON ticket_project_mappings (modified);
CREATE INDEX ticket_project_mappings_get_by_created_and_modified ON ticket_project_mappings (created, modified);

/*
    Cycles belong to the projects.
    Cycle can belong to exactly one project.
*/
CREATE TABLE cycle_project_mappings
(

    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    cycle_id   TEXT    NOT NULL UNIQUE,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (cycle_id, project_id) ON CONFLICT ABORT
);

CREATE INDEX cycle_project_mappings_get_by_project_id ON cycle_project_mappings (project_id);
CREATE INDEX cycle_project_mappings_get_by_cycle_id ON cycle_project_mappings (cycle_id);
CREATE INDEX cycle_project_mappings_get_by_deleted ON cycle_project_mappings (deleted);
CREATE INDEX cycle_project_mappings_get_by_created ON cycle_project_mappings (created);
CREATE INDEX cycle_project_mappings_get_by_modified ON cycle_project_mappings (modified);
CREATE INDEX cycle_project_mappings_get_by_created_and_modified ON cycle_project_mappings (created, modified);

/*
    Tickets can belong to cycles:
*/
CREATE TABLE ticket_cycle_mappings
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    cycle_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, cycle_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_cycle_mappings_get_by_ticket_id ON ticket_cycle_mappings (ticket_id);
CREATE INDEX ticket_cycle_mappings_get_by_cycle_id ON ticket_cycle_mappings (cycle_id);
CREATE INDEX ticket_cycle_mappings_get_by_deleted ON ticket_cycle_mappings (deleted);
CREATE INDEX ticket_cycle_mappings_get_by_created ON ticket_cycle_mappings (created);
CREATE INDEX ticket_cycle_mappings_get_by_modified ON ticket_cycle_mappings (modified);
CREATE INDEX ticket_cycle_mappings_get_by_created_and_modified ON ticket_cycle_mappings (created, modified);

/*
    Tickets can belong to one or more boards:
*/
CREATE TABLE ticket_board_mappings
(

    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    board_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (ticket_id, board_id) ON CONFLICT ABORT
);

CREATE INDEX ticket_board_mappings_get_by_ticket_id ON ticket_board_mappings (ticket_id);
CREATE INDEX ticket_board_mappings_get_by_bord_id ON ticket_board_mappings (board_id);
CREATE INDEX ticket_board_mappings_get_by_deleted ON ticket_board_mappings (deleted);
CREATE INDEX ticket_board_mappings_get_by_created ON ticket_board_mappings (created);
CREATE INDEX ticket_board_mappings_get_by_modified ON ticket_board_mappings (modified);
CREATE INDEX ticket_board_mappings_get_by_created_and_modified ON ticket_board_mappings (created, modified);

/*
    OAuth2 mappings:
*/

/*
    Users can be Yandex OAuth2 account users:
*/
CREATE TABLE users_yandex_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX users_yandex_mappings_get_by_user_id ON users_yandex_mappings (user_id);
CREATE INDEX users_yandex_mappings_get_by_username ON users_yandex_mappings (username);
CREATE INDEX users_yandex_mappings_get_by_deleted ON users_yandex_mappings (deleted);
CREATE INDEX users_yandex_mappings_get_by_created ON users_yandex_mappings (created);
CREATE INDEX users_yandex_mappings_get_by_modified ON users_yandex_mappings (modified);
CREATE INDEX users_yandex_mappings_get_by_created_and_modified ON users_yandex_mappings (created, modified);

/*
    Users can be Google OAuth2 account users:
*/
CREATE TABLE users_google_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX users_google_mappings_get_by_user_id ON users_google_mappings (user_id);
CREATE INDEX users_google_mappings_get_by_username ON users_google_mappings (username);
CREATE INDEX users_google_mappings_get_by_deleted ON users_google_mappings (deleted);
CREATE INDEX users_google_mappings_get_by_created ON users_google_mappings (created);
CREATE INDEX users_google_mappings_get_by_modified ON users_google_mappings (modified);
CREATE INDEX users_google_mappings_get_by_created_and_modified ON users_google_mappings (created, modified);

/*
    User access rights:
*/

/*
    User belongs to organizations:
*/
CREATE TABLE user_organization_mappings
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)),
    UNIQUE (user_id, organization_id) ON CONFLICT ABORT
);

CREATE INDEX user_organization_mappings_get_by_user_id ON user_organization_mappings (user_id);
CREATE INDEX user_organization_mappings_get_by_organization_id ON user_organization_mappings (organization_id);
CREATE INDEX user_organization_mappings_get_by_deleted ON user_organization_mappings (deleted);
CREATE INDEX user_organization_mappings_get_by_created ON user_organization_mappings (created);
CREATE INDEX user_organization_mappings_get_by_modified ON user_organization_mappings (modified);
CREATE INDEX user_organization_mappings_get_by_created_and_modified ON user_organization_mappings (created, modified);

/*
    User belongs to the organization's teams:
*/
CREATE TABLE user_team_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0,
    UNIQUE (user_id, team_id) ON CONFLICT ABORT
);

CREATE INDEX user_team_mappings_get_by_user_id ON user_team_mappings (user_id);
CREATE INDEX user_team_mappings_get_by_team_id ON user_team_mappings (team_id);
CREATE INDEX user_team_mappings_get_by_deleted ON user_team_mappings (deleted);
CREATE INDEX user_team_mappings_get_by_created ON user_team_mappings (created);
CREATE INDEX user_team_mappings_get_by_modified ON user_team_mappings (modified);
CREATE INDEX user_team_mappings_get_by_created_and_modified ON user_team_mappings (created, modified);

/*
    User has the permissions.
    Each permission has be associated to the proper permission context.
*/
CREATE TABLE permission_user_mappings
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

CREATE INDEX permission_user_mappings_get_by_user_id ON permission_user_mappings (user_id);
CREATE INDEX permission_user_mappings_get_by_permission_id ON permission_user_mappings (permission_id);
CREATE INDEX permission_user_mappings_get_by_permission_context_id ON permission_user_mappings (permission_context_id);

CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_id
    ON permission_user_mappings (user_id, permission_id);

CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_context_id
    ON permission_user_mappings (user_id, permission_context_id);

CREATE INDEX permission_user_mappings_get_by_permission_id_and_permission_context_id
    ON permission_user_mappings (permission_id, permission_context_id);

CREATE INDEX permission_user_mappings_get_by_deleted ON permission_user_mappings (deleted);
CREATE INDEX permission_user_mappings_get_by_created ON permission_user_mappings (created);
CREATE INDEX permission_user_mappings_get_by_modified ON permission_user_mappings (modified);
CREATE INDEX permission_user_mappings_get_by_created_and_modified ON permission_user_mappings (created, modified);

/*
    Team has the permissions.
    Each team permission has be associated to the proper permission context.
    All team members (users) will inherit team's permissions.
*/
CREATE TABLE permission_team_mappings
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

CREATE INDEX permission_team_mappings_get_by_team_id ON permission_team_mappings (team_id);
CREATE INDEX permission_team_mappings_get_by_permission_id ON permission_team_mappings (permission_id);

CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_id
    ON permission_team_mappings (team_id, permission_id);

CREATE INDEX permission_team_mappings_get_by_permission_context_id ON permission_team_mappings (permission_context_id);

CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_context_id
    ON permission_team_mappings (team_id, permission_context_id);

CREATE INDEX permission_team_mappings_get_by_deleted ON permission_team_mappings (deleted);
CREATE INDEX permission_team_mappings_get_by_created ON permission_team_mappings (created);
CREATE INDEX permission_team_mappings_get_by_modified ON permission_team_mappings (modified);
CREATE INDEX permission_team_mappings_get_by_created_and_modified ON permission_team_mappings (created, modified);

/*
    The configuration data for the extension.
    Basically it represents the meta-data associated with each extension.
    Each configuration property can be enabled or disabled.
*/
CREATE TABLE configuration_data_extension_mappings
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
    ON configuration_data_extension_mappings (extension_id);

CREATE INDEX configuration_data_extension_mappings_get_by_property ON configuration_data_extension_mappings (property);
CREATE INDEX configuration_data_extension_mappings_get_by_value ON configuration_data_extension_mappings (value);

CREATE INDEX configuration_data_extension_mappings_get_by_property_and_value
    ON configuration_data_extension_mappings (property, value);

CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property
    ON configuration_data_extension_mappings (extension_id, property);

CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property_and_value
    ON configuration_data_extension_mappings (extension_id, property, value);

CREATE INDEX configuration_data_extension_mappings_get_by_enabled ON configuration_data_extension_mappings (enabled);
CREATE INDEX configuration_data_extension_mappings_get_by_deleted ON configuration_data_extension_mappings (deleted);
CREATE INDEX configuration_data_extension_mappings_get_by_created ON configuration_data_extension_mappings (created);
CREATE INDEX configuration_data_extension_mappings_get_by_modified ON configuration_data_extension_mappings (modified);

CREATE INDEX configuration_data_extension_mappings_get_by_created_and_modified
    ON configuration_data_extension_mappings (created, modified);

/*
    Extensions meta-data:
    Associate the various information with different extension.
    Meta-data information are the extension specific.
*/
CREATE TABLE extensions_meta_data
(

    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX extensions_meta_data_get_by_extension_id ON extensions_meta_data (extension_id);
CREATE INDEX extensions_meta_data_get_by_property ON extensions_meta_data (property);
CREATE INDEX extensions_meta_data_get_by_value ON extensions_meta_data (value);
CREATE INDEX extensions_meta_data_get_by_property_and_value ON extensions_meta_data (property, value);

CREATE INDEX extensions_meta_data_get_by_extension_id_and_property_and_value
    ON extensions_meta_data (extension_id, property, value);

CREATE INDEX extensions_meta_data_get_by_extension_id_and_property ON extensions_meta_data (extension_id, property);
CREATE INDEX extensions_meta_data_get_by_deleted ON extensions_meta_data (deleted);
CREATE INDEX extensions_meta_data_get_by_created ON extensions_meta_data (created);
CREATE INDEX extensions_meta_data_get_by_modified ON extensions_meta_data (modified);
CREATE INDEX extensions_meta_data_get_by_created_and_modified ON extensions_meta_data (created, modified);
