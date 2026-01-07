# Click-to-Deploy Development Plan

**Version:** 1.0  
**Date:** January 2026  
**Status:** Active Development Plan

This document outlines the step-by-step development plan for building the Click-to-Deploy PaaS module.

---

## 📋 Development Overview

**Total Estimated Duration:** 12 weeks  
**Team Size:** 2-3 developers  
**Technology Stack:** Go (backend), React (frontend), PostgreSQL/MariaDB, OpenStack

---

## 🎯 Phase 0: Project Setup & Infrastructure (Week 0)

### Objectives
- Set up development environment
- Initialize project structure
- Configure development tools and CI/CD
- Set up local development infrastructure

### Tasks

#### 0.1 Repository Setup
- [ ] Create Git repository
- [ ] Initialize Go module (`go mod init`)
- [ ] Set up `.gitignore` (Go, Node.js, IDE files)
- [ ] Create project structure (as per Section 11 of spec)
- [ ] Set up branch protection rules
- [ ] Configure pre-commit hooks (gofmt, golint)

**Time Estimate:** 4 hours

#### 0.2 Development Environment
- [ ] Set up local PostgreSQL/MariaDB database
- [ ] Configure SQLite for local development
- [ ] Set up Docker Compose for dependencies (BuildKit, Registry)
- [ ] Create `.env.example` with all required variables
- [ ] Set up local Caddy instance (for testing routing)
- [ ] Configure IDE/editor (VS Code, GoLand, etc.)

**Time Estimate:** 8 hours

#### 0.3 Development Tools
- [ ] Set up Go linter (golangci-lint)
- [ ] Configure code formatter (gofmt, goimports)
- [ ] Set up testing framework (testify)
- [ ] Configure logging (zerolog or logrus)
- [ ] Set up API documentation (Swagger/OpenAPI)
- [ ] Configure Makefile for common tasks

**Time Estimate:** 6 hours

#### 0.4 CI/CD Pipeline
- [ ] Set up GitHub Actions / GitLab CI
- [ ] Configure Go build and test jobs
- [ ] Set up database migration testing
- [ ] Configure frontend build pipeline
- [ ] Set up Docker image building
- [ ] Configure deployment pipeline (staging)

**Time Estimate:** 8 hours

**Phase 0 Deliverables:**
- ✅ Project repository initialized
- ✅ Development environment ready
- ✅ CI/CD pipeline configured
- ✅ Local infrastructure running

**Total Time:** ~26 hours (3-4 days)

---

## 🏗️ Phase 1: Foundation (Weeks 1-2)

### Objectives
- Build core API server structure
- Implement database schema and migrations
- Set up authentication middleware
- Create basic CRUD operations for projects and services

### Week 1 Tasks

#### 1.1 Project Structure & Dependencies
- [ ] Create project directory structure
- [ ] Initialize Go modules and add dependencies:
  - [ ] `github.com/go-chi/chi/v5` (router)
  - [ ] `github.com/jackc/pgx/v5` (PostgreSQL driver)
  - [ ] `modernc.org/sqlite` (SQLite driver)
  - [ ] `go-sql-driver/mysql` (MariaDB driver)
  - [ ] `golang-migrate/migrate` (migrations)
  - [ ] `casdoor/casdoor-go-sdk` (authentication)
  - [ ] `kelseyhightower/envconfig` (config)
  - [ ] `nhooyr.io/websocket` or Centrifugo client
- [ ] Set up `cmd/server/main.go` entry point
- [ ] Create `internal/config` package

**Time Estimate:** 8 hours

#### 1.2 Database Setup
- [ ] Create migration directories (`migrations/postgres`, `migrations/sqlite`, `migrations/mysql`)
- [ ] Write initial migration (001_initial.up.sql) for all databases:
  - [ ] Projects table
  - [ ] Services table
  - [ ] Deployments table
  - [ ] Environment variables table
  - [ ] Databases table
  - [ ] Volumes table
  - [ ] Custom domains table
  - [ ] Git connections table
  - [ ] Git sources table
  - [ ] Jobs table
  - [ ] Deployment logs table
