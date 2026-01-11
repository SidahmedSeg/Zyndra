-- Pending Changes Flow
-- Migration: 004_pending_commits.down.sql

ALTER TABLE services DROP COLUMN IF EXISTS auto_deploy;
ALTER TABLE services DROP COLUMN IF EXISTS pending_changes_count;
DROP TABLE IF EXISTS pending_commits;

