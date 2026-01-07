# Quick Start Guide - Day 1

This guide will get you up and running with Click-to-Deploy development in the first day.

---

## Prerequisites

Before starting, ensure you have:

- **Go 1.22+** installed
- **Node.js 18+** and npm installed
- **Docker** and Docker Compose installed
- **PostgreSQL** or **MariaDB** (or use Docker)
- **Git** installed
- **Code editor** (VS Code, GoLand, etc.)

---

## Step 1: Project Setup (30 minutes)

### 1.1 Create Project Structure

```bash
# Create project directory
mkdir click-deploy
cd click-deploy

# Initialize Go module
go mod init github.com/intelifox/click-deploy

# Create directory structure
mkdir -p cmd/server
mkdir -p internal/{api,auth,build,config,domain,git,infra,proxy,store,stream,worker}
mkdir -p migrations/{postgres,sqlite,mysql}
mkdir -p web/src/{components,stores,api}
```

### 1.2 Create Basic Files

**`cmd/server/main.go`**
```go
package main

import (
    "fmt"
    "log"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

**`.gitignore`**
```
# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work

# Dependencies
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# Environment
.env
.env.local

# Database
*.db
*.sqlite

# Frontend
web/node_modules/
web/dist/
web/.vite/

# Logs
*.log
```

**`Makefile`**
```makefile
.PHONY: run test migrate-up migrate-down

run:
	go run cmd/server/main.go

test:
	go test ./...

migrate-up:
	migrate -path migrations/postgres -database "postgres://user:pass@localhost:5432/clickdeploy?sslmode=disable" up

migrate-down:
	migrate -path migrations/postgres -database "postgres://user:pass@localhost:5432/clickdeploy?sslmode=disable" down
```

---

## Step 2: Install Dependencies (15 minutes)

```bash
# Install Go dependencies
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4
go get github.com/kelseyhightower/envconfig
go get github.com/joho/godotenv

# Install migrate CLI (for database migrations)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Step 3: Set Up Local Database (20 minutes)

### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL
docker run -d \
  --name click-deploy-db \
  -e POSTGRES_USER=clickdeploy \
  -e POSTGRES_PASSWORD=devpassword \
  -e POSTGRES_DB=clickdeploy \
  -p 5432:5432 \
  postgres:15

# Verify it's running
docker ps | grep click-deploy-db
```

### Option B: Using Local PostgreSQL

```bash
# Create database
createdb clickdeploy

# Or using psql
psql -U postgres -c "CREATE DATABASE clickdeploy;"
```

---

## Step 4: Create First Migration (30 minutes)

**`migrations/postgres/001_initial.up.sql`**
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

**`migrations/postgres/001_initial.down.sql`**
```sql
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS projects;
```

**Run migration:**
```bash
# Set database URL
export DATABASE_URL="postgres://clickdeploy:devpassword@localhost:5432/clickdeploy?sslmode=disable"

# Run migration
migrate -path migrations/postgres -database "$DATABASE_URL" up
```

---

## Step 5: Create Configuration Package (20 minutes)

**`internal/config/config.go`**
```go
package config

import (
    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    // Server
    Port string `envconfig:"PORT" default:"8080"`
    
    // Database
    DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
    
    // Casdoor
    CasdoorEndpoint   string `envconfig:"CASDOOR_ENDPOINT" required:"true"`
    CasdoorClientID  string `envconfig:"CASDOOR_CLIENT_ID" required:"true"`
    CasdoorClientSecret string `envconfig:"CASDOOR_CLIENT_SECRET" required:"true"`
    
    // OpenStack
    InfraServiceURL  string `envconfig:"INFRA_SERVICE_URL" required:"true"`
    InfraServiceAPIKey string `envconfig:"INFRA_SERVICE_API_KEY" required:"true"`
    
    // Registry
    RegistryURL      string `envconfig:"REGISTRY_URL" required:"true"`
    RegistryUsername string `envconfig:"REGISTRY_USERNAME" required:"true"`
    RegistryPassword string `envconfig:"REGISTRY_PASSWORD" required:"true"`
    
    // GitHub
    GitHubClientID     string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
    GitHubClientSecret string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
}

