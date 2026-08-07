-- Coze Studio PostgreSQL Schema
-- Converted from MySQL to PostgreSQL

-- Create database (run separately with superuser)
-- CREATE DATABASE opencoze WITH ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8';

-- Set client encoding
SET client_encoding = 'UTF8';

-- Create 'agent_to_database' table
CREATE TABLE IF NOT EXISTS agent_to_database (
    id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL,
    database_id BIGINT NOT NULL,
    is_draft BOOLEAN NOT NULL,
    prompt_disable BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_agent_db_draft ON agent_to_database (agent_id, database_id, is_draft);
COMMENT ON TABLE agent_to_database IS 'agent_to_database info';
COMMENT ON COLUMN agent_to_database.id IS 'ID';
COMMENT ON COLUMN agent_to_database.agent_id IS 'Agent ID';
COMMENT ON COLUMN agent_to_database.database_id IS 'ID of database_info';
COMMENT ON COLUMN agent_to_database.is_draft IS 'Is draft';
COMMENT ON COLUMN agent_to_database.prompt_disable IS 'Support prompt calls: TRUE not supported, FALSE supported';

-- Create 'agent_tool_draft' table
CREATE TABLE IF NOT EXISTS agent_tool_draft (
    id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    tool_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    sub_url VARCHAR(512) NOT NULL DEFAULT '',
    method VARCHAR(64) NOT NULL DEFAULT '',
    tool_name VARCHAR(255) NOT NULL DEFAULT '',
    tool_version VARCHAR(255) NOT NULL DEFAULT '',
    operation JSONB,
    source SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_agent_plugin_tool ON agent_tool_draft (agent_id, plugin_id, tool_id);
CREATE INDEX IF NOT EXISTS idx_agent_tool_bind ON agent_tool_draft (agent_id, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_agent_tool_id ON agent_tool_draft (agent_id, tool_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_agent_tool_name ON agent_tool_draft (agent_id, tool_name);
COMMENT ON TABLE agent_tool_draft IS 'Draft Agent Tool';

-- Create 'agent_tool_version' table
CREATE TABLE IF NOT EXISTS agent_tool_version (
    id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    tool_id BIGINT NOT NULL DEFAULT 0,
    agent_version VARCHAR(255) NOT NULL DEFAULT '',
    tool_name VARCHAR(255) NOT NULL DEFAULT '',
    tool_version VARCHAR(255) NOT NULL DEFAULT '',
    sub_url VARCHAR(512) NOT NULL DEFAULT '',
    method VARCHAR(64) NOT NULL DEFAULT '',
    operation JSONB,
    created_at BIGINT NOT NULL DEFAULT 0,
    source SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_agent_tool_id_created_at ON agent_tool_version (agent_id, tool_id, created_at);
CREATE INDEX IF NOT EXISTS idx_agent_tool_name_created_at ON agent_tool_version (agent_id, tool_name, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_agent_tool_id_agent_version ON agent_tool_version (agent_id, tool_id, agent_version);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_agent_tool_name_agent_version ON agent_tool_version (agent_id, tool_name, agent_version);
COMMENT ON TABLE agent_tool_version IS 'Agent Tool Version';

-- Create 'api_key' table
CREATE TABLE IF NOT EXISTS api_key (
    id BIGSERIAL,
    api_key VARCHAR(255) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL DEFAULT 0,
    expired_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    last_used_at BIGINT NOT NULL DEFAULT 0,
    ak_type SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
COMMENT ON TABLE api_key IS 'api key table';

-- Create 'app_connector_release_ref' table
CREATE TABLE IF NOT EXISTS app_connector_release_ref (
    id BIGINT NOT NULL DEFAULT 0,
    record_id BIGINT NOT NULL DEFAULT 0,
    connector_id BIGINT,
    publish_config JSONB,
    publish_status SMALLINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_record_connector ON app_connector_release_ref (record_id, connector_id);
COMMENT ON TABLE app_connector_release_ref IS 'Connector Release Record Reference';

-- Create 'app_conversation_template_draft' table
CREATE TABLE IF NOT EXISTS app_conversation_template_draft (
    id BIGINT NOT NULL,
    app_id BIGINT NOT NULL,
    space_id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    template_id BIGINT NOT NULL,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_space_id_app_id_template_id ON app_conversation_template_draft (space_id, app_id, template_id);
COMMENT ON TABLE app_conversation_template_draft IS 'app_conversation_template_draft';

-- Create 'app_conversation_template_online' table
CREATE TABLE IF NOT EXISTS app_conversation_template_online (
    id BIGINT NOT NULL,
    app_id BIGINT NOT NULL,
    space_id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    template_id BIGINT NOT NULL,
    version VARCHAR(256) NOT NULL,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_space_id_app_id_template_id_version ON app_conversation_template_online (space_id, app_id, template_id, version);
COMMENT ON TABLE app_conversation_template_online IS 'app_conversation_template_online';

-- Create 'app_draft' table
CREATE TABLE IF NOT EXISTS app_draft (
    id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    owner_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
COMMENT ON TABLE app_draft IS 'Draft Application';

-- Create 'app_dynamic_conversation_draft' table
CREATE TABLE IF NOT EXISTS app_dynamic_conversation_draft (
    id BIGINT NOT NULL,
    app_id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    user_id BIGINT NOT NULL,
    connector_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_id_connector_id_user_id ON app_dynamic_conversation_draft (app_id, connector_id, user_id);
CREATE INDEX IF NOT EXISTS idx_connector_id_user_id_name ON app_dynamic_conversation_draft (connector_id, user_id, name);
COMMENT ON TABLE app_dynamic_conversation_draft IS 'app_dynamic_conversation_draft';

-- Create 'app_dynamic_conversation_online' table
CREATE TABLE IF NOT EXISTS app_dynamic_conversation_online (
    id BIGINT NOT NULL,
    app_id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    user_id BIGINT NOT NULL,
    connector_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_id_connector_id_user_id ON app_dynamic_conversation_online (app_id, connector_id, user_id);
CREATE INDEX IF NOT EXISTS idx_connector_id_user_id_name ON app_dynamic_conversation_online (connector_id, user_id, name);
COMMENT ON TABLE app_dynamic_conversation_online IS 'app_dynamic_conversation_online';

-- Create 'app_release_record' table
CREATE TABLE IF NOT EXISTS app_release_record (
    id BIGINT NOT NULL DEFAULT 0,
    app_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    owner_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    connector_ids JSONB,
    extra_info JSONB,
    version VARCHAR(255) NOT NULL DEFAULT '',
    version_desc TEXT,
    publish_status SMALLINT NOT NULL DEFAULT 0,
    publish_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_publish_at ON app_release_record (app_id, publish_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_app_version_connector ON app_release_record (app_id, version);
COMMENT ON TABLE app_release_record IS 'Application Release Record';

-- Create 'app_static_conversation_draft' table
CREATE TABLE IF NOT EXISTS app_static_conversation_draft (
    id BIGINT NOT NULL,
    template_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    connector_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_id_user_id_template_id ON app_static_conversation_draft (connector_id, user_id, template_id);
COMMENT ON TABLE app_static_conversation_draft IS 'app_static_conversation_draft';

-- Create 'app_static_conversation_online' table
CREATE TABLE IF NOT EXISTS app_static_conversation_online (
    id BIGINT NOT NULL,
    template_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    connector_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_id_user_id_template_id ON app_static_conversation_online (connector_id, user_id, template_id);
COMMENT ON TABLE app_static_conversation_online IS 'app_static_conversation_online';

-- Create 'chat_flow_role_config' table
CREATE TABLE IF NOT EXISTS chat_flow_role_config (
    id BIGINT NOT NULL,
    workflow_id BIGINT NOT NULL,
    connector_id BIGINT,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    version VARCHAR(256),
    avatar VARCHAR(256) NOT NULL,
    background_image_info TEXT,
    onboarding_info TEXT,
    suggest_reply_info TEXT,
    audio_config TEXT,
    user_input_config VARCHAR(256) NOT NULL,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_id_version ON chat_flow_role_config (connector_id, version);
CREATE INDEX IF NOT EXISTS idx_workflow_id_version ON chat_flow_role_config (workflow_id, version);
COMMENT ON TABLE chat_flow_role_config IS 'chat_flow_role_config';

-- Create 'connector_workflow_version' table
CREATE TABLE IF NOT EXISTS connector_workflow_version (
    id BIGSERIAL,
    app_id BIGINT NOT NULL,
    connector_id BIGINT NOT NULL,
    workflow_id BIGINT NOT NULL,
    version VARCHAR(256) NOT NULL,
    created_at BIGINT NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_id_workflow_id_create_at ON connector_workflow_version (connector_id, workflow_id, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_connector_id_workflow_id_version ON connector_workflow_version (connector_id, workflow_id, version);
COMMENT ON TABLE connector_workflow_version IS 'connector workflow version';

-- Create 'conversation' table
CREATE TABLE IF NOT EXISTS conversation (
    id BIGSERIAL,
    name VARCHAR(255) DEFAULT '',
    connector_id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    scene SMALLINT NOT NULL DEFAULT 0,
    section_id BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT DEFAULT 0,
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    ext TEXT,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_bot_status ON conversation (connector_id, agent_id, creator_id);
COMMENT ON TABLE conversation IS 'conversation info record';

-- Create 'data_copy_task' table
CREATE TABLE IF NOT EXISTS data_copy_task (
    master_task_id VARCHAR(128) DEFAULT '',
    origin_data_id BIGINT NOT NULL DEFAULT 0,
    target_data_id BIGINT NOT NULL DEFAULT 0,
    origin_space_id BIGINT NOT NULL DEFAULT 0,
    target_space_id BIGINT NOT NULL DEFAULT 0,
    origin_user_id BIGINT NOT NULL DEFAULT 0,
    target_user_id BIGINT DEFAULT 0,
    origin_app_id BIGINT NOT NULL DEFAULT 0,
    target_app_id BIGINT NOT NULL DEFAULT 0,
    data_type SMALLINT NOT NULL DEFAULT 0,
    ext_info VARCHAR(255) NOT NULL DEFAULT '',
    start_time BIGINT DEFAULT 0,
    finish_time BIGINT,
    status SMALLINT NOT NULL DEFAULT 1,
    error_msg VARCHAR(128),
    id BIGSERIAL,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_master_task_id_origin_data_id_data_type ON data_copy_task (master_task_id, origin_data_id, data_type);
COMMENT ON TABLE data_copy_task IS 'data copy task record';

-- Create 'draft_database_info' table
CREATE TABLE IF NOT EXISTS draft_database_info (
    id BIGINT NOT NULL,
    app_id BIGINT,
    space_id BIGINT NOT NULL,
    related_online_id BIGINT NOT NULL,
    is_visible SMALLINT NOT NULL DEFAULT 1,
    prompt_disabled SMALLINT NOT NULL DEFAULT 0,
    table_name VARCHAR(255) NOT NULL,
    table_desc VARCHAR(256),
    table_field TEXT,
    creator_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(255) NOT NULL,
    physical_table_name VARCHAR(255),
    rw_mode BIGINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_space_app_creator_deleted ON draft_database_info (space_id, app_id, creator_id, deleted_at);
COMMENT ON TABLE draft_database_info IS 'draft database info';

-- Create 'files' table
CREATE TABLE IF NOT EXISTS files (
    id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    tos_uri VARCHAR(1024) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0,
    comment VARCHAR(1024) NOT NULL DEFAULT '',
    source SMALLINT NOT NULL DEFAULT 0,
    creator_id VARCHAR(512) NOT NULL DEFAULT '',
    content_type VARCHAR(255) NOT NULL DEFAULT '',
    coze_account_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON files (creator_id);
COMMENT ON TABLE files IS 'file resource table';

-- Create 'knowledge' table
CREATE TABLE IF NOT EXISTS knowledge (
    id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL DEFAULT '',
    app_id BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    status SMALLINT NOT NULL DEFAULT 1,
    description TEXT,
    icon_uri VARCHAR(150),
    format_type SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_id ON knowledge (app_id);
CREATE INDEX IF NOT EXISTS idx_creator_id ON knowledge (creator_id);
CREATE INDEX IF NOT EXISTS idx_space_id_deleted_at_updated_at ON knowledge (space_id, deleted_at, updated_at);
COMMENT ON TABLE knowledge IS 'knowledge table';

-- Create 'knowledge_document' table
CREATE TABLE IF NOT EXISTS knowledge_document (
    id BIGINT NOT NULL,
    knowledge_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(150) NOT NULL DEFAULT '',
    file_extension VARCHAR(20) NOT NULL DEFAULT '0',
    document_type INT NOT NULL DEFAULT 0,
    uri TEXT,
    size BIGINT NOT NULL DEFAULT 0,
    slice_count BIGINT NOT NULL DEFAULT 0,
    char_count BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    source_type INT DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    fail_reason TEXT,
    parse_rule JSONB,
    table_info JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON knowledge_document (creator_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_id_deleted_at_updated_at ON knowledge_document (knowledge_id, deleted_at, updated_at);
COMMENT ON TABLE knowledge_document IS 'knowledge document info';

-- Create 'knowledge_document_review' table
CREATE TABLE IF NOT EXISTS knowledge_document_review (
    id BIGINT NOT NULL DEFAULT 0,
    knowledge_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(150) NOT NULL DEFAULT '',
    type VARCHAR(10) NOT NULL DEFAULT '0',
    uri TEXT,
    format_type SMALLINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    chunk_resp_uri TEXT,
    deleted_at TIMESTAMPTZ(3),
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_dataset_id ON knowledge_document_review (knowledge_id, status, updated_at);
COMMENT ON TABLE knowledge_document_review IS 'Document slice preview info';

-- Create 'knowledge_document_slice' table
CREATE TABLE IF NOT EXISTS knowledge_document_slice (
    id BIGINT NOT NULL DEFAULT 0,
    knowledge_id BIGINT NOT NULL DEFAULT 0,
    document_id BIGINT NOT NULL DEFAULT 0,
    content TEXT,
    sequence DECIMAL(20,5) NOT NULL,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    creator_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    fail_reason TEXT,
    hit BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_document_id_deleted_at_sequence ON knowledge_document_slice (document_id, deleted_at, sequence);
CREATE INDEX IF NOT EXISTS idx_knowledge_id_document_id ON knowledge_document_slice (knowledge_id, document_id);
CREATE INDEX IF NOT EXISTS idx_sequence ON knowledge_document_slice (sequence);
COMMENT ON TABLE knowledge_document_slice IS 'knowledge document slice';

-- Create 'kv_entries' table
CREATE TABLE IF NOT EXISTS kv_entries (
    id BIGSERIAL,
    namespace VARCHAR(255) NOT NULL,
    key_data VARCHAR(255) NOT NULL,
    value_data BYTEA,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_namespace_key ON kv_entries (namespace, key_data);
COMMENT ON TABLE kv_entries IS 'kv data';

-- Create 'message' table
CREATE TABLE IF NOT EXISTS message (
    id BIGSERIAL,
    run_id BIGINT NOT NULL DEFAULT 0,
    conversation_id BIGINT NOT NULL DEFAULT 0,
    user_id VARCHAR(60) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    role VARCHAR(100) NOT NULL DEFAULT '',
    content_type VARCHAR(100) NOT NULL DEFAULT '',
    content TEXT,
    message_type VARCHAR(100) NOT NULL DEFAULT '',
    display_content TEXT,
    ext TEXT,
    section_id BIGINT,
    broken_position INT DEFAULT -1,
    status SMALLINT NOT NULL DEFAULT 0,
    model_content TEXT,
    meta_info TEXT,
    reasoning_content TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_conversation_id ON message (conversation_id);
CREATE INDEX IF NOT EXISTS idx_run_id ON message (run_id);
COMMENT ON TABLE message IS 'message record';

-- Create 'model_entity' table
CREATE TABLE IF NOT EXISTS model_entity (
    id BIGINT NOT NULL,
    meta_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    default_params JSONB,
    scenario BIGINT NOT NULL,
    status INT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_scenario ON model_entity (scenario);
CREATE INDEX IF NOT EXISTS idx_status ON model_entity (status);
COMMENT ON TABLE model_entity IS 'Model information';

-- Create 'model_instance' table
CREATE TABLE IF NOT EXISTS model_instance (
    id BIGSERIAL,
    type SMALLINT NOT NULL,
    provider JSONB NOT NULL,
    display_info JSONB NOT NULL,
    connection JSONB NOT NULL,
    capability JSONB NOT NULL,
    parameters JSONB NOT NULL,
    extra JSONB,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE model_instance IS 'Model Instance Management Table';

-- Create 'model_meta' table
CREATE TABLE IF NOT EXISTS model_meta (
    id BIGINT NOT NULL,
    model_name VARCHAR(128) NOT NULL,
    protocol VARCHAR(128) NOT NULL,
    icon_uri VARCHAR(255) NOT NULL DEFAULT '',
    capability JSONB,
    conn_config JSONB,
    status INT NOT NULL DEFAULT 1,
    description VARCHAR(2048) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT,
    icon_url VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_status ON model_meta (status);
COMMENT ON TABLE model_meta IS 'Model metadata';

-- Create 'node_execution' table
CREATE TABLE IF NOT EXISTS node_execution (
    id BIGINT NOT NULL,
    execute_id BIGINT NOT NULL,
    node_id VARCHAR(128) NOT NULL,
    node_name VARCHAR(128) NOT NULL,
    node_type VARCHAR(128) NOT NULL,
    created_at BIGINT NOT NULL,
    status SMALLINT NOT NULL,
    duration BIGINT,
    input TEXT,
    output TEXT,
    raw_output TEXT,
    error_info TEXT,
    error_level VARCHAR(32),
    input_tokens BIGINT,
    output_tokens BIGINT,
    updated_at BIGINT,
    composite_node_index BIGINT,
    composite_node_items TEXT,
    parent_node_id VARCHAR(128),
    sub_execute_id BIGINT,
    extra TEXT,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_execute_id_node_id ON node_execution (execute_id, node_id);
CREATE INDEX IF NOT EXISTS idx_execute_id_parent_node_id ON node_execution (execute_id, parent_node_id);
COMMENT ON TABLE node_execution IS 'Node run record';

-- Create 'online_database_info' table
CREATE TABLE IF NOT EXISTS online_database_info (
    id BIGINT NOT NULL,
    app_id BIGINT,
    space_id BIGINT NOT NULL,
    related_draft_id BIGINT NOT NULL,
    is_visible SMALLINT NOT NULL DEFAULT 1,
    prompt_disabled SMALLINT NOT NULL DEFAULT 0,
    table_name VARCHAR(255) NOT NULL,
    table_desc VARCHAR(256),
    table_field TEXT,
    creator_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(255) NOT NULL,
    physical_table_name VARCHAR(255),
    rw_mode BIGINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_space_app_creator_deleted ON online_database_info (space_id, app_id, creator_id, deleted_at);
COMMENT ON TABLE online_database_info IS 'online database info';

-- Create 'plugin' table
CREATE TABLE IF NOT EXISTS plugin (
    id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    developer_id BIGINT NOT NULL DEFAULT 0,
    app_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    server_url VARCHAR(512) NOT NULL DEFAULT '',
    plugin_type SMALLINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    version VARCHAR(255) NOT NULL DEFAULT '',
    version_desc TEXT,
    manifest JSONB,
    openapi_doc JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_space_created_at ON plugin (space_id, created_at);
CREATE INDEX IF NOT EXISTS idx_space_updated_at ON plugin (space_id, updated_at);
COMMENT ON TABLE plugin IS 'Latest Plugin';

-- Create 'plugin_draft' table
CREATE TABLE IF NOT EXISTS plugin_draft (
    id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    developer_id BIGINT NOT NULL DEFAULT 0,
    app_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    server_url VARCHAR(512) NOT NULL DEFAULT '',
    plugin_type SMALLINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    manifest JSONB,
    openapi_doc JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_id ON plugin_draft (app_id, id);
CREATE INDEX IF NOT EXISTS idx_space_app_created_at ON plugin_draft (space_id, app_id, created_at);
CREATE INDEX IF NOT EXISTS idx_space_app_updated_at ON plugin_draft (space_id, app_id, updated_at);
COMMENT ON TABLE plugin_draft IS 'Draft Plugin';

-- Create 'plugin_oauth_auth' table
CREATE TABLE IF NOT EXISTS plugin_oauth_auth (
    id BIGINT NOT NULL DEFAULT 0,
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    plugin_id BIGINT NOT NULL DEFAULT 0,
    is_draft BOOLEAN NOT NULL DEFAULT FALSE,
    oauth_config JSONB,
    access_token TEXT,
    refresh_token TEXT,
    token_expired_at BIGINT,
    next_token_refresh_at BIGINT,
    last_active_at BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_last_active_at ON plugin_oauth_auth (last_active_at);
CREATE INDEX IF NOT EXISTS idx_last_token_expired_at ON plugin_oauth_auth (token_expired_at);
CREATE INDEX IF NOT EXISTS idx_next_token_refresh_at ON plugin_oauth_auth (next_token_refresh_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_user_plugin_is_draft ON plugin_oauth_auth (user_id, plugin_id, is_draft);
COMMENT ON TABLE plugin_oauth_auth IS 'Plugin OAuth Authorization Code Info';

-- Create 'plugin_version' table
CREATE TABLE IF NOT EXISTS plugin_version (
    id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    developer_id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    app_id BIGINT NOT NULL DEFAULT 0,
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    server_url VARCHAR(512) NOT NULL DEFAULT '',
    plugin_type SMALLINT NOT NULL DEFAULT 0,
    version VARCHAR(255) NOT NULL DEFAULT '',
    version_desc TEXT,
    manifest JSONB,
    openapi_doc JSONB,
    created_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_plugin_version ON plugin_version (plugin_id, version);
COMMENT ON TABLE plugin_version IS 'Plugin Version';

-- Create 'prompt_resource' table
CREATE TABLE IF NOT EXISTS prompt_resource (
    id BIGSERIAL,
    space_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    prompt_text TEXT,
    status INT NOT NULL,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON prompt_resource (creator_id);
COMMENT ON TABLE prompt_resource IS 'prompt_resource';

-- Create 'run_record' table
CREATE TABLE IF NOT EXISTS run_record (
    id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL DEFAULT 0,
    section_id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    source SMALLINT NOT NULL DEFAULT 0,
    status VARCHAR(255) NOT NULL DEFAULT '',
    creator_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    failed_at BIGINT NOT NULL DEFAULT 0,
    last_error TEXT,
    completed_at BIGINT NOT NULL DEFAULT 0,
    chat_request TEXT,
    ext TEXT,
    usage JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_c_s ON run_record (conversation_id, section_id);
COMMENT ON TABLE run_record IS 'run record';

-- Create 'shortcut_command' table
CREATE TABLE IF NOT EXISTS shortcut_command (
    id BIGSERIAL,
    object_id BIGINT NOT NULL DEFAULT 0,
    command_id BIGINT NOT NULL DEFAULT 0,
    command_name VARCHAR(255) NOT NULL DEFAULT '',
    shortcut_command VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    send_type SMALLINT NOT NULL DEFAULT 0,
    tool_type SMALLINT NOT NULL DEFAULT 0,
    work_flow_id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    plugin_tool_name VARCHAR(255) NOT NULL DEFAULT '',
    template_query TEXT,
    components JSONB,
    card_schema TEXT,
    tool_info JSONB,
    status SMALLINT NOT NULL DEFAULT 0,
    creator_id BIGINT DEFAULT 0,
    is_online SMALLINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    shortcut_icon JSONB,
    plugin_tool_id BIGINT NOT NULL DEFAULT 0,
    source SMALLINT DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_object_command_id_type ON shortcut_command (object_id, command_id, is_online);
COMMENT ON TABLE shortcut_command IS 'shortcut_command';

-- Create 'single_agent_draft' table
CREATE TABLE IF NOT EXISTS single_agent_draft (
    id BIGSERIAL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    icon_uri VARCHAR(255) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    variables_meta_id BIGINT,
    model_info JSONB,
    onboarding_info JSONB,
    prompt JSONB,
    plugin JSONB,
    knowledge JSONB,
    workflow JSONB,
    suggest_reply JSONB,
    jump_config JSONB,
    background_image_info_list JSONB,
    database_config JSONB,
    bot_mode SMALLINT NOT NULL DEFAULT 0,
    layout_info TEXT,
    shortcut_command JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON single_agent_draft (creator_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_agent_id ON single_agent_draft (agent_id);
COMMENT ON TABLE single_agent_draft IS 'Single Agent Draft Copy Table';

-- Create 'single_agent_publish' table
CREATE TABLE IF NOT EXISTS single_agent_publish (
    id BIGSERIAL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    publish_id VARCHAR(50) NOT NULL DEFAULT '',
    connector_ids JSONB,
    version VARCHAR(255) NOT NULL DEFAULT '',
    publish_info TEXT,
    publish_time BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    extra JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_agent_id_version ON single_agent_publish (agent_id, version);
CREATE INDEX IF NOT EXISTS idx_creator_id ON single_agent_publish (creator_id);
CREATE INDEX IF NOT EXISTS idx_publish_id ON single_agent_publish (publish_id);
COMMENT ON TABLE single_agent_publish IS 'Bot connector and release version info';

-- Create 'single_agent_version' table
CREATE TABLE IF NOT EXISTS single_agent_version (
    id BIGSERIAL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    icon_uri VARCHAR(255) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT 0,
    bot_mode SMALLINT NOT NULL DEFAULT 0,
    layout_info TEXT,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ(3),
    variables_meta_id BIGINT,
    model_info JSONB,
    onboarding_info JSONB,
    prompt JSONB,
    plugin JSONB,
    knowledge JSONB,
    workflow JSONB,
    suggest_reply JSONB,
    jump_config JSONB,
    connector_id BIGINT NOT NULL,
    version VARCHAR(255) NOT NULL DEFAULT '',
    background_image_info_list JSONB,
    database_config JSONB,
    shortcut_command JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON single_agent_version (creator_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_agent_id_and_version_connector_id ON single_agent_version (agent_id, version, connector_id);
COMMENT ON TABLE single_agent_version IS 'Single Agent Version Copy Table';

-- Create 'space' table
CREATE TABLE IF NOT EXISTS space (
    id BIGSERIAL,
    owner_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(200) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    icon_uri VARCHAR(200) NOT NULL DEFAULT '',
    creator_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_creator_id ON space (creator_id);
CREATE INDEX IF NOT EXISTS idx_owner_id ON space (owner_id);
COMMENT ON TABLE space IS 'Space Table';

-- Create 'space_user' table
CREATE TABLE IF NOT EXISTS space_user (
    id BIGSERIAL,
    space_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL DEFAULT 0,
    role_type INT NOT NULL DEFAULT 3,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_user_id ON space_user (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_space_user ON space_user (space_id, user_id);
COMMENT ON TABLE space_user IS 'Space Member Table';

-- Create 'template' table
CREATE TABLE IF NOT EXISTS template (
    id BIGSERIAL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    workflow_id BIGINT NOT NULL DEFAULT 0,
    space_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    heat BIGINT NOT NULL DEFAULT 0,
    product_entity_type BIGINT NOT NULL DEFAULT 0,
    meta_info JSONB,
    agent_extra JSONB,
    workflow_extra JSONB,
    project_extra JSONB,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_agent_id ON template (agent_id);
COMMENT ON TABLE template IS 'Template Info Table';

-- Create 'tool' table
CREATE TABLE IF NOT EXISTS tool (
    id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    version VARCHAR(255) NOT NULL DEFAULT '',
    sub_url VARCHAR(512) NOT NULL DEFAULT '',
    method VARCHAR(64) NOT NULL DEFAULT '',
    operation JSONB,
    activated_status SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_plugin_activated_status ON tool (plugin_id, activated_status);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_plugin_sub_url_method ON tool (plugin_id, sub_url, method);
COMMENT ON TABLE tool IS 'Latest Tool';

-- Create 'tool_draft' table
CREATE TABLE IF NOT EXISTS tool_draft (
    id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    sub_url VARCHAR(512) NOT NULL DEFAULT '',
    method VARCHAR(64) NOT NULL DEFAULT '',
    operation JSONB,
    debug_status SMALLINT NOT NULL DEFAULT 0,
    activated_status SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_plugin_created_at_id ON tool_draft (plugin_id, created_at, id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_plugin_sub_url_method ON tool_draft (plugin_id, sub_url, method);
COMMENT ON TABLE tool_draft IS 'Draft Tool';

-- Create 'tool_version' table
CREATE TABLE IF NOT EXISTS tool_version (
    id BIGINT NOT NULL DEFAULT 0,
    tool_id BIGINT NOT NULL DEFAULT 0,
    plugin_id BIGINT NOT NULL DEFAULT 0,
    version VARCHAR(255) NOT NULL DEFAULT '',
    sub_url VARCHAR(512) NOT NULL DEFAULT '',
    method VARCHAR(64) NOT NULL DEFAULT '',
    operation JSONB,
    created_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_tool_version ON tool_version (tool_id, version);
COMMENT ON TABLE tool_version IS 'Tool Version';

-- Create 'user' table
CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL,
    name VARCHAR(128) NOT NULL DEFAULT '',
    unique_name VARCHAR(128) NOT NULL DEFAULT '',
    email VARCHAR(128) NOT NULL DEFAULT '',
    password VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(512) NOT NULL DEFAULT '',
    icon_uri VARCHAR(512) NOT NULL DEFAULT '',
    user_verified BOOLEAN NOT NULL DEFAULT FALSE,
    locale VARCHAR(128) NOT NULL DEFAULT '',
    session_key VARCHAR(256) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_session_key ON "user" (session_key);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_email ON "user" (email);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_unique_name ON "user" (unique_name);
COMMENT ON TABLE "user" IS 'User Table';

-- Create 'variable_instance' table
CREATE TABLE IF NOT EXISTS variable_instance (
    id BIGINT NOT NULL DEFAULT 0,
    biz_type SMALLINT NOT NULL,
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    version VARCHAR(255) NOT NULL,
    keyword VARCHAR(255) NOT NULL,
    type SMALLINT NOT NULL,
    content TEXT,
    connector_uid VARCHAR(255) NOT NULL,
    connector_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_connector_key ON variable_instance (biz_id, biz_type, version, connector_uid, connector_id);
COMMENT ON TABLE variable_instance IS 'KV Memory';

-- Create 'variables_meta' table
CREATE TABLE IF NOT EXISTS variables_meta (
    id BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL,
    biz_type SMALLINT NOT NULL,
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    variable_list JSONB,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    version VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_user_key ON variables_meta (creator_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_project_key ON variables_meta (biz_id, biz_type, version);
COMMENT ON TABLE variables_meta IS 'KV Memory meta';

-- Create 'workflow_draft' table
CREATE TABLE IF NOT EXISTS workflow_draft (
    id BIGINT NOT NULL,
    canvas TEXT,
    input_params TEXT,
    output_params TEXT,
    test_run_success BOOLEAN NOT NULL DEFAULT FALSE,
    modified BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at BIGINT,
    deleted_at TIMESTAMPTZ(3),
    commit_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_updated_at ON workflow_draft (updated_at DESC);
COMMENT ON TABLE workflow_draft IS 'Workflow canvas draft table';

-- Create 'workflow_execution' table
CREATE TABLE IF NOT EXISTS workflow_execution (
    id BIGINT NOT NULL,
    workflow_id BIGINT NOT NULL,
    version VARCHAR(50),
    space_id BIGINT NOT NULL,
    mode SMALLINT NOT NULL,
    operator_id BIGINT NOT NULL,
    connector_id BIGINT,
    connector_uid VARCHAR(64),
    created_at BIGINT NOT NULL,
    log_id VARCHAR(128),
    status SMALLINT,
    duration BIGINT,
    input TEXT,
    output TEXT,
    error_code VARCHAR(255),
    fail_reason TEXT,
    input_tokens BIGINT,
    output_tokens BIGINT,
    updated_at BIGINT,
    root_execution_id BIGINT,
    parent_node_id VARCHAR(128),
    app_id BIGINT,
    node_count SMALLINT,
    resume_event_id BIGINT,
    agent_id BIGINT,
    sync_pattern SMALLINT,
    commit_id VARCHAR(255),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_workflow_id ON workflow_execution (workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_execution_app_id ON workflow_execution (app_id);
COMMENT ON TABLE workflow_execution IS 'Workflow execution record';

-- Create 'workflow_meta' table
CREATE TABLE IF NOT EXISTS workflow_meta (
    id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    description VARCHAR(2000) NOT NULL,
    icon_uri VARCHAR(256) NOT NULL,
    status SMALLINT NOT NULL,
    content_type SMALLINT NOT NULL,
    mode SMALLINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    deleted_at TIMESTAMPTZ(3),
    creator_id BIGINT NOT NULL,
    tag SMALLINT,
    author_id BIGINT NOT NULL,
    space_id BIGINT NOT NULL,
    updater_id BIGINT,
    source_id BIGINT,
    app_id BIGINT,
    latest_version VARCHAR(50),
    latest_version_ts BIGINT,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_app_id ON workflow_meta (app_id);
CREATE INDEX IF NOT EXISTS idx_latest_version_ts ON workflow_meta (latest_version_ts DESC);
CREATE INDEX IF NOT EXISTS idx_space_id_app_id_status_latest_version_ts ON workflow_meta (space_id, app_id, status, latest_version_ts);
COMMENT ON TABLE workflow_meta IS 'The workflow metadata table';

-- Create 'workflow_reference' table
CREATE TABLE IF NOT EXISTS workflow_reference (
    id BIGINT NOT NULL,
    referred_id BIGINT NOT NULL,
    referring_id BIGINT NOT NULL,
    refer_type SMALLINT NOT NULL,
    referring_biz_type SMALLINT NOT NULL,
    created_at BIGINT NOT NULL,
    status SMALLINT NOT NULL,
    deleted_at TIMESTAMPTZ(3),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_referred_id_referring_biz_type_status ON workflow_reference (referred_id, referring_biz_type, status);
CREATE INDEX IF NOT EXISTS idx_referring_id_status ON workflow_reference (referring_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_referred_id_referring_id_refer_type ON workflow_reference (referred_id, referring_id, refer_type);
COMMENT ON TABLE workflow_reference IS 'The workflow association table';

-- Create 'workflow_snapshot' table
CREATE TABLE IF NOT EXISTS workflow_snapshot (
    workflow_id BIGINT NOT NULL,
    commit_id VARCHAR(255) NOT NULL,
    canvas TEXT,
    input_params TEXT,
    output_params TEXT,
    created_at BIGINT NOT NULL,
    id BIGSERIAL,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_workflow_id_commit_id ON workflow_snapshot (workflow_id, commit_id);
COMMENT ON TABLE workflow_snapshot IS 'snapshot for executed workflow draft';

-- Create 'workflow_version' table
CREATE TABLE IF NOT EXISTS workflow_version (
    id BIGSERIAL,
    workflow_id BIGINT NOT NULL,
    version VARCHAR(50) NOT NULL,
    version_description VARCHAR(2000) NOT NULL,
    canvas TEXT,
    input_params TEXT,
    output_params TEXT,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ(3),
    commit_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_id_created_at ON workflow_version (workflow_id, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_workflow_id_version ON workflow_version (workflow_id, version);
COMMENT ON TABLE workflow_version IS 'Workflow Canvas Version Information Table';

-- Create 'eval_set' table
CREATE TABLE IF NOT EXISTS eval_set (
    id BIGSERIAL,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    schema_json JSONB,
    item_count BIGINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_eval_set_created_by ON eval_set (created_by);
COMMENT ON TABLE eval_set IS 'evaluation set';
COMMENT ON COLUMN eval_set.id IS 'id';
COMMENT ON COLUMN eval_set.space_id IS 'space id';
COMMENT ON COLUMN eval_set.name IS 'evaluation set name';
COMMENT ON COLUMN eval_set.schema_json IS 'field schema definition';
COMMENT ON COLUMN eval_set.item_count IS 'item count';
COMMENT ON COLUMN eval_set.status IS 'status 0 draft 1 enabled';
COMMENT ON COLUMN eval_set.created_by IS 'creator id';

-- Create 'eval_set_item' table
CREATE TABLE IF NOT EXISTS eval_set_item (
    id BIGSERIAL,
    eval_set_id BIGINT NOT NULL,
    data_json JSONB NOT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_eval_set_item_set ON eval_set_item (eval_set_id);
COMMENT ON TABLE eval_set_item IS 'evaluation set item';
COMMENT ON COLUMN eval_set_item.id IS 'id';
COMMENT ON COLUMN eval_set_item.eval_set_id IS 'evaluation set id';
COMMENT ON COLUMN eval_set_item.data_json IS 'item data';

-- Create 'evaluator' table
CREATE TABLE IF NOT EXISTS evaluator (
    id BIGSERIAL,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    type SMALLINT NOT NULL DEFAULT 1,
    model_id VARCHAR(64) NOT NULL DEFAULT '',
    prompt TEXT NOT NULL,
    temperature DOUBLE PRECISION NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_evaluator_created_by ON evaluator (created_by);
COMMENT ON TABLE evaluator IS 'evaluator';
COMMENT ON COLUMN evaluator.id IS 'id';
COMMENT ON COLUMN evaluator.space_id IS 'space id';
COMMENT ON COLUMN evaluator.name IS 'evaluator name';
COMMENT ON COLUMN evaluator.type IS 'type 1 prompt';
COMMENT ON COLUMN evaluator.model_id IS 'model id';
COMMENT ON COLUMN evaluator.prompt IS 'scoring prompt template';
COMMENT ON COLUMN evaluator.status IS 'status 0 draft 1 enabled';
COMMENT ON COLUMN evaluator.created_by IS 'creator id';

-- Create 'experiment' table
CREATE TABLE IF NOT EXISTS experiment (
    id BIGSERIAL,
    space_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    eval_set_id BIGINT NOT NULL,
    target_type SMALLINT NOT NULL,
    target_id VARCHAR(128) NOT NULL DEFAULT '',
    target_config_json JSONB,
    evaluator_ids JSONB,
    concurrency INT NOT NULL DEFAULT 1,
    status SMALLINT NOT NULL DEFAULT 0,
    run_stats_json JSONB,
    error_msg VARCHAR(2000) NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ(3),
    finished_at TIMESTAMPTZ(3),
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_experiment_created_by ON experiment (created_by);
COMMENT ON TABLE experiment IS 'evaluation experiment';
COMMENT ON COLUMN experiment.id IS 'id';
COMMENT ON COLUMN experiment.space_id IS 'space id';
COMMENT ON COLUMN experiment.name IS 'experiment name';
COMMENT ON COLUMN experiment.eval_set_id IS 'evaluation set id';
COMMENT ON COLUMN experiment.target_type IS 'target type 1 agent 2 workflow 3 chatflow';
COMMENT ON COLUMN experiment.target_id IS 'target id';
COMMENT ON COLUMN experiment.target_config_json IS 'target config';
COMMENT ON COLUMN experiment.evaluator_ids IS 'evaluator id list';
COMMENT ON COLUMN experiment.concurrency IS 'concurrency';
COMMENT ON COLUMN experiment.status IS 'status 0 pending 1 running 2 success 3 failed 4 partial';
COMMENT ON COLUMN experiment.run_stats_json IS 'run stats';
COMMENT ON COLUMN experiment.created_by IS 'creator id';

-- Create 'experiment_item_result' table
CREATE TABLE IF NOT EXISTS experiment_item_result (
    id BIGSERIAL,
    experiment_id BIGINT NOT NULL,
    eval_set_item_id BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    input_json JSONB,
    actual_output TEXT,
    output_json JSONB,
    target_error_msg VARCHAR(2000) NOT NULL DEFAULT '',
    evaluator_results_json JSONB,
    tokens_used_json JSONB,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_experiment_item_result_exp ON experiment_item_result (experiment_id);
COMMENT ON TABLE experiment_item_result IS 'experiment item result';
COMMENT ON COLUMN experiment_item_result.id IS 'id';
COMMENT ON COLUMN experiment_item_result.experiment_id IS 'experiment id';
COMMENT ON COLUMN experiment_item_result.eval_set_item_id IS 'evaluation set item id';
COMMENT ON COLUMN experiment_item_result.status IS 'status 0 pending 1 success 2 failed';
COMMENT ON COLUMN experiment_item_result.input_json IS 'input snapshot';
COMMENT ON COLUMN experiment_item_result.actual_output IS 'actual output';
COMMENT ON COLUMN experiment_item_result.evaluator_results_json IS 'evaluator results';
COMMENT ON COLUMN experiment_item_result.tokens_used_json IS 'tokens used';

-- Create 'experiment_aggr_result' table
CREATE TABLE IF NOT EXISTS experiment_aggr_result (
    id BIGSERIAL,
    experiment_id BIGINT NOT NULL,
    evaluator_scores_json JSONB,
    total_item_count BIGINT NOT NULL DEFAULT 0,
    success_item_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_experiment_aggr_exp ON experiment_aggr_result (experiment_id);
COMMENT ON TABLE experiment_aggr_result IS 'experiment aggregation result';
COMMENT ON COLUMN experiment_aggr_result.id IS 'id';
COMMENT ON COLUMN experiment_aggr_result.experiment_id IS 'experiment id';
COMMENT ON COLUMN experiment_aggr_result.evaluator_scores_json IS 'evaluator scores aggregation';

-- Create 'evaluator_custom_template' table
CREATE TABLE IF NOT EXISTS evaluator_custom_template (
    id BIGSERIAL,
    name VARCHAR(100) NOT NULL DEFAULT '',
    description VARCHAR(2000) NOT NULL DEFAULT '',
    prompt TEXT NOT NULL,
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_eval_custom_tpl_created_by ON evaluator_custom_template (created_by);
COMMENT ON TABLE evaluator_custom_template IS 'user custom evaluator template';
COMMENT ON COLUMN evaluator_custom_template.id IS 'id';
COMMENT ON COLUMN evaluator_custom_template.name IS 'template name';
COMMENT ON COLUMN evaluator_custom_template.prompt IS 'scoring prompt template';
