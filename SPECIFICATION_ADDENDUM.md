# Click-to-Deploy Specification Addendum
## Missing Technical Aspects - Detailed Specifications

**Version:** 1.1  
**Date:** January 2026  
**Status:** Addendum to v1.0 Specification

This document addresses all missing technical aspects identified in the specification analysis. These sections should be integrated into the main specification document.

---

## 8. Container Runtime and Orchestration

### 8.1 Container Runtime Specification

**Technology:** OpenStack Container Service (Zun) or Nova with Docker

**Decision:** Use OpenStack for container orchestration via HTTP API calls. OpenStack is already deployed and running. All container operations are performed through the INTELIFOX OpenStack Service HTTP API.

#### 8.1.1 Container Runtime Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Click to Deploy                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Worker Process                                          │  │
│  │  - Container lifecycle management                        │  │
│  │  - HTTP calls to OpenStack Service                       │  │
│  └───────────────────────┬──────────────────────────────────┘  │
└──────────────────────────┼─────────────────────────────────────┘
                           │ HTTP API Calls
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│          INTELIFOX OpenStack Service                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  OpenStack Zun (Container Service)                      │  │
│  │  OR                                                      │  │
│  │  Nova (Compute) + Docker on instances                   │  │
│  └───────────────────────┬──────────────────────────────────┘  │
└──────────────────────────┼─────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  OpenStack Container/Instance                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Application Container                                   │  │
│  │  - Managed by OpenStack                                  │  │
│  │  - Image from Harbor registry                            │  │
│  │  - Environment variables injected                        │  │
│  │  - Resource limits enforced                              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### 8.1.2 Container Configuration

**Container Image Requirements:**
- Must expose application on specified port (from service configuration)
- Should include health check endpoint (default: `/health`, configurable)
- Must be compatible with Linux x86_64 architecture
- Should use non-root user for security
- Image must be available in Harbor registry

**Container Runtime Settings (via OpenStack API):**
```yaml
Container Runtime: OpenStack Zun or Nova+Docker
Restart Policy: always (managed by OpenStack)
Network: OpenStack network (project network)
User: 1000:1000 (non-root, if supported)
Read-only Root FS: true (where supported)
Resource Limits:
  CPU: Based on instance size (OpenStack flavor)
  Memory: Based on instance size (OpenStack flavor)
  Pids Limit: 1024 (if supported)
```

#### 8.1.3 Container Lifecycle Management via HTTP API

**All container operations are performed via HTTP calls to INTELIFOX OpenStack Service:**

**Startup Sequence:**
1. Call OpenStack API to create container/instance with image reference
2. OpenStack pulls image from Harbor registry (credentials provided)
3. OpenStack creates container with environment variables
4. OpenStack attaches volumes (if configured)
5. OpenStack starts container with restart policy
6. Click-to-Deploy polls for container status
7. Perform health check once container is running
8. Mark service as `live`

**Shutdown Sequence:**
1. Call OpenStack API to stop container (graceful shutdown)
2. OpenStack waits up to 30 seconds for graceful shutdown
3. OpenStack force kills if container doesn't stop
4. Call OpenStack API to delete container
5. OpenStack cleans up volumes (if not persistent)

**Implementation:**
```go
// internal/infra/containers.go
type ContainerClient struct {
    baseURL    string
    httpClient *http.Client
    apiKey     string
}

type CreateContainerRequest struct {
    TenantID      string            `json:"-"`              // Header, not in body
    Name          string            `json:"name"`
    Image         string            `json:"image"`           // Full image path: registry.armonika.cloud/org/project/service:tag
    ImagePullSecret string          `json:"image_pull_secret,omitempty"` // Registry credentials
    Command       []string          `json:"command,omitempty"`
    EnvVars       map[string]string `json:"env_vars"`
    Ports         []PortMapping     `json:"ports"`
    Volumes       []VolumeMount     `json:"volumes,omitempty"`
    Resources     ResourceLimits    `json:"resources"`
    RestartPolicy string            `json:"restart_policy"` // always, never, on-failure
    NetworkID     string            `json:"network_id"`      // Project network
}

type PortMapping struct {
    ContainerPort int    `json:"container_port"`
    HostPort      int    `json:"host_port,omitempty"` // If not specified, OpenStack assigns
    Protocol      string `json:"protocol"`            // tcp, udp
}

type VolumeMount struct {
    VolumeID   string `json:"volume_id"`   // Cinder volume ID
    MountPath  string `json:"mount_path"`  // Path in container
    ReadOnly   bool   `json:"read_only,omitempty"`
}

type ResourceLimits struct {
    CPU    string `json:"cpu"`     // e.g., "2" or "1000m"
    Memory string `json:"memory"`  // e.g., "2Gi" or "2048Mi"
    Pids   int    `json:"pids,omitempty"`
}

type Container struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Status      string            `json:"status"`      // Creating, Running, Stopped, Error
    Image       string            `json:"image"`
    IPAddress   string            `json:"ip_address"` // Private IP on project network
    HostPort    int               `json:"host_port"`   // Port mapped to host
    CreatedAt   time.Time         `json:"created_at"`
    StartedAt   *time.Time        `json:"started_at,omitempty"`
}

// CreateContainer creates a container via OpenStack API
func (c *ContainerClient) CreateContainer(ctx context.Context, req CreateContainerRequest) (*Container, error) {
    resp, err := c.doRequest(ctx, "POST", "/api/containers", req.TenantID, req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var container Container
    if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
        return nil, err
    }
    
    return &container, nil
}

// GetContainerStatus retrieves container status
func (c *ContainerClient) GetContainerStatus(ctx context.Context, tenantID, containerID string) (*Container, error) {
    resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/containers/%s", containerID), tenantID, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var container Container
    if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
        return nil, err
    }
    
    return &container, nil
}

// StopContainer stops a running container
func (c *ContainerClient) StopContainer(ctx context.Context, tenantID, containerID string, timeout int) error {
    req := map[string]interface{}{
        "timeout": timeout, // Seconds to wait for graceful shutdown
    }
    resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/containers/%s/stop", containerID), tenantID, req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

// DeleteContainer deletes a container
func (c *ContainerClient) DeleteContainer(ctx context.Context, tenantID, containerID string) error {
    resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/containers/%s", containerID), tenantID, nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

// WaitForContainerStatus waits for container to reach desired status
func (c *ContainerClient) WaitForContainerStatus(ctx context.Context, tenantID, containerID, desiredStatus string, timeout time.Duration) (*Container, error) {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-ticker.C:
            if time.Now().After(deadline) {
                return nil, fmt.Errorf("timeout waiting for container status %s", desiredStatus)
            }
            
            container, err := c.GetContainerStatus(ctx, tenantID, containerID)
            if err != nil {
                return nil, err
            }
            
            if container.Status == desiredStatus {
                return container, nil
            }
            
            // Check for error states
            if container.Status == "Error" || container.Status == "Failed" {
                return nil, fmt.Errorf("container entered error state: %s", container.Status)
            }
        }
    }
}
```

