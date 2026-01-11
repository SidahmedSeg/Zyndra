-- Pending Changes Flow
-- Migration: 004_pending_commits.up.sql
-- Store pending commits instead of auto-deploying

-- Pending commits table
CREATE TABLE pending_commits (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id      UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    commit_sha      VARCHAR(40) NOT NULL,
    commit_message  TEXT,
    commit_author   VARCHAR(255),
    commit_url      TEXT,
    branch          VARCHAR(255),
    pushed_at       TIMESTAMPTZ DEFAULT now(),
    acknowledged    BOOLEAN DEFAULT FALSE,
    deployed        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_pending_commits_service ON pending_commits(service_id);
CREATE INDEX idx_pending_commits_unacknowledged ON pending_commits(service_id, acknowledged) WHERE acknowledged = FALSE;

-- Add pending_changes_count to services (for quick access)
ALTER TABLE services ADD COLUMN pending_changes_count INT DEFAULT 0;

-- Add auto_deploy column to services (to control per-service behavior)
ALTER TABLE services ADD COLUMN auto_deploy BOOLEAN DEFAULT FALSE;

