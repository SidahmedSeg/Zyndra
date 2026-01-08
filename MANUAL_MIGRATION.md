# Manual Migration Instructions

If automatic migrations aren't working, you can run them manually.

## Option 1: Using Railway Database Connection (Recommended)

1. **Get your Database URL from Railway:**
   - Go to Railway Dashboard
   - Click on your Database service
   - Go to "Variables" tab
   - Copy the `DATABASE_URL` value (or `POSTGRES_URL`)

2. **Connect to Railway Database:**
   - In Railway, click on your Database service
   - Click "Connect" or "Query" tab
   - This opens a database console

3. **Run the migrations:**
   Copy and paste the SQL from each migration file:

   **First, run `migrations/postgres/001_initial.up.sql`:**
   ```sql
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
   ```

   **Then, run `migrations/postgres/002_tables.up.sql`** (see the file for full SQL)

4. **Verify tables were created:**
   ```sql
   \dt
   ```
   Or:
   ```sql
   SELECT table_name FROM information_schema.tables 
   WHERE table_schema = 'public' 
   ORDER BY table_name;
   ```

## Option 2: Using psql Command Line

1. **Install psql** (if not already installed):
   - macOS: `brew install postgresql`
   - Linux: `sudo apt-get install postgresql-client`
   - Windows: Download from PostgreSQL website

2. **Get your Database URL from Railway:**
   - Format: `postgresql://user:password@host:port/database`

3. **Run migrations:**
   ```bash
   # Make script executable
   chmod +x run_migrations.sh
   
   # Run migrations
   ./run_migrations.sh "postgresql://user:pass@host:port/dbname"
   ```

   Or manually:
   ```bash
   psql "postgresql://user:pass@host:port/dbname" -f migrations/postgres/001_initial.up.sql
   psql "postgresql://user:pass@host:port/dbname" -f migrations/postgres/002_tables.up.sql
   ```

## Option 3: Using Railway CLI

1. **Install Railway CLI:**
   ```bash
   npm i -g @railway/cli
   ```

2. **Connect to your project:**
   ```bash
   railway link
   ```

3. **Run migrations:**
   ```bash
   railway run psql $DATABASE_URL -f migrations/postgres/001_initial.up.sql
   railway run psql $DATABASE_URL -f migrations/postgres/002_tables.up.sql
   ```

## Option 4: Direct SQL Execution (Simplest)

1. **Get Database URL from Railway**

2. **Use Railway's Query interface:**
   - Go to Railway → Your Database → Query tab
   - Copy the entire content of `migrations/postgres/001_initial.up.sql`
   - Paste and execute
   - Then copy and execute `migrations/postgres/002_tables.up.sql`

## Verification

After running migrations, verify tables exist:

```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
```

You should see:
- `projects`
- `services`
- `git_connections`
- `git_sources`
- `deployments`
- `deployment_logs`
- `env_vars`
- `databases`
- `volumes`
- `custom_domains`
- `jobs`
- `registry_credentials`
- `schema_migrations`