#### 8.1.4 Container Deployment Flow

```go
// internal/worker/deploy.go
func (w *Worker) deployContainer(ctx context.Context, service *Service, imageTag string) error {
    project, err := w.store.GetProject(ctx, service.ProjectID)
    if err != nil {
        return err
    }
    
    // Resolve environment variables
    envVars, err := w.resolveEnvironmentVariables(ctx, service.ID)
    if err != nil {
        return err
    }
    
    // Get registry credentials
    registryAuth, err := w.getRegistryCredentials(ctx, project.CasdoorOrgID)
    if err != nil {
        return err
    }
    
    // Prepare container request
    containerReq := infra.CreateContainerRequest{
        TenantID:        project.OpenStackTenantID,
        Name:            service.Name,
        Image:           fmt.Sprintf("%s/%s/%s/%s:%s", 
            w.config.RegistryURL,
            project.CasdoorOrgID,
            project.ID,
            service.Name,
            imageTag,
        ),
        ImagePullSecret: registryAuth.Auth, // Base64 encoded registry credentials
        EnvVars:         envVars,
        Ports: []infra.PortMapping{
            {
                ContainerPort: service.Port,
                Protocol:      "tcp",
            },
        },
        Resources: infra.ResourceLimits{
            CPU:    w.getCPUForSize(service.InstanceSize),
            Memory: w.getMemoryForSize(service.InstanceSize),
            Pids:   1024,
        },
        RestartPolicy: "always",
        NetworkID:      project.OpenStackNetworkID,
    }
    
    // Attach volumes if configured
    volumes, err := w.store.GetVolumesByService(ctx, service.ID)
    if err != nil {
        return err
    }
    for _, vol := range volumes {
        containerReq.Volumes = append(containerReq.Volumes, infra.VolumeMount{
            VolumeID:  vol.OpenStackVolumeID,
            MountPath:  vol.MountPath,
            ReadOnly:   false,
        })
    }
    
    // Create container via OpenStack API
    container, err := w.infraClient.CreateContainer(ctx, containerReq)
    if err != nil {
        return err
    }
    
    // Save container ID to service
    err = w.store.UpdateServiceContainerID(ctx, service.ID, container.ID)
    if err != nil {
        return err
    }
    
    // Wait for container to be running
    container, err = w.infraClient.WaitForContainerStatus(
        ctx,
        project.OpenStackTenantID,
        container.ID,
        "Running",
        120*time.Second,
    )
    if err != nil {
        return err
    }
    
    // Update service with container IP and port
    err = w.store.UpdateServiceContainerInfo(ctx, service.ID, container.IPAddress, container.HostPort)
    if err != nil {
        return err
    }
    
    // Perform health check
    return w.performHealthCheck(ctx, service)
}
```

#### 8.1.5 OpenStack API Endpoints

**Container Management Endpoints (INTELIFOX OpenStack Service):**

| Operation | HTTP Method | Endpoint | Description |
|-----------|-------------|----------|-------------|
| Create Container | POST | `/api/containers` | Create new container |
| Get Container | GET | `/api/containers/:id` | Get container details |
| List Containers | GET | `/api/containers` | List all containers in tenant |
| Stop Container | POST | `/api/containers/:id/stop` | Stop running container |
| Start Container | POST | `/api/containers/:id/start` | Start stopped container |
| Restart Container | POST | `/api/containers/:id/restart` | Restart container |
| Delete Container | DELETE | `/api/containers/:id` | Delete container |
| Get Container Logs | GET | `/api/containers/:id/logs` | Get container logs |
| Execute Command | POST | `/api/containers/:id/exec` | Execute command in container |

**Request Headers:**
```
Authorization: Bearer <api_key>
X-Tenant-ID: <openstack_tenant_id>
Content-Type: application/json
```

**Note:** The actual endpoint structure will be adapted to match the existing INTELIFOX OpenStack Service API contract.

---

## 9. Image Registry Authentication

### 9.1 Registry Authentication Method

**Approach:** Registry credentials passed to OpenStack via HTTP API during container creation

#### 9.1.1 Authentication Flow

1. **Registry Credentials Storage:**
   - Credentials stored encrypted in Click-to-Deploy database
   - Per-organization registry credentials (can override global)
   - Credentials retrieved when creating container

2. **Credential Injection to OpenStack:**
   - Credentials passed in `image_pull_secret` field of container creation request
   - OpenStack uses credentials to authenticate with Harbor registry
   - OpenStack handles image pull and credential management
   - Credentials never stored on container instances

3. **Credential Format (for OpenStack API):**
```json
{
  "registry_url": "registry.armonika.cloud",
  "username": "click-deploy",
  "password": "encrypted_token",
  "auth": "base64(username:password)"
}
```

#### 9.1.2 Implementation

**Registry Credential Retrieval:**
```go
// internal/worker/deploy.go
type RegistryAuth struct {
    RegistryURL string `json:"registry_url"`
    Username   string `json:"username"`
    Password   string `json:"password"`
    Auth       string `json:"auth"` // Base64 encoded username:password
}

func (w *Worker) getRegistryCredentials(ctx context.Context, orgID string) (*RegistryAuth, error) {
    // Get organization's registry credentials (or use global default)
    creds, err := w.store.GetRegistryCredentials(ctx, orgID)
    if err != nil {
        // Fallback to global credentials
        creds, err = w.store.GetGlobalRegistryCredentials(ctx)
        if err != nil {
            return nil, err
        }
    }
    
    // Decrypt password
    password, err := decrypt(creds.EncryptedPassword, w.config.EncryptionKey)
    if err != nil {
        return nil, err
    }
    
    // Create auth string (base64 encoded username:password)
    auth := base64.StdEncoding.EncodeToString([]byte(creds.Username + ":" + password))
    
    return &RegistryAuth{
        RegistryURL: creds.RegistryURL,
        Username:   creds.Username,
        Password:   password,
        Auth:       auth,
    }, nil
}
```

**Container Creation with Registry Auth:**
```go
// When creating container via OpenStack API
containerReq := infra.CreateContainerRequest{
    TenantID: project.OpenStackTenantID,
    Name:     service.Name,
    Image:    fmt.Sprintf("%s/%s/%s/%s:%s", 
        registryAuth.RegistryURL,
        project.CasdoorOrgID,
        project.ID,
        service.Name,
        imageTag,
    ),
    ImagePullSecret: registryAuth.Auth, // Base64 encoded credentials
    // ... rest of container config
}

// OpenStack will use ImagePullSecret to authenticate with registry
// when pulling the image
```

#### 9.1.3 OpenStack Image Pull Process

**How OpenStack Handles Registry Authentication:**

1. **Container Creation Request:**
   - Click-to-Deploy sends container creation request with `image_pull_secret`
   - OpenStack receives request and extracts registry credentials

2. **Image Pull:**
   - OpenStack authenticates with Harbor registry using provided credentials
   - OpenStack pulls image from registry
   - Image cached in OpenStack (if supported)

