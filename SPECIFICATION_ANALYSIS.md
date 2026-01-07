# Click-to-Deploy Specification Analysis

## Executive Summary
This document identifies contradictions, inconsistencies, and missing technical aspects in the Click-to-Deploy specification.

---

## 🔴 CRITICAL CONTRADICTIONS

### 1. **Load Balancer Architecture Contradiction**
**Location:** Section 2.1 vs Section 8.7 vs Section 12.5

**Issue:**
- Section 2.1: "Each service gets a direct floating IP (no load balancer for MVP)"
- Section 8.7: "Each service gets a direct floating IP (no load balancer for MVP)"
- Section 12.5 Phase 4: "Compute, LB, networking" (mentions LB in deliverables)

**Resolution Needed:**
- Clarify: No load balancer in MVP, or remove LB from Phase 4 deliverables
- If using Caddy as reverse proxy, it's technically a load balancer - clarify terminology

### 2. **SSL Certificate Management Contradiction**
**Location:** Section 8.4.3 vs Section 2.4

**Issue:**
- Section 8.4.3: "Store certificate in Barbican (OpenStack secret store), Configure Octavia listener with SNI"
- Section 2.4: Architecture uses Caddy for SSL termination, not Octavia

**Resolution Needed:**
- Decide: Use Caddy for SSL (as per architecture) OR use Octavia + Barbican
- If using Caddy, remove Octavia/Barbican references from SSL section
- If using Octavia, update architecture diagrams

### 3. **Database Provisioning Method Unclear**
**Location:** Section 8.7 vs Section 9.2 vs Section 12.6

**Issue:**
- Section 8.7: "Create Trove/Nova database" (mentions Trove)
- Section 9.2: "Create Database Instance (Nova)" (only Nova)
- Section 12.6: "Implement Trove database provisioning (or Nova + Docker)"

**Resolution Needed:**
- Clarify: Use OpenStack Trove (DBaaS) OR Nova + Docker containers OR Nova + pre-built images
- Each approach has different implications for management, backups, upgrades

### 4. **Job Queue Database Support**
**Location:** Section 10.1 vs Section 2.1

**Issue:**
- Section 10.1: "The job queue uses PostgreSQL with the SKIP LOCKED pattern"
- Section 2.1: "Flexible database: SQLite (dev), PostgreSQL or MariaDB (production)"
- SKIP LOCKED is PostgreSQL-specific, not available in SQLite

**Resolution Needed:**
- Clarify: Job queue requires PostgreSQL/MariaDB, or implement SQLite-compatible locking
- Document that SQLite mode cannot use job queue (or use file-based queue for SQLite)

---

## ⚠️ INCONSISTENCIES

### 5. **WebSocket vs Centrifugo**
**Location:** Throughout document

**Issue:**
- Section 3.1: Lists `nhooyr.io/websocket`
- Section 5.3.4: Endpoints use `WS` protocol
- Section 11: Project structure includes `websocket.go`
- But we discussed using Centrifugo instead

**Resolution Needed:**
- Update specification to use Centrifugo OR keep WebSocket
- If Centrifugo: Update all references, add Centrifugo server component

### 6. **Instance Count Configuration**
**Location:** Section 5.3.3 vs Architecture

**Issue:**
- API example shows `"instance_count": 1`
- Architecture says "one instance per service"
- But API allows instance_count parameter

**Resolution Needed:**
- Clarify: Is instance_count supported in MVP? If not, remove from API
- If supported, document how multiple instances work with single floating IP

### 7. **Service Type Field**
**Location:** Section 4.3.4

**Issue:**
- `services.type` field: `-- app, database, volume`
- But databases and volumes have separate tables
- Services table seems to only be for app services

**Resolution Needed:**
- Clarify: `services.type` should only be 'app', or remove databases/volumes from enum
- Update schema to reflect actual usage

---

## 🔍 MISSING TECHNICAL ASPECTS

### 8. **Container Runtime Not Specified**
**Location:** Missing

**Issue:**
- No specification of container runtime (Docker, containerd, CRI-O)
- No specification of container orchestration (if any)
- How are containers managed on instances?

