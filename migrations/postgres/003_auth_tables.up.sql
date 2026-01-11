-- Phase 1: Custom JWT Authentication
-- Migration: 003_auth_tables.up.sql
-- This migration creates the tables needed for Zyndra's own authentication system

-- Users table
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(255),
    avatar_url      TEXT,
    email_verified  BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);

-- Organizations (multi-tenant support)
CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) UNIQUE NOT NULL,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_owner ON organizations(owner_id);

-- Organization members (many-to-many relationship)
CREATE TABLE org_members (
    org_id          UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(50) DEFAULT 'member', -- owner, admin, member
    joined_at       TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (org_id, user_id)
);

CREATE INDEX idx_org_members_user ON org_members(user_id);

-- Refresh tokens for token rotation
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- Add org_id to projects table (will be populated during migration)
-- We keep casdoor_org_id for backwards compatibility during transition
ALTER TABLE projects ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX idx_projects_org ON projects(org_id);

-- Add user_id for created_by tracking
ALTER TABLE projects ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Update git_connections to support new auth (keep casdoor_org_id for migration)
ALTER TABLE git_connections ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE git_connections ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_git_connections_new_org ON git_connections(org_id);

-- Update registry_credentials (keep casdoor_org_id for migration)
ALTER TABLE registry_credentials ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

