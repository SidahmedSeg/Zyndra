-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE projects (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    casdoor_org_id      VARCHAR(255) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    slug                VARCHAR(100) NOT NULL,
    description         TEXT,
    openstack_tenant_id VARCHAR(255) NOT NULL,
    openstack_network_id VARCHAR(255),
    default_region      VARCHAR(100),
    auto_deploy         BOOLEAN DEFAULT true,
    created_by          VARCHAR(255),
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now(),
    UNIQUE(casdoor_org_id, slug)
);

CREATE INDEX idx_projects_casdoor_org ON projects(casdoor_org_id);
CREATE INDEX idx_projects_tenant ON projects(openstack_tenant_id);

-- Services table
CREATE TABLE services (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    type                VARCHAR(50) NOT NULL DEFAULT 'app',
    status              VARCHAR(50) DEFAULT 'pending',
    instance_size       VARCHAR(50) DEFAULT 'medium',
    port                INT DEFAULT 8080,
    openstack_instance_id   VARCHAR(255),
    openstack_fip_id        VARCHAR(255),
    openstack_fip_address   INET,
    security_group_id       VARCHAR(255),
    subdomain           VARCHAR(100) UNIQUE,
    generated_url       TEXT,
    current_image_tag   VARCHAR(255),
    canvas_x            INT DEFAULT 0,
    canvas_y            INT DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_services_project ON services(project_id);
CREATE INDEX idx_services_subdomain ON services(subdomain);

