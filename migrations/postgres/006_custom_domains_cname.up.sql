-- Add missing cname columns to custom_domains table
ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS cname VARCHAR(255);
ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS cname_target VARCHAR(255);
ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS ssl_enabled BOOLEAN DEFAULT true;
ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS validation_token VARCHAR(255);

