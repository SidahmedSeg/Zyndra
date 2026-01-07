-- Git Connections (per Organization)
CREATE TABLE git_connections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    casdoor_org_id  VARCHAR(255) NOT NULL,
    provider        VARCHAR(50) NOT NULL, -- github, gitlab
    access_token    TEXT NOT NULL, -- encrypted
    refresh_token   TEXT, -- encrypted
    token_expires_at TIMESTAMPTZ,
    account_name    VARCHAR(255), -- GitHub/GitLab username or org
    account_id      VARCHAR(255), -- Provider's account ID
    connected_by    VARCHAR(255), -- Casdoor user ID who connected
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_git_connections_org ON git_connections(casdoor_org_id);

-- Git Sources (per Service)
CREATE TABLE git_sources (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    git_connection_id UUID REFERENCES git_connections(id),
    provider        VARCHAR(50) NOT NULL, -- github, gitlab
    repo_owner      VARCHAR(255) NOT NULL,
    repo_name       VARCHAR(255) NOT NULL,
    branch          VARCHAR(255) DEFAULT 'main',
    root_dir        VARCHAR(255) DEFAULT '/',
    webhook_id      VARCHAR(255),
    webhook_secret  TEXT, -- encrypted
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_git_sources_service ON git_sources(service_id);

-- Deployments
CREATE TABLE deployments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    commit_sha      VARCHAR(40),
    commit_message  TEXT,
    commit_author   VARCHAR(255),
    status          VARCHAR(50) DEFAULT 'queued',
    image_tag       VARCHAR(255),
    build_duration  INT, -- seconds
    deploy_duration INT, -- seconds
    error_message   TEXT,
    triggered_by    VARCHAR(50), -- webhook, manual, rollback
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_deployments_service ON deployments(service_id);
CREATE INDEX idx_deployments_status ON deployments(status);

-- Deployment Logs
CREATE TABLE deployment_logs (
    id              BIGSERIAL PRIMARY KEY,
    deployment_id   UUID REFERENCES deployments(id) ON DELETE CASCADE,
    timestamp       TIMESTAMPTZ DEFAULT now(),
    phase           VARCHAR(50), -- clone, build, push, deploy
    level           VARCHAR(20), -- info, warn, error
    message         TEXT,
    metadata        JSONB
);

CREATE INDEX idx_deployment_logs_deployment ON deployment_logs(deployment_id);
CREATE INDEX idx_deployment_logs_timestamp ON deployment_logs(timestamp);

-- Environment Variables
CREATE TABLE env_vars (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    key             VARCHAR(255) NOT NULL,
    value           TEXT, -- encrypted, NULL if linked to database
    is_secret       BOOLEAN DEFAULT false,
    -- Database linking (for DATABASE_URL, etc.)
    linked_database_id UUID, -- Will reference databases(id) after table creation
    link_type       VARCHAR(50), -- connection_url, host, port, username, password, database
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(service_id, key)
);

CREATE INDEX idx_env_vars_service ON env_vars(service_id);
CREATE INDEX idx_env_vars_linked_db ON env_vars(linked_database_id);

-- Databases (Managed)
CREATE TABLE databases (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    engine          VARCHAR(50) NOT NULL, -- postgresql, mysql, redis
    version         VARCHAR(50),
    size            VARCHAR(50) DEFAULT 'small',
    -- Auto-created volume
    volume_id       UUID, -- Will reference volumes(id) after table creation
    volume_size_mb  INT DEFAULT 500, -- starts at 500MB
    -- Networking (internal only, NO public access)
    internal_hostname VARCHAR(255), -- e.g., pg7743.internal.armonika.cloud
    internal_ip     INET,
    port            INT,
    -- Credentials (auto-generated)
    username        VARCHAR(255),
    password        TEXT, -- encrypted
    database_name   VARCHAR(255),
    -- Generated connection URL (uses internal hostname)
    connection_url  TEXT, -- e.g., postgresql://user:pass@pg7743.internal.armonika.cloud:5432/mydb
    -- OpenStack resources
    openstack_instance_id VARCHAR(255),
    openstack_port_id     VARCHAR(255),
    security_group_id     VARCHAR(255),
    -- Status
    status          VARCHAR(50) DEFAULT 'pending',
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_databases_service ON databases(service_id);
CREATE INDEX idx_databases_hostname ON databases(internal_hostname);

-- Add foreign key constraint for env_vars.linked_database_id
ALTER TABLE env_vars ADD CONSTRAINT fk_env_vars_linked_database 
    FOREIGN KEY (linked_database_id) REFERENCES databases(id) ON DELETE SET NULL;

-- Volumes
CREATE TABLE volumes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    size_mb         INT NOT NULL, -- size in megabytes
    mount_path      VARCHAR(255), -- e.g., /var/lib/postgresql/data
    -- Attachment
    attached_to_service_id  UUID REFERENCES services(id),
    attached_to_database_id UUID REFERENCES databases(id),
    -- OpenStack
    openstack_volume_id VARCHAR(255),
    openstack_attachment_id VARCHAR(255),
    -- Status
    status          VARCHAR(50) DEFAULT 'pending', -- pending, available, attached
    -- Type
    volume_type     VARCHAR(50) DEFAULT 'user', -- user, database_auto
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_volumes_project ON volumes(project_id);
CREATE INDEX idx_volumes_attached_service ON volumes(attached_to_service_id);
CREATE INDEX idx_volumes_attached_database ON volumes(attached_to_database_id);

-- Add foreign key constraint for databases.volume_id
ALTER TABLE databases ADD CONSTRAINT fk_databases_volume 
    FOREIGN KEY (volume_id) REFERENCES volumes(id) ON DELETE SET NULL;

-- Custom Domains
CREATE TABLE custom_domains (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    domain          VARCHAR(255) NOT NULL UNIQUE,
    -- Validation
    validation_type VARCHAR(50) DEFAULT 'cname', -- cname, txt
    validation_target VARCHAR(255), -- CNAME target (e.g., back4525.projects.armonika.cloud)
    validated_at    TIMESTAMPTZ,
    status          VARCHAR(50) DEFAULT 'pending', -- pending, validating, active, failed
    -- SSL/TLS
    ssl_status      VARCHAR(50) DEFAULT 'pending', -- pending, provisioning, active
    ssl_expires_at  TIMESTAMPTZ,
    ssl_issued_at   TIMESTAMPTZ,
    ssl_issuer      VARCHAR(255),
    ssl_auto_renew  BOOLEAN DEFAULT true,
    -- Metadata
    is_primary      BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_custom_domains_service ON custom_domains(service_id);
CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_status ON custom_domains(status);

-- Job Queue
CREATE TABLE jobs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type            VARCHAR(50) NOT NULL, -- build, deploy, destroy, provision_db
    payload         JSONB NOT NULL,
    status          VARCHAR(50) DEFAULT 'pending',
    attempts        INT DEFAULT 0,
    max_attempts    INT DEFAULT 3,
    error           TEXT,
    run_at          TIMESTAMPTZ DEFAULT now(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    locked_by       VARCHAR(255),
    locked_until    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_jobs_status_run_at ON jobs(status, run_at);
CREATE INDEX idx_jobs_locked ON jobs(locked_until);

-- Registry Credentials
CREATE TABLE registry_credentials (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    casdoor_org_id  VARCHAR(255), -- NULL for global credentials
    registry_url    VARCHAR(255) NOT NULL,
    username        VARCHAR(255) NOT NULL,
    encrypted_password TEXT NOT NULL, -- AES-256-GCM encrypted
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(casdoor_org_id) -- One credential set per org (or global)
);

CREATE INDEX idx_registry_credentials_org ON registry_credentials(casdoor_org_id);