func Load() (*Config, error) {
    // Load .env file (optional, for local development)
    _ = godotenv.Load()
    
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

**`.env.example`**
```env
# Server
PORT=8080

# Database
DATABASE_URL=postgres://clickdeploy:devpassword@localhost:5432/clickdeploy?sslmode=disable

# Casdoor
CASDOOR_ENDPOINT=http://localhost:8000
CASDOOR_CLIENT_ID=your-client-id
CASDOOR_CLIENT_SECRET=your-client-secret

# OpenStack
INFRA_SERVICE_URL=http://localhost:8081
INFRA_SERVICE_API_KEY=your-api-key

# Registry
REGISTRY_URL=registry.armonika.cloud
REGISTRY_USERNAME=click-deploy
REGISTRY_PASSWORD=your-registry-password

# GitHub
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
```

**Create `.env` file:**
```bash
cp .env.example .env
# Edit .env with your actual values
```

---

## Step 6: Create Basic Store Package (30 minutes)

**`internal/store/db.go`**
```go
package store

import (
    "context"
    "database/sql"
    "fmt"
    
    _ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
    *sql.DB
}

func New(databaseURL string) (*DB, error) {
    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &DB{DB: db}, nil
}

func (db *DB) Close() error {
    return db.DB.Close()
}
```

**`internal/store/projects.go`**
```go
package store

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/google/uuid"
)

type Project struct {
    ID                uuid.UUID
    CasdoorOrgID      string
    Name              string
    Slug              string
    Description       sql.NullString
    OpenStackTenantID string
    OpenStackNetworkID sql.NullString
    DefaultRegion     sql.NullString
    AutoDeploy        bool
    CreatedBy         sql.NullString
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

func (db *DB) CreateProject(ctx context.Context, p *Project) error {
    query := `
        INSERT INTO projects (
            casdoor_org_id, name, slug, description,
            openstack_tenant_id, openstack_network_id,
            default_region, auto_deploy, created_by
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, created_at, updated_at
    `
    
    err := db.QueryRowContext(ctx, query,
        p.CasdoorOrgID, p.Name, p.Slug, p.Description,
        p.OpenStackTenantID, p.OpenStackNetworkID,
        p.DefaultRegion, p.AutoDeploy, p.CreatedBy,
    ).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
    
    return err
}

func (db *DB) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
    var p Project
    query := `SELECT * FROM projects WHERE id = $1`
    
    err := db.QueryRowContext(ctx, query, id).Scan(
        &p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
        &p.OpenStackTenantID, &p.OpenStackNetworkID,
        &p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
        &p.CreatedAt, &p.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    return &p, err
}

func (db *DB) ListProjectsByOrg(ctx context.Context, orgID string) ([]*Project, error) {
    query := `SELECT * FROM projects WHERE casdoor_org_id = $1 ORDER BY created_at DESC`
    
    rows, err := db.QueryContext(ctx, query, orgID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var projects []*Project
    for rows.Next() {
        var p Project
        err := rows.Scan(
            &p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
            &p.OpenStackTenantID, &p.OpenStackNetworkID,
            &p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
            &p.CreatedAt, &p.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        projects = append(projects, &p)
    }
    
    return projects, rows.Err()
}
```

---

## Step 7: Create Basic API Handler (20 minutes)

**`internal/api/projects.go`**
```go
package api

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type ProjectHandler struct {
    store *store.DB
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
    // Get org_id from context (set by auth middleware)
    orgID := r.Context().Value("org_id").(string)
    
    projects, err := h.store.ListProjectsByOrg(r.Context(), orgID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        http.Error(w, "Invalid project ID", http.StatusBadRequest)
        return
    }
    
    project, err := h.store.GetProject(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(project)
}
```

**Update `cmd/server/main.go`:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "github.com/intelifox/click-deploy/internal/api"
    "github.com/intelifox/click-deploy/internal/config"
    "github.com/intelifox/click-deploy/internal/store"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Connect to database
    db, err := store.New(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Set up router
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    
    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    // API routes
    r.Route("/v1/click-deploy", func(r chi.Router) {
        // TODO: Add auth middleware
        projectHandler := &api.ProjectHandler{Store: db}
        r.Get("/projects", projectHandler.ListProjects)
        r.Get("/projects/{id}", projectHandler.GetProject)
    })
    
    // Start server
    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: r,
    }
    
    // Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerConn {
            log.Fatal("Server failed:", err)
        }
    }()
    
    fmt.Printf("Server starting on :%s\n", cfg.Port)
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    fmt.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    fmt.Println("Server exited")
}
```

---

## Step 8: Test It Works (10 minutes)

```bash
# Start the server
go run cmd/server/main.go

# In another terminal, test the health endpoint
curl http://localhost:8080/health
# Should return: OK

# Test projects endpoint (will need auth middleware first)
curl http://localhost:8080/v1/click-deploy/projects
```

---

## Next Steps

1. ✅ **You've completed Day 1 setup!**
2. Continue with **Phase 1** tasks from the Development Plan
3. Next: Implement authentication middleware
4. Then: Complete Projects CRUD operations
5. Follow the Development Plan week by week

---

## Troubleshooting

### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U clickdeploy -d clickdeploy
```

### Migration Issues
```bash
# Check migration status
migrate -path migrations/postgres -database "$DATABASE_URL" version

# Force version (if needed)
migrate -path migrations/postgres -database "$DATABASE_URL" force 1
```

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

---

**Congratulations! You're ready to start development! 🚀**