**Resolution Needed:**
- Specify: Docker daemon on each instance? containerd? Zun (OpenStack)?
- Document container lifecycle management
- Specify how containers are started/stopped/restarted

### 9. **Image Registry Authentication**
**Location:** Missing

**Issue:**
- Section 3.2 mentions Harbor registry
- Section 13.1 has REGISTRY_USERNAME/PASSWORD
- But no specification of how instances authenticate to pull images

**Resolution Needed:**
- Document: How are registry credentials injected to instances?
- Specify: Image pull secrets, cloud-init scripts, or instance metadata?
- Security: How are credentials secured during image pull?

### 10. **Environment Variable Injection Method**
**Location:** Section 8.7

**Issue:**
- Section 8.7 mentions "Inject environment variables"
- But doesn't specify HOW (cloud-init? container env? config file?)

**Resolution Needed:**
- Specify: Method for injecting env vars into containers
- Document: Format, encryption, and security considerations
- Clarify: How database-linked variables are resolved at runtime

### 11. **Health Check Implementation**
**Location:** Section 8.7

**Issue:**
- Health check assumes `/health` endpoint exists
- Not all applications have health endpoints
- No specification of health check configuration

**Resolution Needed:**
- Document: Configurable health check path (default: `/health`)
- Specify: Health check interval, timeout, failure threshold
- Document: What happens if health check fails?

### 12. **Database Internal DNS Implementation**
**Location:** Section 9.1

**Issue:**
- Section 9.1 mentions internal DNS hostnames (e.g., `pg7743.internal.armonika.cloud`)
- But doesn't specify HOW internal DNS is implemented

**Resolution Needed:**
- Specify: OpenStack Designate for internal DNS? Separate DNS server?
- Document: DNS record creation process
- Clarify: How services resolve internal hostnames

### 13. **Volume Attachment Timing**
**Location:** Section 8.7

**Issue:**
- Section 8.7 shows instance creation but doesn't mention when volumes are attached
- Section 9.2 shows volume creation for databases, but timing unclear

**Resolution Needed:**
- Document: When volumes are attached (before instance start? after?)
- Specify: Volume attachment process and mount point configuration
- Clarify: How mount paths are configured in containers

### 14. **Rollback Implementation Details**
**Location:** Section 7.7

**Issue:**
- Section 7.7 mentions rollback but doesn't specify:
  - How previous images are stored/retrieved
  - Image tag versioning strategy
  - What happens to current instance during rollback

**Resolution Needed:**
- Document: Image versioning/tagging strategy
- Specify: Rollback process (same as redeploy with old image?)
- Clarify: Data persistence during rollback

### 15. **Database Backup Implementation**
**Location:** Section 9.4

**Issue:**
- Section 9.4 mentions "Snapshot volume to Swift"
- But doesn't specify:
  - Backup scheduling
  - Backup retention policy
  - Backup restoration process
  - How backups are triggered

**Resolution Needed:**
- Document: Backup trigger mechanism (manual vs scheduled)
- Specify: Backup storage location and format
- Clarify: Restoration process and downtime implications

### 16. **Service Discovery Mechanism**
**Location:** Missing

**Issue:**
- Services need to find databases via internal hostnames
- But no specification of service discovery mechanism
- How do services resolve `pg7743.internal.armonika.cloud`?

**Resolution Needed:**
- Document: DNS-based service discovery (assumed)
- Specify: DNS resolution configuration on instances
- Clarify: Fallback mechanisms if DNS fails

### 17. **Monitoring and Alerting**
**Location:** Missing

**Issue:**
- Section 5.3.8 mentions metrics endpoint
- But no specification of:
  - What metrics are collected
  - How metrics are stored
  - Alerting mechanisms
  - Monitoring dashboards

**Resolution Needed:**
- Document: Metrics collection method (Prometheus? Custom?)
- Specify: Metrics retention and storage
- Clarify: Alerting rules and notification channels

### 18. **Worker Scaling and High Availability**
**Location:** Section 13.1

**Issue:**
- Section 13.1 mentions `WORKER_COUNT` but only for single instance
- No specification of:
  - Horizontal worker scaling
  - Worker high availability
  - Job queue distribution across workers