3. **Container Creation:**
   - OpenStack creates container from pulled image
   - Credentials not stored in container or on instance
   - Container runs with image already available

**Security Considerations:**
- Credentials encrypted at rest in Click-to-Deploy database (AES-256-GCM)
- Credentials transmitted over TLS to OpenStack API
- OpenStack handles credential management (not exposed to containers)
- Credentials never stored on container instances
- Use short-lived tokens where possible (refresh via API)
- Per-organization credentials for multi-tenancy

#### 9.1.4 Registry Credential Storage Schema

```sql
CREATE TABLE registry_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    casdoor_org_id  VARCHAR(255), -- NULL for global credentials
    registry_url    VARCHAR(255) NOT NULL,
    username        VARCHAR(255) NOT NULL,
    encrypted_password TEXT NOT NULL, -- AES-256-GCM encrypted
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(casdoor_org_id) -- One credential set per org (or global)
);

CREATE INDEX idx_registry_credentials_org ON registry_credentials(casdoor_org_id);
```

**Credential Priority:**
1. Organization-specific credentials (if exists)
2. Global default credentials (fallback)
3. Error if no credentials found

---

## 10. Environment Variable Injection

### 10.1 Injection Method

**Approach:** Inject via Docker container environment variables at runtime

#### 10.1.1 Variable Resolution Process

1. **Static Variables:** Directly from `env_vars` table
2. **Database-Linked Variables:** Resolved at deployment time from `databases` table
3. **System Variables:** Auto-injected (e.g., `PORT`, `NODE_ENV`)

#### 10.1.2 Variable Resolution Implementation

```go
// internal/worker/deploy.go
func resolveEnvironmentVariables(ctx context.Context, serviceID uuid.UUID, store *store.Store) (map[string]string, error) {
    envVars := make(map[string]string)
    
    // Get all environment variables for service
    vars, err := store.GetEnvVarsByService(ctx, serviceID)
    if err != nil {
        return nil, err
    }
    
    for _, v := range vars {
        if v.LinkedDatabaseID != nil {
            // Resolve database connection URL
            db, err := store.GetDatabase(ctx, *v.LinkedDatabaseID)
            if err != nil {
                return nil, err
            }
            
            switch v.LinkType {
            case "connection_url":
                envVars[v.Key] = db.ConnectionURL
            case "host":
                envVars[v.Key] = db.InternalHostname
            case "port":
                envVars[v.Key] = strconv.Itoa(db.Port)
            case "username":
                envVars[v.Key] = db.Username
            case "password":
                // Decrypt password
                password, err := decrypt(db.Password, encryptionKey)
                if err != nil {
                    return nil, err
                }
                envVars[v.Key] = password
            case "database":
                envVars[v.Key] = db.DatabaseName
            }
        } else {
            // Static variable - decrypt if secret
            if v.IsSecret {
                value, err := decrypt(v.Value, encryptionKey)
                if err != nil {
                    return nil, err
                }
                envVars[v.Key] = value
            } else {
                envVars[v.Key] = v.Value
            }
        }
    }
    
    // Add system variables
    envVars["PORT"] = strconv.Itoa(service.Port)
    envVars["NODE_ENV"] = "production"
    
    return envVars, nil
}
```

#### 10.1.3 Variable Format for Docker

```go
func formatEnvVars(envVars map[string]string) string {
    var parts []string
    for key, value := range envVars {
        // Escape special characters
        escapedValue := strings.ReplaceAll(value, `"`, `\"`)
        parts = append(parts, fmt.Sprintf(`-e "%s=%s"`, key, escapedValue))
    }
    return strings.Join(parts, " \\\n  ")
}
```

**Security:**
- Secrets encrypted at rest in database
- Decrypted only during deployment
- Never logged or exposed in API responses
- Transmitted securely to instances via cloud-init (TLS)

---

## 11. Health Check Configuration

### 11.1 Health Check Specification

**Default Configuration:**
- **Path:** `/health` (configurable per service)
- **Method:** GET
- **Interval:** 10 seconds
- **Timeout:** 5 seconds
- **Success Threshold:** 1 consecutive success
- **Failure Threshold:** 3 consecutive failures
- **Initial Delay:** 30 seconds (after container start)

#### 11.1.1 Health Check Implementation

```go
// internal/worker/healthcheck.go
type HealthCheckConfig struct {
    Path            string        `json:"path"`            // Default: "/health"
    Method          string        `json:"method"`          // Default: "GET"
    Interval        time.Duration `json:"interval"`        // Default: 10s
    Timeout         time.Duration `json:"timeout"`         // Default: 5s
    SuccessThreshold int          `json:"success_threshold"` // Default: 1
    FailureThreshold int          `json:"failure_threshold"` // Default: 3
    InitialDelay    time.Duration `json:"initial_delay"`   // Default: 30s
}