- [ ] Write down migrations (001_initial.down.sql)
- [ ] Test migrations on all three database types
- [ ] Create `internal/store` package with database connection logic
- [ ] Implement multi-database support (SQLite/PostgreSQL/MariaDB)

**Time Estimate:** 16 hours

#### 1.3 Authentication Middleware
- [ ] Create `internal/auth` package
- [ ] Implement Casdoor JWT validation
- [ ] Create authentication middleware
- [ ] Extract user context (user_id, org_id, roles) from JWT
- [ ] Add context helpers for accessing user info
- [ ] Write unit tests for JWT validation

**Time Estimate:** 12 hours

#### 1.4 API Router Setup
- [ ] Create `internal/api/router.go`
- [ ] Set up Chi router with middleware:
  - [ ] Authentication middleware
  - [ ] Logging middleware
  - [ ] CORS middleware
  - [ ] Request ID middleware
  - [ ] Recovery middleware
- [ ] Create base API structure (`/v1/click-deploy`)
- [ ] Set up health check endpoint (`/health`)
- [ ] Set up metrics endpoint (`/metrics`)

**Time Estimate:** 8 hours

**Week 1 Deliverables:**
- ✅ Project structure complete
- ✅ Database migrations working
- ✅ Authentication middleware functional
- ✅ Basic API router set up

**Week 1 Total:** ~44 hours

### Week 2 Tasks