**Resolution Needed:**
- Document: How multiple worker instances coordinate
- Specify: Leader election or distributed processing
- Clarify: Job queue locking across multiple workers

### 19. **Database Connection Pooling**
**Location:** Missing

**Issue:**
- Services connect to databases but no mention of:
  - Connection pooling
  - Connection limits
  - Connection timeout configuration

**Resolution Needed:**
- Document: Connection pooling strategy (if any)
- Specify: Maximum connections per database
- Clarify: How connection limits are enforced

### 20. **Custom Domain SSL Renewal**
**Location:** Section 8.4.3

**Issue:**
- Section 8.4.3 mentions "Certificate renewal job runs daily"
- But doesn't specify:
  - How renewal is triggered
  - What happens if renewal fails
  - Certificate expiration monitoring

**Resolution Needed:**
- Document: SSL renewal process and automation
- Specify: Failure handling and retry logic
- Clarify: Certificate expiration alerts

### 21. **Git Webhook Security**
**Location:** Section 5.3.10

**Issue:**
- Webhook endpoints mentioned but no specification of:
  - Webhook signature validation details
  - Rate limiting
  - Webhook retry mechanism
  - Webhook delivery guarantees

**Resolution Needed:**
- Document: Webhook signature validation algorithm
- Specify: Rate limiting and abuse prevention
- Clarify: Webhook delivery and retry policies

### 22. **Container Image Lifecycle**
**Location:** Missing

**Issue:**
- Images are built and pushed to registry
- But no specification of:
  - Image retention policy
  - Old image cleanup
  - Image size limits
  - Registry storage management

**Resolution Needed:**
- Document: Image retention and cleanup policies
- Specify: Maximum image size limits
- Clarify: Registry storage management and quotas

### 23. **Network Isolation Between Projects**
**Location:** Section 9.1

**Issue:**
- Section 9.1 mentions "Each project has isolated networking"
- But doesn't specify:
  - How network isolation is enforced
  - Whether projects can communicate (if needed)
  - Network segmentation strategy

**Resolution Needed:**
- Document: Network isolation implementation (VLANs? VPCs?)
- Specify: Inter-project communication rules
- Clarify: Network security boundaries

### 24. **Instance Failure and Recovery**
**Location:** Missing

**Issue:**
- No specification of:
  - What happens when an instance crashes
  - Automatic restart policies
  - Instance health monitoring
  - Failure recovery mechanisms

**Resolution Needed:**
- Document: Instance restart policies
- Specify: Health monitoring and failure detection
- Clarify: Automatic recovery vs manual intervention

### 25. **Database Credential Rotation**
**Location:** Section 9.3

**Issue:**
- Section 9.3 mentions "If DB credentials rotate, services auto-redeploy"
- But doesn't specify:
  - How credentials are rotated
  - Rotation trigger mechanism
  - Service redeployment process during rotation

**Resolution Needed:**
- Document: Credential rotation process and automation
- Specify: How services are notified of credential changes
- Clarify: Zero-downtime rotation strategy

---

## 📋 RECOMMENDATIONS

### High Priority
1. **Resolve SSL certificate management contradiction** (Caddy vs Octavia)
2. **Clarify database provisioning method** (Trove vs Nova)
3. **Specify container runtime and orchestration**
4. **Document image registry authentication**
5. **Clarify job queue database requirements** (PostgreSQL vs SQLite)

### Medium Priority
6. **Document environment variable injection method**
7. **Specify health check configuration**
8. **Clarify internal DNS implementation**
9. **Document rollback implementation details**
10. **Specify service discovery mechanism**

### Low Priority
11. **Add monitoring and alerting specification**
12. **Document worker scaling strategy**
13. **Specify database backup automation**
14. **Clarify network isolation details**
15. **Document instance failure recovery**

---

## ✅ POSITIVE ASPECTS

The specification is generally well-structured and covers:
- Clear architecture overview
- Comprehensive data models
- Detailed API specification
- Good security considerations
- Well-defined user flows
- Clear development phases

Most issues are related to implementation details that need clarification rather than fundamental design flaws.