func (w *Worker) performHealthCheck(ctx context.Context, service *Service) error {
    config := service.HealthCheckConfig
    if config == nil {
        config = &HealthCheckConfig{
            Path:            "/health",
            Method:          "GET",
            Interval:        10 * time.Second,
            Timeout:         5 * time.Second,
            SuccessThreshold: 1,
            FailureThreshold: 3,
            InitialDelay:    30 * time.Second,
        }
    }
    
    // Wait for initial delay
    time.Sleep(config.InitialDelay)
    
    url := fmt.Sprintf("https://%s%s", service.GeneratedURL, config.Path)
    client := &http.Client{Timeout: config.Timeout}
    
    consecutiveFailures := 0
    consecutiveSuccesses := 0
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        req, err := http.NewRequest(config.Method, url, nil)
        if err != nil {
            return err
        }
        
        resp, err := client.Do(req)
        if err != nil || resp.StatusCode >= 400 {
            consecutiveFailures++
            consecutiveSuccesses = 0
            
            if consecutiveFailures >= config.FailureThreshold {
                return fmt.Errorf("health check failed after %d consecutive failures", consecutiveFailures)
            }
        } else {
            consecutiveSuccesses++
            consecutiveFailures = 0
            
            if consecutiveSuccesses >= config.SuccessThreshold {
                return nil // Health check passed
            }
        }
        
        resp.Body.Close()
        time.Sleep(config.Interval)
    }
}
```

#### 11.1.2 Health Check Database Schema Addition

```sql
-- Add to services table
ALTER TABLE services ADD COLUMN health_check_path VARCHAR(255) DEFAULT '/health';
ALTER TABLE services ADD COLUMN health_check_interval INT DEFAULT 10; -- seconds
ALTER TABLE services ADD COLUMN health_check_timeout INT DEFAULT 5; -- seconds
ALTER TABLE services ADD COLUMN health_check_failure_threshold INT DEFAULT 3;
```

#### 11.1.3 Health Check Failure Handling

**On Health Check Failure:**
1. Mark service status as `unhealthy`
2. Log failure to deployment logs
3. Optionally trigger automatic restart (configurable)
4. Send notification to user (if configured)
5. Keep instance running for debugging (don't auto-destroy)

**Automatic Restart Policy:**
- Configurable per service (default: disabled)
- Maximum restart attempts: 3
- Cooldown period: 5 minutes between restart attempts

---

## 12. Database Internal DNS Implementation

### 12.1 Internal DNS Architecture

**Technology:** OpenStack Designate for DNS management

**DNS Zones:**
- **Public Zone:** `projects.armonika.cloud` (for services)
- **Internal Zone:** `internal.armonika.cloud` (for databases)

#### 12.1.1 DNS Record Creation

```go
// internal/infra/dns.go
func (c *Client) CreateInternalDNSRecord(ctx context.Context, req CreateDNSRecordRequest) error {
    // Create A record in internal zone
    record := map[string]interface{}{
        "name":        req.Hostname, // e.g., "pg7743.internal.armonika.cloud"
        "type":        "A",
        "records":     []string{req.IPAddress},
        "ttl":         300,
        "description": fmt.Sprintf("Internal DNS for %s", req.ResourceType),
    }
    
    resp, err := c.doRequest(ctx, "POST", "/api/dns/zones/internal.armonika.cloud/records", req.TenantID, record)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

#### 12.1.2 DNS Resolution on Instances

**Configuration via cloud-init:**
```bash
# Configure DNS resolution
cat > /etc/systemd/resolved.conf <<EOF
[ResolvConf]
DNS=10.0.0.1  # OpenStack Designate DNS server
Domains=internal.armonika.cloud
EOF

systemctl restart systemd-resolved
```

**DNS Server:**
- OpenStack Designate provides DNS service
- Internal DNS server IP: `10.0.0.1` (configured in Neutron)
- All instances use this DNS server for internal resolution

#### 12.1.3 Service Discovery

**Process:**
1. Database created → Internal hostname generated (e.g., `pg7743.internal.armonika.cloud`)
2. DNS A record created in Designate internal zone
3. IP address: Database instance private IP (from project subnet)
4. Services resolve hostname via DNS to get IP
5. Connection uses internal IP (never exposed publicly)

**Connection Flow:**
```
Service Container
    │
    ├─> Resolve: pg7743.internal.armonika.cloud
    │   └─> DNS Query to 10.0.0.1
    │       └─> Returns: 10.0.1.50
    │
    └─> Connect to: postgresql://user:pass@10.0.1.50:5432/db
        └─> Direct connection on project network
```

---

## 13. Volume Attachment Timing and Process

### 13.1 Volume Attachment Specification

**Timing:** Volumes attached **before** container start to ensure availability

#### 13.1.1 Volume Attachment Flow

```
1. Create Cinder Volume (via INTELIFOX Service)
   └─> Volume ID returned
   
2. Attach Volume to Instance (via Cinder API)
   └─> Volume attached to /dev/vdb (or next available)
   
3. Format Volume (if new, via cloud-init)
   └─> mkfs.ext4 /dev/vdb
   
4. Mount Volume (via cloud-init)
   └─> mount /dev/vdb /mnt/data
   
5. Create Mount Point in Container
   └─> Docker bind mount: /mnt/data -> /var/lib/app/data
   
6. Start Container with Volume Mounted
```

#### 13.1.2 Implementation

```go
// internal/worker/volume.go
func (w *Worker) attachVolumeToInstance(ctx context.Context, volumeID, instanceID string, mountPath string) error {
    // 1. Attach volume via Cinder API
    attachment, err := w.infraClient.AttachVolume(ctx, infra.AttachVolumeRequest{
        VolumeID:  volumeID,
        InstanceID: instanceID,
    })
    if err != nil {
        return err
    }
    
    // 2. Wait for attachment to complete
    time.Sleep(5 * time.Second) // Allow time for attachment
    
    // 3. Format and mount via cloud-init (if new volume)
    // This is handled in the instance user-data script
    
    return nil
}
```

**Cloud-init Script for Volume Mounting:**
```bash
# Attach and mount volume
DEVICE="/dev/vdb"
MOUNT_POINT="/mnt/data"

# Check if device exists
if [ -b "$DEVICE" ]; then
    # Check if filesystem exists
    if ! blkid "$DEVICE" > /dev/null 2>&1; then
        # Format volume
        mkfs.ext4 "$DEVICE"
    fi
    
    # Create mount point
    mkdir -p "$MOUNT_POINT"
    
    # Mount volume
    mount "$DEVICE" "$MOUNT_POINT"
    
    # Add to fstab for persistence
    echo "$DEVICE $MOUNT_POINT ext4 defaults 0 2" >> /etc/fstab
fi
```

**Docker Volume Mount:**
```go
// In container creation
hostConfig := &docker.HostConfig{
    Binds: []string{
        fmt.Sprintf("%s:%s:rw", mountPath, containerMountPath),
    },
}
```

#### 13.1.3 Volume Attachment for Databases

**Database Volumes:**
- Auto-created when database is provisioned
- Attached during database instance creation
- Mounted at database-specific path (e.g., `/var/lib/postgresql/data`)
- Never detached (persistent storage)

---

## 14. Rollback Implementation Details

### 14.1 Image Versioning Strategy

**Tagging Scheme:**
- **Format:** `{service-name}-{deployment-id}-{commit-sha[:7]}`
- **Example:** `backend-550e8400-a1b2c3d`
- **Latest Tag:** `{service-name}:latest` (points to current deployment)

#### 14.1.1 Image Storage

**Registry Organization:**
```
registry.armonika.cloud/
  └─ click-deploy/
      └─ {org-id}/
          └─ {project-id}/
              └─ {service-name}/
                  ├─ latest
                  ├─ 550e8400-a1b2c3d
                  ├─ 550e8401-d4e5f6a
                  └─ 550e8402-b7c8d9e
```

**Image Retention:**
- Keep last 10 deployments per service
- Older images automatically cleaned up
- Latest 3 images always retained (for quick rollback)

#### 14.1.2 Rollback Process

```go
// internal/worker/rollback.go
func (w *Worker) rollbackDeployment(ctx context.Context, serviceID, targetDeploymentID uuid.UUID) error {
    // 1. Get target deployment
    targetDeployment, err := w.store.GetDeployment(ctx, targetDeploymentID)
    if err != nil {
        return err
    }
    
    // 2. Verify deployment belongs to service
    if targetDeployment.ServiceID != serviceID {
        return errors.New("deployment does not belong to service")
    }
    
    // 3. Get service
    service, err := w.store.GetService(ctx, serviceID)
    if err != nil {
        return err
    }
    
    // 4. Create new deployment record (for rollback)
    newDeployment := &Deployment{
        ServiceID:     serviceID,
        ImageTag:      targetDeployment.ImageTag,
        CommitSHA:     targetDeployment.CommitSHA,
        CommitMessage: targetDeployment.CommitMessage,
        TriggeredBy:   "rollback",
        Status:        "queued",
    }
    deploymentID, err := w.store.CreateDeployment(ctx, newDeployment)
    if err != nil {
        return err
    }
    
    // 5. Trigger redeploy with old image
    return w.queueJob(ctx, Job{
        Type: "redeploy",
        Payload: map[string]interface{}{
            "service_id":  serviceID,
            "image_tag":   targetDeployment.ImageTag,
            "deployment_id": deploymentID,
        },
    })
}
```

#### 14.1.3 Rollback Redeployment

**Process:**
1. Create new instance with old image tag
2. Wait for health check to pass
3. Reassign floating IP from old instance to new instance
4. Update Caddy route (if IP changed)
5. Terminate old instance
6. Mark rollback as complete

**Zero-Downtime:**
- Achieved via floating IP reassignment
- Old instance kept running until new instance is healthy
- Typical rollback time: 60-90 seconds

---

## 15. Database Backup Implementation

### 15.1 Backup Strategy

**Backup Types:**
1. **Manual Backups:** Triggered by user via API/UI
2. **Scheduled Backups:** Automated daily backups (configurable)
3. **Pre-upgrade Backups:** Automatic before database version upgrades

#### 15.1.1 Backup Process

```go
// internal/worker/database.go
func (w *Worker) backupDatabase(ctx context.Context, databaseID uuid.UUID) error {
    db, err := w.store.GetDatabase(ctx, databaseID)
    if err != nil {
        return err
    }
    
    // 1. Create volume snapshot
    snapshot, err := w.infraClient.CreateVolumeSnapshot(ctx, infra.CreateVolumeSnapshotRequest{
        VolumeID: db.VolumeID,
        Name:     fmt.Sprintf("backup-%s-%d", db.ID, time.Now().Unix()),
    })
    if err != nil {
        return err
    }
    
    // 2. Upload snapshot to Swift (object storage)
    backupURL, err := w.uploadSnapshotToSwift(ctx, snapshot.ID, db.ID)
    if err != nil {
        return err
    }
    
    // 3. Record backup in database
    backup := &DatabaseBackup{
        DatabaseID:  databaseID,
        SnapshotID:  snapshot.ID,
        SwiftURL:   backupURL,
        Size:        snapshot.Size,
        CreatedAt:   time.Now(),
        Status:      "completed",
    }
    return w.store.CreateDatabaseBackup(ctx, backup)
}
```

#### 15.1.2 Backup Storage

**Swift Object Storage:**
- Container: `click-deploy-backups`
- Path: `{org-id}/{project-id}/{database-id}/{timestamp}.snapshot`
- Retention: 30 days (configurable)
- Encryption: At rest encryption enabled

#### 15.1.3 Backup Scheduling

**Scheduled Backup Job:**
```go
// Runs daily at 2 AM UTC
func (w *Worker) scheduleDatabaseBackups(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Get all databases with scheduled backups enabled
            databases, err := w.store.GetDatabasesWithScheduledBackups(ctx)
            if err != nil {
                log.Error("Failed to get databases for backup", err)
                continue
            }
            
            for _, db := range databases {
                if err := w.queueJob(ctx, Job{
                    Type: "backup_database",
                    Payload: map[string]interface{}{
                        "database_id": db.ID,
                    },
                }); err != nil {
                    log.Error("Failed to queue backup job", err)
                }
            }
        }
    }
}
```

#### 15.1.4 Backup Restoration

**Restore Process:**
1. User selects backup from backup list
2. System creates new volume from snapshot
3. Creates new database instance with restored volume
4. Updates connection URLs for linked services
5. Optionally triggers service redeployment

**Database Schema Addition:**
```sql
CREATE TABLE database_backups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id     UUID REFERENCES databases(id) ON DELETE CASCADE,
    snapshot_id     VARCHAR(255) NOT NULL,
    swift_url       TEXT NOT NULL,
    size_mb         INT NOT NULL,
    status          VARCHAR(50) DEFAULT 'pending',
    created_at      TIMESTAMPTZ DEFAULT now(),
    expires_at      TIMESTAMPTZ -- Auto-cleanup after expiration
);

CREATE INDEX idx_database_backups_database ON database_backups(database_id);
CREATE INDEX idx_database_backups_expires ON database_backups(expires_at);
```

---

## 16. Service Discovery Mechanism

### 16.1 Service Discovery Architecture

**Method:** DNS-based service discovery via OpenStack Designate

#### 16.1.1 Discovery Process

1. **Database Registration:**
   - Database created → Internal hostname generated
   - DNS A record created in `internal.armonika.cloud` zone
   - Record points to database private IP

2. **Service Resolution:**
   - Service container queries DNS for database hostname
   - DNS returns private IP address
   - Service connects directly using private IP

#### 16.1.2 DNS Configuration on Instances

**Cloud-init DNS Setup:**
```bash
# Configure systemd-resolved for internal DNS
cat > /etc/systemd/resolved.conf <<EOF
[ResolvConf]
DNS=10.0.0.1
Domains=internal.armonika.cloud projects.armonika.cloud
DNSSEC=no
Cache=yes
EOF

systemctl restart systemd-resolved
```

#### 16.1.3 Connection String Resolution

**Database Connection URL Format:**
```
postgresql://{username}:{password}@{hostname}:{port}/{database}
```

**Resolution Flow:**
1. Service receives connection URL with hostname (e.g., `pg7743.internal.armonika.cloud`)
2. Application performs DNS lookup
3. DNS returns private IP (e.g., `10.0.1.50`)
4. Application connects to `10.0.1.50:5432`
5. Connection stays within project network (never leaves OpenStack)

**Fallback Mechanism:**
- If DNS resolution fails, connection fails immediately
- No fallback to IP addresses (security: hostnames only)
- Retry logic handled by application connection pool

---

## 17. Monitoring and Alerting

### 17.1 Metrics Collection

**Metrics Collected:**
- **Instance Metrics:** CPU, memory, disk I/O, network I/O
- **Application Metrics:** Request rate, response time, error rate
- **Database Metrics:** Connection count, query performance, storage usage
- **Deployment Metrics:** Build time, deploy time, success rate

#### 17.1.1 Metrics Collection Method

**Technology:** Prometheus + Grafana (optional integration)

**Collection Process:**
1. **Instance Metrics:** Collected via OpenStack Ceilometer/Monasca
2. **Application Metrics:** Exposed via `/metrics` endpoint (Prometheus format)
3. **Database Metrics:** Collected via database monitoring queries
4. **Deployment Metrics:** Stored in Click-to-Deploy database

#### 17.1.2 Metrics Storage

**Database Schema:**
```sql
CREATE TABLE service_metrics (
    id              BIGSERIAL PRIMARY KEY,
    service_id      UUID REFERENCES services(id) ON DELETE CASCADE,
    timestamp       TIMESTAMPTZ DEFAULT now(),
    cpu_percent     DECIMAL(5,2),
    memory_mb       INT,
    disk_io_read_mb DECIMAL(10,2),
    disk_io_write_mb DECIMAL(10,2),
    network_in_mb    DECIMAL(10,2),
    network_out_mb  DECIMAL(10,2),
    request_count    INT,
    error_count      INT,
    avg_response_ms  DECIMAL(10,2)
);

CREATE INDEX idx_service_metrics_service_time ON service_metrics(service_id, timestamp);
```

**Retention Policy:**
- Raw metrics: 7 days
- Aggregated metrics (hourly): 30 days
- Aggregated metrics (daily): 1 year

#### 17.1.3 Alerting Rules

**Default Alerts:**
1. **High CPU Usage:** > 80% for 5 minutes
2. **High Memory Usage:** > 90% for 5 minutes
3. **Service Down:** Health check failed for 3 consecutive attempts
4. **High Error Rate:** > 5% error rate for 5 minutes
5. **Disk Space:** > 85% disk usage

**Alert Channels:**
- Email (to project owner)
- Webhook (configurable)
- In-app notifications

**Implementation:**
```go
// internal/monitoring/alerts.go
func (m *Monitor) checkAlerts(ctx context.Context) {
    services, _ := m.store.GetAllServices(ctx)
    
    for _, service := range services {
        metrics, _ := m.store.GetLatestMetrics(ctx, service.ID)
        
        // Check CPU
        if metrics.CPUPercent > 80 {
            m.triggerAlert(ctx, Alert{
                ServiceID: service.ID,
                Type:     "high_cpu",
                Severity: "warning",
                Message: fmt.Sprintf("CPU usage is %.2f%%", metrics.CPUPercent),
            })
        }
        
        // Check memory
        if metrics.MemoryMB > service.MemoryLimit*0.9 {
            m.triggerAlert(ctx, Alert{
                ServiceID: service.ID,
                Type:     "high_memory",
                Severity: "warning",
                Message: fmt.Sprintf("Memory usage is %d MB (%.2f%%)", metrics.MemoryMB, float64(metrics.MemoryMB)/float64(service.MemoryLimit)*100),
            })
        }
        
        // Check health
        if service.Status == "unhealthy" {
            m.triggerAlert(ctx, Alert{
                ServiceID: service.ID,
                Type:     "service_down",
                Severity: "critical",
                Message: "Service health check is failing",
            })
        }
    }
}
```

---

## 18. Worker Scaling and High Availability

### 18.1 Worker Architecture

**Design:** Stateless workers with distributed job queue

#### 18.1.1 Multi-Worker Coordination

**Job Queue Locking:**
- Uses PostgreSQL `SKIP LOCKED` for distributed locking
- Multiple workers can run simultaneously
- Each worker processes jobs independently

**Worker Identification:**
```go
// internal/worker/runner.go
type Worker struct {
    ID        string // Unique worker ID (hostname + PID)
    store     *store.Store
    jobTypes  []string // Types of jobs this worker can process
}

func NewWorker(store *store.Store) *Worker {
    hostname, _ := os.Hostname()
    return &Worker{
        ID: fmt.Sprintf("%s-%d", hostname, os.Getpid()),
        store: store,
        jobTypes: []string{"build", "deploy", "provision_infra", "provision_db"},
    }
}
```

#### 18.1.2 Job Distribution

**Process:**
1. Multiple workers poll job queue simultaneously
2. Each worker uses `SKIP LOCKED` to claim a job
3. Only one worker can claim a specific job
4. Workers process jobs in parallel

**Job Claiming:**
```sql
-- Multiple workers can run this simultaneously
UPDATE jobs SET
    status = 'running',
    locked_by = $1, -- Worker ID
    locked_until = now() + interval '10 minutes',
    started_at = now(),
    attempts = attempts + 1
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending'
        AND run_at <= now()
        AND attempts < max_attempts
        AND type = ANY($2) -- Worker's job types
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED -- Prevents conflicts
    LIMIT 1
)
RETURNING *;
```

#### 18.1.3 Worker High Availability

**Failure Handling:**
- If worker crashes, job lock expires after 10 minutes
- Expired jobs automatically retried
- No single point of failure

**Scaling:**
- Add more worker instances to increase throughput
- Workers automatically distribute load
- No configuration needed for new workers

**Health Monitoring:**
```go
// Worker heartbeat
func (w *Worker) sendHeartbeat(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.store.UpdateWorkerHeartbeat(ctx, w.ID)
        }
    }
}
```

---

## 19. Database Connection Pooling

### 19.1 Connection Pooling Strategy

**Approach:** Application-level connection pooling (not managed by platform)

#### 19.1.1 Connection Limits

**Per Database:**
- **PostgreSQL:** Max 100 connections (configurable)
- **MySQL:** Max 150 connections (configurable)
- **Redis:** No connection limit (connection pooling in client)

**Per Service:**
- Recommended: 10-20 connections per service instance
- Maximum: Based on database connection limit / number of services

#### 19.1.2 Connection Pool Configuration

**Recommended Settings (Application Level):**
```javascript
// Node.js example
const pool = new Pool({
  host: process.env.DATABASE_HOST,
  port: process.env.DATABASE_PORT,
  database: process.env.DATABASE_NAME,
  user: process.env.DATABASE_USER,
  password: process.env.DATABASE_PASSWORD,
  max: 10, // Maximum pool size
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});
```

**Database-Level Configuration:**
```sql
-- PostgreSQL
ALTER DATABASE mydb SET max_connections = 100;

-- MySQL
SET GLOBAL max_connections = 150;
```

#### 19.1.3 Connection Monitoring

**Metrics Collected:**
- Active connections per database
- Connection wait time
- Connection errors
- Pool exhaustion events

**Alerts:**
- Connection count > 80% of limit
- Connection wait time > 1 second
- Connection errors > 10 per minute

---

## 20. Custom Domain SSL Renewal

### 20.1 SSL Certificate Lifecycle

**Certificate Provisioning:**
- Initial: Via Let's Encrypt HTTP-01 challenge (Caddy handles)
- Renewal: Automatic via Caddy (before expiration)

#### 20.1.1 Renewal Process

**Caddy Automatic Renewal:**
- Caddy automatically renews certificates 30 days before expiration
- No manual intervention required
- Renewal happens transparently

**Monitoring:**
```go
// internal/worker/ssl.go
func (w *Worker) monitorSSLCertificates(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour) // Check daily
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            domains, _ := w.store.GetCustomDomainsWithSSL(ctx)
            
            for _, domain := range domains {
                // Check certificate expiration via Caddy API
                certInfo, err := w.caddyClient.GetCertificateInfo(ctx, domain.Domain)
                if err != nil {
                    log.Error("Failed to get certificate info", err)
                    continue
                }
                
                // Update database
                w.store.UpdateDomainSSLStatus(ctx, domain.ID, certInfo)
                
                // Alert if expiring soon (< 7 days)
                if certInfo.DaysUntilExpiry < 7 {
                    w.triggerAlert(ctx, Alert{
                        Type: "ssl_expiring_soon",
                        Message: fmt.Sprintf("SSL certificate for %s expires in %d days", domain.Domain, certInfo.DaysUntilExpiry),
                    })
                }
            }
        }
    }
}
```

#### 20.1.2 Certificate Storage

**Caddy Certificate Storage:**
- Certificates stored in Caddy's data directory
- Backed up to Swift object storage (optional)
- Automatic rotation and cleanup

**Database Tracking:**
```sql
-- Update custom_domains table
ALTER TABLE custom_domains ADD COLUMN ssl_issued_at TIMESTAMPTZ;
ALTER TABLE custom_domains ADD COLUMN ssl_expires_at TIMESTAMPTZ;
ALTER TABLE custom_domains ADD COLUMN ssl_issuer VARCHAR(255);
ALTER TABLE custom_domains ADD COLUMN ssl_auto_renew BOOLEAN DEFAULT true;
```

---

## 21. Git Webhook Security

### 21.1 Webhook Security Implementation

#### 21.1.1 Signature Validation

**GitHub Webhooks:**
```go
// internal/api/webhooks.go
func validateGitHubWebhook(r *http.Request, secret string, body []byte) bool {
    signature := r.Header.Get("X-Hub-Signature-256")
    if signature == "" {
        return false
    }
    
    // Remove "sha256=" prefix
    signature = strings.TrimPrefix(signature, "sha256=")
    
    // Compute HMAC
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

**GitLab Webhooks:**
```go
func validateGitLabWebhook(r *http.Request, secret string, body []byte) bool {
    token := r.Header.Get("X-Gitlab-Token")
    return token == secret
}
```

#### 21.1.2 Rate Limiting

**Implementation:**
```go
// Rate limit: 10 webhooks per minute per repository
func (h *WebhookHandler) rateLimitWebhook(repoOwner, repoName string) error {
    key := fmt.Sprintf("webhook:%s:%s", repoOwner, repoName)
    
    count, err := h.redis.Incr(ctx, key)
    if err != nil {
        return err
    }
    
    if count == 1 {
        h.redis.Expire(ctx, key, 60*time.Second)
    }
    
    if count > 10 {
        return errors.New("webhook rate limit exceeded")
    }
    
    return nil
}
```

#### 21.1.3 Webhook Retry Mechanism

**GitHub/GitLab Retry:**
- Providers retry failed webhooks automatically
- Exponential backoff: 1s, 2s, 4s, 8s, 16s
- Maximum 5 retry attempts
- Webhook endpoint must be idempotent

**Idempotency:**
```go
// Check if deployment already exists for this commit
func (h *WebhookHandler) handlePushWebhook(ctx context.Context, event *PushEvent) error {
    // Check for existing deployment
    existing, err := h.store.GetDeploymentByCommit(ctx, event.Repository, event.CommitSHA)
    if err == nil && existing != nil {
        // Already processed, return success
        return nil
    }
    
    // Create new deployment
    // ...
}
```

---

## 22. Container Image Lifecycle

### 22.1 Image Retention Policy

**Retention Rules:**
- **Latest 10 deployments:** Always retained
- **Latest 3 deployments:** Never deleted (for quick rollback)
- **Older images:** Deleted after 30 days
- **Failed builds:** Deleted after 7 days

#### 22.1.1 Image Cleanup Process

```go
// internal/worker/cleanup.go
func (w *Worker) cleanupOldImages(ctx context.Context) {
    // Get all services
    services, _ := w.store.GetAllServices(ctx)
    
    for _, service := range services {
        // Get deployments for service
        deployments, _ := w.store.GetDeploymentsByService(ctx, service.ID, 100)
        
        // Keep latest 10
        if len(deployments) > 10 {
            toDelete := deployments[10:]
            
            for _, deployment := range toDelete {
                // Check if it's in the protected list (latest 3)
                if deployment.ID == deployments[0].ID || 
                   deployment.ID == deployments[1].ID || 
                   deployment.ID == deployments[2].ID {
                    continue
                }
                
                // Check age
                age := time.Since(deployment.CreatedAt)
                if age > 30*24*time.Hour {
                    // Delete image from registry
                    w.registryClient.DeleteImage(ctx, service.Image, deployment.ImageTag)
                }
            }
        }
    }
}
```

#### 22.1.2 Image Size Limits

**Limits:**
- **Maximum image size:** 10 GB
- **Warning threshold:** 5 GB
- **Build failure:** If image exceeds 10 GB

**Monitoring:**
```go
func (w *Worker) checkImageSize(ctx context.Context, image, tag string) error {
    size, err := w.registryClient.GetImageSize(ctx, image, tag)
    if err != nil {
        return err
    }
    
    if size > 10*1024*1024*1024 { // 10 GB
        return errors.New("image size exceeds 10 GB limit")
    }
    
    if size > 5*1024*1024*1024 { // 5 GB
        // Log warning
        log.Warn("Image size exceeds 5 GB", "image", image, "tag", tag, "size", size)
    }
    
    return nil
}
```

---

## 23. Network Isolation Between Projects

### 23.1 Network Segmentation

**Architecture:** Each project gets isolated network (VPC/subnet)

#### 23.1.1 Project Network Creation

```go
// internal/infra/network.go
func (c *Client) CreateProjectNetwork(ctx context.Context, req CreateProjectNetworkRequest) (*Network, error) {
    // Create isolated network for project
    network := map[string]interface{}{
        "name":         fmt.Sprintf("project-%s", req.ProjectID),
        "tenant_id":    req.TenantID,
        "admin_state_up": true,
    }
    
    resp, err := c.doRequest(ctx, "POST", "/api/networks", req.TenantID, network)
    // ...
    
    // Create subnet
    subnet := map[string]interface{}{
        "network_id": network.ID,
        "cidr":       req.CIDR, // e.g., "10.0.1.0/24"
        "name":       fmt.Sprintf("project-%s-subnet", req.ProjectID),
    }
    
    // ...
}
```

#### 23.1.2 Network Isolation Rules

**Security Groups:**
- Default: Deny all inter-project traffic
- Allow: Traffic within same project only
- Allow: Outbound internet traffic (for package downloads, etc.)

**Implementation:**
```go
// Default security group for project
defaultSG := &SecurityGroup{
    Name: "project-default",
    Rules: []SecurityGroupRule{
        // Allow all outbound
        {Direction: "egress", Protocol: "any", RemoteIP: "0.0.0.0/0"},
        // Deny inter-project (default)
        // Allow intra-project (explicit rules added per service)
    },
}
```

#### 23.1.3 Inter-Project Communication

**Policy:** Disabled by default (projects are isolated)

**If Needed (Future):**
- VPC peering between projects
- Shared services network
- Explicit allow rules in security groups

---

## 24. Instance Failure and Recovery

### 24.1 Failure Detection

**Methods:**
1. **Health Check Failures:** 3 consecutive failures
2. **Instance Status:** OpenStack reports instance as ERROR
3. **Container Status:** Container exits unexpectedly

#### 24.1.1 Automatic Recovery

**Recovery Policy (Configurable per service):**
- **Enabled:** Automatic restart on failure
- **Max Restarts:** 3 attempts
- **Cooldown:** 5 minutes between restart attempts
- **Action:** Recreate instance with same image

**Implementation:**
```go
// internal/worker/recovery.go
func (w *Worker) handleInstanceFailure(ctx context.Context, serviceID uuid.UUID) error {
    service, err := w.store.GetService(ctx, serviceID)
    if err != nil {
        return err
    }
    
    // Check if auto-recovery is enabled
    if !service.AutoRecover {
        // Just mark as failed, don't recover
        return w.store.UpdateServiceStatus(ctx, serviceID, "failed")
    }
    
    // Check restart count
    if service.RestartCount >= 3 {
        // Max restarts reached, mark as failed
        return w.store.UpdateServiceStatus(ctx, serviceID, "failed")
    }
    
    // Wait for cooldown
    if time.Since(service.LastRestartAt) < 5*time.Minute {
        return errors.New("still in cooldown period")
    }
    
    // Trigger redeployment
    return w.queueJob(ctx, Job{
        Type: "redeploy",
        Payload: map[string]interface{}{
            "service_id": serviceID,
            "image_tag":  service.CurrentImageTag,
            "reason":     "automatic_recovery",
        },
    })
}
```

#### 24.1.2 Health Monitoring

**Continuous Monitoring:**
```go
func (w *Worker) monitorServiceHealth(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            services, _ := w.store.GetAllLiveServices(ctx)
            
            for _, service := range services {
                // Check health
                healthy, err := w.checkServiceHealth(ctx, service)
                if err != nil || !healthy {
                    // Mark as unhealthy
                    w.store.UpdateServiceStatus(ctx, service.ID, "unhealthy")
                    
                    // Trigger recovery if enabled
                    if service.AutoRecover {
                        w.handleInstanceFailure(ctx, service.ID)
                    }
                }
            }
        }
    }
}
```

---

## 25. Database Credential Rotation

### 25.1 Credential Rotation Process

**Trigger:** Manual (via API) or scheduled (configurable)

#### 25.1.1 Rotation Implementation

```go
// internal/worker/database.go
func (w *Worker) rotateDatabaseCredentials(ctx context.Context, databaseID uuid.UUID) error {
    db, err := w.store.GetDatabase(ctx, databaseID)
    if err != nil {
        return err
    }
    
    // 1. Generate new credentials
    newPassword := generateSecurePassword(32)
    newUsername := fmt.Sprintf("db_user_%s", randomString(8))
    
    // 2. Update database (via SSH or API)
    err = w.updateDatabaseCredentials(ctx, db, newUsername, newPassword)
    if err != nil {
        return err
    }
    
    // 3. Update connection URL
    newConnectionURL := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
        db.Engine,
        newUsername,
        newPassword,
        db.InternalHostname,
        db.Port,
        db.DatabaseName,
    )
    
    // 4. Update database record
    err = w.store.UpdateDatabaseCredentials(ctx, databaseID, newUsername, newPassword, newConnectionURL)
    if err != nil {
        return err
    }
    
    // 5. Trigger redeployment of linked services
    linkedServices, _ := w.store.GetServicesLinkedToDatabase(ctx, databaseID)
    for _, service := range linkedServices {
        w.queueJob(ctx, Job{
            Type: "redeploy",
            Payload: map[string]interface{}{
                "service_id": service.ID,
                "reason":     "database_credential_rotation",
            },
        })
    }
    
    return nil
}
```

#### 25.1.2 Zero-Downtime Rotation

**Strategy:**
1. Create new user with new password
2. Grant same permissions as old user
3. Update connection URLs in database
4. Redeploy services (they get new connection URL)
5. Wait for all services to redeploy
6. Drop old user

**Implementation:**
```sql
-- PostgreSQL rotation
CREATE USER new_user WITH PASSWORD 'new_password';
GRANT ALL PRIVILEGES ON DATABASE mydb TO new_user;
-- Services redeploy with new credentials
-- After all services redeployed:
DROP USER old_user;
```

---

## Integration Notes

### How to Integrate This Addendum

1. **Section 8 (Container Runtime):** Add to main spec Section 8 (Build and Deployment Pipeline)
2. **Section 9 (Registry Auth):** Add to Section 8.7 (Infrastructure Provisioning)
3. **Section 10 (Env Vars):** Add to Section 8.7
4. **Section 11 (Health Check):** Add to Section 8.7
5. **Section 12 (DNS):** Add to Section 9 (Database & Networking)
6. **Section 13 (Volumes):** Add to Section 8.7
7. **Section 14 (Rollback):** Add to Section 8.8 (Redeployment)
8. **Section 15 (Backups):** Add to Section 9.4 (Database Volume Management)
9. **Section 16 (Service Discovery):** Add to Section 9.1 (Network Topology)
10. **Section 17 (Monitoring):** Add new section after Section 10
11. **Section 18 (Workers):** Add to Section 10 (Job Queue System)
12. **Section 19 (Connection Pooling):** Add to Section 9 (Database & Networking)
13. **Section 20 (SSL Renewal):** Add to Section 8.4 (Custom Domains)
14. **Section 21 (Webhooks):** Add to Section 5.3.10 (Webhooks)
15. **Section 22 (Image Lifecycle):** Add to Section 8.5 (Railpack Integration)
16. **Section 23 (Network Isolation):** Add to Section 9.1 (Network Topology)
17. **Section 24 (Recovery):** Add new section after Section 8
18. **Section 25 (Credential Rotation):** Add to Section 9.3 (Environment Variable Linking)

---

## Database Schema Updates Required

```sql
-- Services table additions
ALTER TABLE services ADD COLUMN health_check_path VARCHAR(255) DEFAULT '/health';
ALTER TABLE services ADD COLUMN health_check_interval INT DEFAULT 10;
ALTER TABLE services ADD COLUMN health_check_timeout INT DEFAULT 5;
ALTER TABLE services ADD COLUMN health_check_failure_threshold INT DEFAULT 3;
ALTER TABLE services ADD COLUMN auto_recover BOOLEAN DEFAULT false;
ALTER TABLE services ADD COLUMN restart_count INT DEFAULT 0;
ALTER TABLE services ADD COLUMN last_restart_at TIMESTAMPTZ;

-- Custom domains additions
ALTER TABLE custom_domains ADD COLUMN ssl_issued_at TIMESTAMPTZ;
ALTER TABLE custom_domains ADD COLUMN ssl_expires_at TIMESTAMPTZ;
ALTER TABLE custom_domains ADD COLUMN ssl_issuer VARCHAR(255);
ALTER TABLE custom_domains ADD COLUMN ssl_auto_renew BOOLEAN DEFAULT true;

-- New tables
CREATE TABLE service_metrics (...);
CREATE TABLE database_backups (...);
CREATE TABLE alerts (...);
```

---

**End of Addendum**

