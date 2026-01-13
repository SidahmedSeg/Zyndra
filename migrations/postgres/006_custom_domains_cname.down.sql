-- Remove cname columns from custom_domains table
ALTER TABLE custom_domains DROP COLUMN IF EXISTS cname;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS cname_target;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS ssl_enabled;
ALTER TABLE custom_domains DROP COLUMN IF EXISTS validation_token;