#### 2.1 Projects CRUD
- [ ] Create `internal/api/projects.go`
- [ ] Implement `GET /projects` (list organization's projects)
- [ ] Implement `POST /projects` (create project)
- [ ] Implement `GET /projects/:id` (get project details)
- [ ] Implement `PATCH /projects/:id` (update project)
- [ ] Implement `DELETE /projects/:id` (delete project)
- [ ] Create `internal/store/projects.go` with database queries
- [ ] Write unit tests for project operations
- [ ] Write integration tests for API endpoints

**Time Estimate:** 16 hours

#### 2.2 Services CRUD
- [ ] Create `internal/api/services.go`
- [ ] Implement `GET /projects/:id/services` (list services)
- [ ] Implement `POST /projects/:id/services` (create service)
- [ ] Implement `GET /services/:id` (get service details)
- [ ] Implement `PATCH /services/:id` (update service)
- [ ] Implement `DELETE /services/:id` (delete service)
- [ ] Implement `PATCH /services/:id/position` (update canvas position)
- [ ] Create `internal/store/services.go` with database queries
- [ ] Write unit tests and integration tests

**Time Estimate:** 16 hours

#### 2.3 Error Handling & Validation
- [ ] Create `internal/domain/errors.go` with custom error types
- [ ] Implement error response formatting
- [ ] Add request validation (using validator library)
- [ ] Create error middleware for consistent error responses
- [ ] Write error handling tests

**Time Estimate:** 8 hours

#### 2.4 Configuration Management
- [ ] Complete `internal/config/config.go`
- [ ] Load configuration from environment variables
- [ ] Validate required configuration on startup
- [ ] Add configuration for all services (database, OpenStack, etc.)
- [ ] Create configuration documentation

**Time Estimate:** 6 hours

**Week 2 Deliverables:**
- ✅ Projects CRUD API complete
- ✅ Services CRUD API complete
- ✅ Error handling implemented
- ✅ Configuration management complete

**Week 2 Total:** ~46 hours

**Phase 1 Milestone:**
- ✅ API server running
- ✅ Can create/list projects via API
- ✅ Can create/list services via API
- ✅ Authentication working
- ✅ Database migrations tested

**Phase 1 Total:** ~90 hours (2 weeks)

---

## 🔗 Phase 2: Git Integration (Week 3)

### Objectives
- Implement GitHub and GitLab OAuth flows
- Enable repository listing and selection
- Set up webhook registration and handling

### Tasks

#### 2.1 GitHub OAuth
- [ ] Create `internal/git/github.go`
- [ ] Implement OAuth flow initiation (`GET /git/connect/github`)
- [ ] Implement OAuth callback handler (`GET /git/callback/github`)
- [ ] Store OAuth tokens (encrypted) in database
- [ ] Implement token refresh logic
- [ ] Test OAuth flow end-to-end

**Time Estimate:** 12 hours

#### 2.2 GitLab OAuth
- [ ] Create `internal/git/gitlab.go`
- [ ] Implement OAuth flow initiation (`GET /git/connect/gitlab`)
- [ ] Implement OAuth callback handler (`GET /git/callback/gitlab`)
- [ ] Store OAuth tokens (encrypted) in database
- [ ] Implement token refresh logic
- [ ] Test OAuth flow end-to-end

**Time Estimate:** 12 hours

#### 2.3 Repository Operations
- [ ] Implement `GET /git/repos` (list accessible repositories)
- [ ] Implement `GET /git/repos/:owner/:repo/branches` (list branches)
- [ ] Implement `GET /git/repos/:owner/:repo/tree` (get directory structure)
- [ ] Use `go-github` and `go-gitlab` libraries
- [ ] Add caching for repository lists
- [ ] Write tests for repository operations

**Time Estimate:** 10 hours

#### 2.4 Webhook Setup
- [ ] Create `internal/api/webhooks.go`
- [ ] Implement webhook registration when service is created
- [ ] Implement `POST /webhooks/github` handler
- [ ] Implement `POST /webhooks/gitlab` handler
- [ ] Implement webhook signature validation (HMAC-SHA256)
- [ ] Implement rate limiting for webhooks
- [ ] Handle push events and trigger deployments
- [ ] Write tests for webhook handling

**Time Estimate:** 14 hours

#### 2.5 Git Clone Functionality
- [ ] Create `internal/git/clone.go`
- [ ] Implement repository cloning using `go-git`
- [ ] Support private repositories (use OAuth tokens)
- [ ] Implement checkout to specific branch/commit
- [ ] Add temporary directory management
- [ ] Write tests for git operations

**Time Estimate:** 8 hours

**Phase 2 Deliverables:**
- ✅ GitHub OAuth working
- ✅ GitLab OAuth working
- ✅ Repository listing functional
- ✅ Webhooks registered and validated
- ✅ Git clone working

**Phase 2 Total:** ~56 hours (1 week)

---

## 🏗️ Phase 3: Build Pipeline (Weeks 4-5)

### Objectives
- Integrate Railpack for zero-config builds
- Set up BuildKit for container image building
- Integrate with Harbor container registry
- Implement build job processing

### Week 4 Tasks

#### 3.1 BuildKit Setup
- [ ] Set up BuildKit daemon (Docker or standalone)
- [ ] Create `internal/build/buildkit.go`
- [ ] Implement BuildKit client connection
- [ ] Test BuildKit connectivity
- [ ] Configure BuildKit for multi-stage builds

**Time Estimate:** 8 hours

#### 3.2 Railpack Integration
- [ ] Research Railpack API/CLI
- [ ] Create `internal/build/railpack.go`
- [ ] Implement Railpack wrapper for build execution
- [ ] Support runtime detection (Go, Node.js, Python, PHP, Ruby, static)
- [ ] Handle build command customization
- [ ] Test Railpack builds for each runtime

**Time Estimate:** 16 hours

#### 3.3 Container Registry Integration
- [ ] Create `internal/build/registry.go`
- [ ] Implement Harbor registry client
- [ ] Implement image push functionality
- [ ] Implement image tagging strategy
- [ ] Test image push to registry
- [ ] Implement registry authentication

**Time Estimate:** 12 hours

#### 3.4 Build Job Implementation
- [ ] Create `internal/worker/build.go`
- [ ] Implement build job processing:
  - [ ] Clone repository
  - [ ] Run Railpack build
  - [ ] Push image to registry
  - [ ] Update deployment record
- [ ] Implement build log streaming to database
- [ ] Handle build failures and retries
- [ ] Write tests for build process

**Time Estimate:** 16 hours

**Week 4 Deliverables:**
- ✅ BuildKit integrated
- ✅ Railpack builds working
- ✅ Registry integration complete
- ✅ Build jobs processing

**Week 4 Total:** ~52 hours

### Week 5 Tasks

#### 3.5 Dockerfile Fallback
- [ ] Implement Dockerfile detection
- [ ] Support custom Dockerfile paths
- [ ] Implement Dockerfile build (using BuildKit)
- [ ] Allow Dockerfile upload via UI (future)
- [ ] Test Dockerfile builds

**Time Estimate:** 8 hours

#### 3.6 Build Logging
- [ ] Implement build log collection
- [ ] Stream logs to database (`deployment_logs` table)
- [ ] Implement log streaming via Centrifugo (or WebSocket)
- [ ] Add log filtering and search
- [ ] Test log streaming

**Time Estimate:** 10 hours

#### 3.7 Build Cancellation
- [ ] Implement build cancellation
- [ ] Add cancellation endpoint (`POST /deployments/:id/cancel`)
- [ ] Stop BuildKit build process
- [ ] Clean up temporary files
- [ ] Update deployment status

**Time Estimate:** 6 hours

#### 3.8 Build Optimization
- [ ] Implement build caching
- [ ] Optimize image layers
- [ ] Add build time metrics
- [ ] Implement parallel builds (if multiple services)
- [ ] Performance testing

**Time Estimate:** 8 hours

**Week 5 Deliverables:**
- ✅ Dockerfile fallback working
- ✅ Build logs streaming
- ✅ Build cancellation functional
- ✅ Build optimization complete

**Week 5 Total:** ~32 hours

**Phase 3 Milestone:**
- ✅ Push to GitHub triggers image build
- ✅ Images pushed to registry
- ✅ Build logs available in real-time

**Phase 3 Total:** ~84 hours (2 weeks)

---

## ☁️ Phase 4: OpenStack Integration (Weeks 6-7)

### Objectives
- Implement HTTP client for INTELIFOX OpenStack Service
- Integrate infrastructure provisioning
- Implement container deployment via OpenStack
- Set up Caddy routing

### Week 6 Tasks

#### 4.1 OpenStack Client
- [ ] Create `internal/infra/client.go`
- [ ] Implement HTTP client with authentication
- [ ] Add tenant ID header support
- [ ] Implement request/response handling
- [ ] Add retry logic and error handling
- [ ] Write tests for client

**Time Estimate:** 10 hours

#### 4.2 Instance Management
- [ ] Create `internal/infra/instances.go`
- [ ] Implement `CreateInstance` (POST /api/instances)
- [ ] Implement `GetInstance` (GET /api/instances/:id)
- [ ] Implement `WaitForInstanceStatus`
- [ ] Implement `DeleteInstance` (DELETE /api/instances/:id)
- [ ] Test instance operations

**Time Estimate:** 12 hours

#### 4.3 Network Operations
- [ ] Create `internal/infra/network.go`
- [ ] Implement `AllocateFloatingIP` (POST /api/floating-ips)
- [ ] Implement `AttachFloatingIP` (POST /api/floating-ips/:id/attach)
- [ ] Implement `CreateSecurityGroup` (POST /api/security-groups)
- [ ] Implement `CreateDNSRecord` (POST /api/dns/records)
- [ ] Test network operations

**Time Estimate:** 14 hours

#### 4.4 Container Operations
- [ ] Create `internal/infra/containers.go`
- [ ] Implement `CreateContainer` (POST /api/containers)
- [ ] Implement `GetContainerStatus` (GET /api/containers/:id)
- [ ] Implement `StopContainer` (POST /api/containers/:id/stop)
- [ ] Implement `DeleteContainer` (DELETE /api/containers/:id)
- [ ] Implement `WaitForContainerStatus`
- [ ] Test container operations

**Time Estimate:** 12 hours

**Week 6 Deliverables:**
- ✅ OpenStack client functional
- ✅ Instance management working
- ✅ Network operations complete
- ✅ Container operations working

**Week 6 Total:** ~48 hours

### Week 7 Tasks

#### 4.5 Infrastructure Provisioning
- [ ] Create `internal/worker/provision.go`
- [ ] Implement `provision_infra` job:
  - [ ] Generate subdomain
  - [ ] Allocate floating IP
  - [ ] Create DNS record
  - [ ] Create security group
  - [ ] Update service status
- [ ] Test infrastructure provisioning

**Time Estimate:** 12 hours

#### 4.6 Container Deployment
- [ ] Create `internal/worker/deploy.go`
- [ ] Implement `deploy_image` job:
  - [ ] Resolve environment variables
  - [ ] Get registry credentials
  - [ ] Create container via OpenStack API
  - [ ] Wait for container to be running
  - [ ] Attach floating IP
  - [ ] Register route with Caddy
  - [ ] Perform health check
- [ ] Test deployment flow

**Time Estimate:** 16 hours

#### 4.7 Caddy Integration
- [ ] Create `internal/proxy/caddy.go`
- [ ] Implement Caddy Admin API client
- [ ] Implement `AddRoute` (register service route)
- [ ] Implement `RemoveRoute` (remove service route)
- [ ] Implement `AddCustomDomain` (add custom domain route)
- [ ] Test Caddy route management

**Time Estimate:** 10 hours

#### 4.8 Health Checks
- [ ] Create `internal/worker/healthcheck.go`
- [ ] Implement health check logic
- [ ] Support configurable health check paths
- [ ] Implement retry logic
- [ ] Handle health check failures
- [ ] Test health checks

**Time Estimate:** 8 hours

**Week 7 Deliverables:**
- ✅ Infrastructure provisioning working
- ✅ Container deployment functional
- ✅ Caddy routing integrated
- ✅ Health checks working

**Week 7 Total:** ~46 hours

**Phase 4 Milestone:**
- ✅ Built images deployed via OpenStack
- ✅ Services accessible via generated URLs
- ✅ Caddy routing functional

**Phase 4 Total:** ~94 hours (2 weeks)

---

## 💾 Phase 5: Databases & Volumes (Week 8)

### Objectives
- Implement database provisioning
- Implement volume management
- Set up database connection URL generation

### Tasks

#### 5.1 Database Provisioning
- [ ] Create `internal/worker/database.go`
- [ ] Implement `provision_db` job:
  - [ ] Create volume (Cinder)
  - [ ] Create security group
  - [ ] Create database instance (Nova or Trove)
  - [ ] Attach volume
  - [ ] Initialize database
  - [ ] Generate credentials
  - [ ] Create internal DNS record
  - [ ] Generate connection URL
- [ ] Support PostgreSQL, MySQL, Redis
- [ ] Test database provisioning

**Time Estimate:** 20 hours

#### 5.2 Volume Management
- [ ] Create `internal/worker/volume.go`
- [ ] Implement `create_volume` job (POST /api/volumes)
- [ ] Implement `attach_volume` job (POST /api/volumes/:id/attach)
- [ ] Implement `detach_volume` job
- [ ] Implement `destroy_volume` job
- [ ] Test volume operations

**Time Estimate:** 12 hours

#### 5.3 Database API Endpoints
- [ ] Create `internal/api/databases.go`
- [ ] Implement `GET /projects/:id/databases`
- [ ] Implement `POST /projects/:id/databases`
- [ ] Implement `GET /databases/:id`
- [ ] Implement `GET /databases/:id/credentials`
- [ ] Implement `DELETE /databases/:id`
- [ ] Write API tests

**Time Estimate:** 10 hours

#### 5.4 Volume API Endpoints
- [ ] Create `internal/api/volumes.go`
- [ ] Implement `GET /projects/:id/volumes`
- [ ] Implement `POST /projects/:id/volumes`
- [ ] Implement `GET /volumes/:id`
- [ ] Implement `PATCH /volumes/:id/attach`
- [ ] Implement `PATCH /volumes/:id/detach`
- [ ] Implement `DELETE /volumes/:id`
- [ ] Write API tests

**Time Estimate:** 8 hours

#### 5.5 Environment Variable Linking
- [ ] Update `internal/api/env_vars.go`
- [ ] Implement database URL linking
- [ ] Support link types (connection_url, host, port, username, password, database)
- [ ] Implement variable resolution at deployment time
- [ ] Test environment variable linking

**Time Estimate:** 8 hours

**Phase 5 Deliverables:**
- ✅ Database provisioning working
- ✅ Volume management functional
- ✅ Database API complete
- ✅ Volume API complete
- ✅ Environment variable linking working

**Phase 5 Total:** ~58 hours (1 week)

---

## 🎨 Phase 6: UI & Streaming (Weeks 9-10)

### Objectives
- Build Next.js frontend with canvas interface
- Implement real-time log streaming
- Create configuration drawers (large side panels)
- Add deployment progress UI

### Week 9 Tasks

#### 6.1 Frontend Setup
- [ ] Initialize Next.js project (`web/` directory) using Bun
- [ ] Set up Next.js 14+ with App Router
- [ ] Configure Bun as package manager and runtime
- [ ] Install dependencies:
  - [ ] Next.js 14+
  - [ ] React Flow (canvas)
  - [ ] Zustand (state management)
  - [ ] Axios or fetch (HTTP client)
  - [ ] Centrifuge (real-time) or WebSocket
  - [ ] Tailwind CSS
  - [ ] Lucide React (icons)
  - [ ] shadcn/ui (for drawer components)
- [ ] Set up project structure
- [ ] Configure routing (Next.js App Router)

**Time Estimate:** 8 hours

#### 6.2 API Client
- [ ] Create `web/src/api/client.ts`
- [ ] Implement API client with Axios
- [ ] Add authentication (JWT token handling)
- [ ] Implement request/response interceptors
- [ ] Add error handling
- [ ] Create typed API functions

**Time Estimate:** 8 hours

#### 6.3 State Management
- [ ] Create Zustand stores:
  - [ ] `projectsStore.ts` (projects state)
  - [ ] `servicesStore.ts` (services state)
  - [ ] `deploymentsStore.ts` (deployments state)
  - [ ] `canvasStore.ts` (canvas state)
- [ ] Implement state actions
- [ ] Add state persistence (localStorage)

**Time Estimate:** 10 hours

#### 6.4 Canvas Implementation
- [ ] Create `web/src/components/Canvas/Canvas.tsx`
- [ ] Set up React Flow
- [ ] Create node types (ServiceNode, DatabaseNode, VolumeNode)
- [ ] Implement node rendering
- [ ] Implement edge rendering (connections)
- [ ] Add drag and drop
- [ ] Implement canvas zoom/pan

**Time Estimate:** 16 hours

**Week 9 Deliverables:**
- ✅ Frontend project set up
- ✅ API client functional
- ✅ State management working
- ✅ Canvas with nodes rendering

**Week 9 Total:** ~42 hours

### Week 10 Tasks

#### 6.5 Node Components
- [ ] Create `ServiceNode.tsx` component
- [ ] Create `DatabaseNode.tsx` component
- [ ] Create `VolumeNode.tsx` component
- [ ] Implement node status indicators
- [ ] Add node context menus
- [ ] Implement node selection

**Time Estimate:** 12 hours

#### 6.6 Configuration Drawers
- [ ] Create large drawer components (side panels):
  - [ ] `ServiceDrawer.tsx` (with tabs: Source, Instance, Variables, Domains, Deploy, Logs)
  - [ ] `DatabaseDrawer.tsx` (with tabs: Config, Credentials, Backups, Logs)
  - [ ] `VolumeDrawer.tsx` (with tabs: Config, Attached To, Usage)
- [ ] Use shadcn/ui Drawer component (or custom large drawer)
- [ ] Drawer width: ~800px (large enough for full configuration)
- [ ] Implement form validation
- [ ] Add form submission
- [ ] Test drawer workflows

**Time Estimate:** 16 hours

#### 6.7 Real-Time Log Streaming
- [ ] Set up Centrifugo client (or WebSocket)
- [ ] Create `LogStream.tsx` component
- [ ] Implement log streaming for deployments
- [ ] Implement log streaming for services
- [ ] Add log filtering and search
- [ ] Test real-time streaming

**Time Estimate:** 12 hours

#### 6.8 Deployment UI
- [ ] Create deployment progress component
- [ ] Show deployment steps (provision, build, deploy)
- [ ] Display build logs in real-time
- [ ] Show deployment history
- [ ] Implement rollback UI
- [ ] Add deployment status indicators

**Time Estimate:** 10 hours

**Week 10 Deliverables:**
- ✅ Node components complete
- ✅ Configuration drawers functional (large side panels)
- ✅ Real-time log streaming working
- ✅ Deployment UI complete

**Week 10 Total:** ~50 hours

**Phase 6 Milestone:**
- ✅ Full no-code interface working (Next.js + Bun)
- ✅ Canvas with draggable nodes
- ✅ Large configuration drawers (side panels)
- ✅ Real-time logs streaming
- ✅ End-to-end deployment from UI

**Phase 6 Total:** ~92 hours (2 weeks)

---

## ✨ Phase 7: Polish & Production (Weeks 11-12)

### Objectives
- Implement rollback functionality
- Add error handling and retry logic
- Performance optimization
- Security hardening
- Documentation and testing

### Week 11 Tasks

#### 7.1 Rollback Implementation
- [ ] Implement rollback endpoint (`POST /services/:id/rollback/:deployId`)
- [ ] Add rollback job processing
- [ ] Implement image versioning strategy
- [ ] Test rollback functionality
- [ ] Add rollback UI

**Time Estimate:** 12 hours

#### 7.2 Error Handling & Retry
- [ ] Improve error messages (user-friendly)
- [ ] Implement retry logic for failed operations
- [ ] Add exponential backoff
- [ ] Implement circuit breakers for external services
- [ ] Add error recovery mechanisms

**Time Estimate:** 10 hours

#### 7.3 Resource Cleanup
- [ ] Implement cleanup on service deletion
- [ ] Implement cleanup on project deletion
- [ ] Add orphaned resource detection
- [ ] Implement cleanup jobs
- [ ] Test cleanup processes

**Time Estimate:** 8 hours

#### 7.4 Custom Domains
- [ ] Implement `POST /services/:id/domains` (add custom domain)
- [ ] Implement CNAME validation
- [ ] Integrate with Caddy for custom domains
- [ ] Implement SSL certificate provisioning
- [ ] Add domain management UI

**Time Estimate:** 12 hours

#### 7.5 Metrics & Monitoring
- [ ] Implement metrics collection
- [ ] Add metrics endpoint (`GET /services/:id/metrics`)
- [ ] Create metrics visualization components
- [ ] Add alerting rules
- [ ] Test metrics collection

**Time Estimate:** 10 hours

**Week 11 Deliverables:**
- ✅ Rollback functional
- ✅ Error handling improved
- ✅ Resource cleanup working
- ✅ Custom domains functional
- ✅ Metrics collection working

**Week 11 Total:** ~52 hours

### Week 12 Tasks

#### 7.6 Performance Optimization
- [ ] Database query optimization
- [ ] Add database indexes
- [ ] Implement connection pooling
- [ ] Optimize API response times
- [ ] Add caching where appropriate
- [ ] Load testing

**Time Estimate:** 12 hours

#### 7.7 Security Hardening
- [ ] Security audit
- [ ] Implement rate limiting
- [ ] Add input sanitization
- [ ] Review encryption implementation
- [ ] Test for SQL injection, XSS vulnerabilities
- [ ] Implement security headers

**Time Estimate:** 10 hours

#### 7.8 Testing
- [ ] Write integration tests
- [ ] Write end-to-end tests
- [ ] Add API test coverage
- [ ] Test error scenarios
- [ ] Performance testing
- [ ] Load testing

**Time Estimate:** 14 hours

#### 7.9 Documentation
- [ ] Write API documentation (OpenAPI/Swagger)
- [ ] Create user documentation
- [ ] Write deployment guide
- [ ] Create developer guide
- [ ] Document configuration options
- [ ] Create troubleshooting guide

**Time Estimate:** 10 hours

#### 7.10 Production Preparation
- [ ] Set up production database
- [ ] Configure production environment variables
- [ ] Set up monitoring and alerting
- [ ] Create deployment scripts
- [ ] Prepare production release
- [ ] Final testing

**Time Estimate:** 8 hours

**Week 12 Deliverables:**
- ✅ Performance optimized
- ✅ Security hardened
- ✅ Comprehensive testing complete
- ✅ Documentation complete
- ✅ Production ready

**Week 12 Total:** ~54 hours

**Phase 7 Milestone:**
- ✅ Production-ready MVP
- ✅ All features implemented
- ✅ Tested and documented
- ✅ Ready for deployment

**Phase 7 Total:** ~106 hours (2 weeks)

---

## 📊 Summary

| Phase | Duration | Total Hours | Key Deliverables |
|-------|----------|-------------|------------------|
| Phase 0: Setup | 3-4 days | ~26 hours | Project initialized, dev environment ready |
| Phase 1: Foundation | 2 weeks | ~90 hours | API server, database, basic CRUD |
| Phase 2: Git Integration | 1 week | ~56 hours | OAuth, webhooks, repository access |
| Phase 3: Build Pipeline | 2 weeks | ~84 hours | Railpack, BuildKit, registry integration |
| Phase 4: OpenStack | 2 weeks | ~94 hours | Infrastructure provisioning, deployment |
| Phase 5: Databases | 1 week | ~58 hours | Database provisioning, volumes |
| Phase 6: UI & Streaming | 2 weeks | ~92 hours | React frontend, canvas, real-time logs |
| Phase 7: Polish | 2 weeks | ~106 hours | Rollback, testing, documentation |
| **TOTAL** | **12 weeks** | **~606 hours** | **Production-ready MVP** |

---

## 🚀 Getting Started Checklist

### Immediate Next Steps (Day 1)

1. **Set up development environment**
   ```bash
   # Clone repository (when created)
   git clone <repo-url>
   cd click-deploy
   
   # Install Go 1.22+
   # Install Node.js 18+
   # Install Docker
   
   # Set up local database
   docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=dev postgres:15
   
   # Start dependencies
   docker-compose up -d
   ```

2. **Initialize project**
   ```bash
   # Initialize Go module
   go mod init github.com/intelifox/click-deploy
   
   # Create project structure
   mkdir -p cmd/server
   mkdir -p internal/{api,auth,build,config,domain,git,infra,proxy,store,stream,worker}
   mkdir -p migrations/{postgres,sqlite,mysql}
   mkdir -p web/src/{components,stores,api}
   ```

3. **Set up configuration**
   ```bash
   # Copy example env file
   cp .env.example .env
   
   # Edit .env with your values:
   # - DATABASE_URL
   # - CASDOOR_ENDPOINT
   # - GITHUB_CLIENT_ID/SECRET
   # - INFRA_SERVICE_URL
   # - etc.
   ```

4. **Start development**
   ```bash
   # Run database migrations
   make migrate-up
   
   # Start API server
   go run cmd/server/main.go
   
   # Start frontend (in another terminal)
   cd web && npm install && npm run dev
   ```

---

## 📝 Development Guidelines

### Code Standards
- **Go:** Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- **React:** Follow React best practices, use TypeScript
- **Commits:** Use conventional commits (feat:, fix:, docs:, etc.)
- **Testing:** Aim for 80%+ code coverage
- **Documentation:** Document all public APIs

### Git Workflow
- **Main branch:** `main` (production-ready)
- **Development branch:** `develop` (integration branch)
- **Feature branches:** `feature/description` (from develop)
- **Hotfix branches:** `hotfix/description` (from main)

### Daily Standup
- What did you complete yesterday?
- What are you working on today?
- Any blockers?

### Weekly Review
- Review progress against plan
- Adjust timeline if needed
- Demo completed features
- Plan next week's work

---

## 🎯 Success Criteria

The MVP is considered complete when:

- ✅ Users can connect GitHub/GitLab accounts
- ✅ Users can create projects and deploy services
- ✅ Services are automatically built and deployed
- ✅ Services are accessible via generated URLs
- ✅ Users can provision databases and link them to services
- ✅ Real-time logs are streaming
- ✅ Canvas UI is functional
- ✅ Rollback works
- ✅ All critical paths tested
- ✅ Documentation complete

---

## 📞 Support & Resources

### Key Contacts
- **Project Lead:** [Name]
- **OpenStack Integration:** [Name]
- **Frontend Lead:** [Name]

### Useful Links
- [OpenStack API Documentation](https://docs.openstack.org/api-quick-start/)
- [Chi Router Documentation](https://github.com/go-chi/chi)
- [React Flow Documentation](https://reactflow.dev/)
- [Railpack Documentation](https://railpack.com/)

---

**Last Updated:** January 2026  
**Next Review:** End of Phase 1

