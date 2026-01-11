-- Phase 1: Custom JWT Authentication
-- Migration: 003_auth_tables.down.sql
-- This migration rolls back the auth tables

-- Remove new columns from registry_credentials
ALTER TABLE registry_credentials DROP COLUMN IF EXISTS org_id;

-- Remove new columns from git_connections
DROP INDEX IF EXISTS idx_git_connections_new_org;
ALTER TABLE git_connections DROP COLUMN IF EXISTS user_id;
ALTER TABLE git_connections DROP COLUMN IF EXISTS org_id;

-- Remove new columns from projects
ALTER TABLE projects DROP COLUMN IF EXISTS user_id;
DROP INDEX IF EXISTS idx_projects_org;
ALTER TABLE projects DROP COLUMN IF EXISTS org_id;

-- Drop refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;

-- Drop org_members table
DROP TABLE IF EXISTS org_members;

-- Drop organizations table
DROP TABLE IF EXISTS organizations;

-- Drop users table
DROP TABLE IF EXISTS users;

