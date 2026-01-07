# Metrics Collection Strategy

## Recommendation: Deploy Metrics Agent on Instances

**Yes, deploying a metrics agent on instances is the best practice** for Click to Deploy. Here's why and how to implement it.

---

## Why Metrics Agent on Instances?

### ✅ Advantages

1. **Industry Standard**: Node Exporter is the de facto standard for infrastructure metrics
2. **Comprehensive Metrics**: Access to CPU, memory, disk, network at the instance level
3. **Application Metrics**: Can collect container-level metrics via cAdvisor or container runtime
4. **Flexibility**: Works with any infrastructure provider (not just OpenStack)
5. **Standard Prometheus Ecosystem**: Integrates seamlessly with Prometheus scraping
6. **Low Overhead**: Node Exporter uses minimal resources (~10-50MB RAM)

### ❌ Alternative Approaches (and why they're less ideal)

1. **OpenStack Telemetry (Ceilometer/Gnocchi)**
   - ✅ Pros: Native OpenStack integration, no agent needed
   - ❌ Cons: Limited to infrastructure metrics, may not have application-level metrics, requires OpenStack setup
   - **Verdict**: Good for infrastructure-level monitoring, but insufficient for PaaS needs

2. **Direct API Polling**
   - ✅ Pros: No agents needed
   - ❌ Cons: Limited metrics, polling overhead, may not have all needed metrics
   - **Verdict**: Not comprehensive enough

3. **Cloud Provider APIs**
   - ✅ Pros: No agent deployment
   - ❌ Cons: Vendor lock-in, limited metrics, higher latency
   - **Verdict**: Not suitable for multi-cloud PaaS

---

## Recommended Architecture

```
┌─────────────────┐
│  OpenStack      │
│  Instances      │
│                 │
│  ┌───────────┐ │
│  │ Container │ │
│  └───────────┘ │
│  ┌───────────┐ │
│  │Node       │ │  ← Metrics Agent
│  │Exporter   │ │     (port 9100)
│  └───────────┘ │
│  ┌───────────┐ │
│  │cAdvisor   │ │  ← Container Metrics
│  │(optional) │ │     (port 8080)
│  └───────────┘ │
└────────┬───────┘
         │
         │ HTTP Scrape
         │
┌────────▼────────┐
│  Prometheus     │  ← Scrapes all instances
│  Server         │     (service discovery)
└────────┬────────┘
         │
         │ Query API
         │
┌────────▼────────┐
│  Click to       │  ← Our API queries Prometheus
│  Deploy API     │     for metrics display
└─────────────────┘
```

---

## Implementation Plan

### 1. **Node Exporter on Instances**

**What to deploy:**
- **Node Exporter** (prometheus/node-exporter) - Infrastructure metrics
- **cAdvisor** (optional) - Container-level metrics

**How to deploy:**

#### Option A: Cloud-Init Script (Recommended)
Add to instance `user_data` during instance creation:

```yaml
#cloud-config
packages:
  - docker.io

runcmd:
  # Install Node Exporter
  - docker run -d --name=node-exporter --restart=always \
      --net="host" --pid="host" \
      -v "/:/host:ro,rslave" \
      quay.io/prometheus/node-exporter:latest \
      --path.rootfs=/host
  
  # Optional: Install cAdvisor for container metrics
  - docker run -d --name=cadvisor --restart=always \
      --volume=/:/rootfs:ro \
      --volume=/var/run:/var/run:ro \
      --volume=/sys:/sys:ro \
      --volume=/var/lib/docker/:/var/lib/docker:ro \
      --volume=/dev/disk/:/dev/disk:ro \
      --publish=8080:8080 \
      --privileged \
      --device=/dev/kmsg \
      gcr.io/cadvisor/cadvisor:latest
```

#### Option B: Pre-built Image
Create a custom OpenStack image with Node Exporter pre-installed:
- Base image: Ubuntu 22.04 LTS
- Pre-install Docker + Node Exporter + cAdvisor
- Configure auto-start on boot

**Benefits:**
- Faster instance startup
- Consistent configuration
- No cloud-init overhead

### 2. **Prometheus Service Discovery**

Configure Prometheus to discover instances automatically:

```yaml
# prometheus.yml
scrape_configs:
  # OpenStack service discovery
  - job_name: 'openstack-instances'
    openstack_sd_configs:
      - identity_endpoint: 'https://openstack.example.com:5000/v3'
        username: 'prometheus'
        password: 'password'
        project_name: 'monitoring'
        role: 'member'
        region: 'RegionOne'
        port: 9100  # Node Exporter port
        refresh_interval: 60s
    
    relabel_configs:
      # Add labels from OpenStack metadata
      - source_labels: [__meta_openstack_instance_id]
        target_label: instance_id
      - source_labels: [__meta_openstack_instance_name]
        target_label: instance_name
      - source_labels: [__meta_openstack_project_id]
        target_label: project_id
```

**Alternative: File-based Service Discovery**
If OpenStack SD is not available, use file-based discovery:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'click-deploy-instances'
    file_sd_configs:
      - files:
        - '/etc/prometheus/targets/instances.json'
        refresh_interval: 30s
```

**Click to Deploy updates targets file:**
```json
[
  {
    "targets": ["10.0.1.5:9100"],
    "labels": {
      "instance_id": "abc-123",
      "service_id": "svc-456",
      "project_id": "proj-789"
    }
  }
]
```

### 3. **Update Click to Deploy**

#### A. Instance Creation Worker
Update `internal/worker/deploy.go` to include Node Exporter in user_data:

```go
func (w *DeployWorker) createInstanceWithMetrics(ctx context.Context, service *store.Service) error {
    userData := fmt.Sprintf(`#cloud-config
packages:
  - docker.io

runcmd:
  - docker run -d --name=node-exporter --restart=always \
      --net="host" --pid="host" \
      -v "/:/host:ro,rslave" \
      quay.io/prometheus/node-exporter:latest \
      --path.rootfs=/host
`)
    
    // Create instance with user_data
    instance, err := infraClient.CreateInstance(ctx, infra.CreateInstanceRequest{
        Name:     service.Name,
        UserData: userData,
        // ... other fields
    })
    // ...
}
```

#### B. Prometheus Target Registration
After instance creation, register with Prometheus:

```go
// internal/metrics/prometheus_targets.go
func RegisterInstanceTarget(instanceIP string, serviceID, projectID string) error {
    target := PrometheusTarget{
        Targets: []string{fmt.Sprintf("%s:9100", instanceIP)},
        Labels: map[string]string{
            "instance_id": instanceID,
            "service_id":  serviceID,
            "project_id":  projectID,
        },
    }
    // Write to Prometheus file_sd_configs directory
    return writeTargetFile(target)
}
```

#### C. Metrics Collection Worker (Optional)
Create a background worker to periodically collect and store metrics:

```go
// internal/worker/metrics_collector.go
func (w *MetricsCollector) CollectMetrics(ctx context.Context) {
    // Query Prometheus for all instances
    // Store aggregated metrics in our database
    // Or rely on Prometheus as the source of truth (recommended)
}
```

---

## Metrics Mapping

### Node Exporter Metrics → Our Metrics

| Our Metric | Node Exporter Metric | Transformation |
|------------|---------------------|----------------|
| `click_deploy_service_cpu_usage_percent` | `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)` | CPU usage calculation |
| `click_deploy_service_memory_usage_bytes` | `node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes` | Memory used |
| `click_deploy_service_network_traffic_in_bytes_total` | `node_network_receive_bytes_total` | Network inbound |
| `click_deploy_service_network_traffic_out_bytes_total` | `node_network_transmit_bytes_total` | Network outbound |

### Application Metrics (from cAdvisor)

| Our Metric | cAdvisor Metric | Notes |
|------------|-----------------|-------|
| Container CPU | `container_cpu_usage_seconds_total` | Per-container CPU |
| Container Memory | `container_memory_usage_bytes` | Per-container memory |
| Container Network | `container_network_receive_bytes_total` | Container-level network |

### Request Metrics (from Application)

For request count, response time, and error rate, we need application-level metrics:

**Option 1: Application Exports Prometheus Metrics**
- Application exposes `/metrics` endpoint
- Prometheus scrapes it
- Use `http_requests_total`, `http_request_duration_seconds`, etc.

**Option 2: Caddy Metrics**
- Use Caddy's Prometheus metrics endpoint
- Scrape from Caddy instance
- Get request metrics at the reverse proxy level

**Option 3: Application Instrumentation**
- Add Prometheus client library to applications
- Export custom metrics
- Scrape from application endpoints

---

## Implementation Steps

### Phase 1: Basic Infrastructure Metrics (Week 1)
1. ✅ Add Node Exporter to instance user_data
2. ✅ Configure Prometheus file-based service discovery
3. ✅ Update instance creation worker to register targets
4. ✅ Test metrics collection

### Phase 2: Container Metrics (Week 2)
1. ✅ Add cAdvisor to instances
2. ✅ Configure Prometheus to scrape cAdvisor
3. ✅ Update metrics queries to use container metrics

### Phase 3: Application Metrics (Week 3)
1. ✅ Integrate Caddy metrics (if using Caddy)
2. ✅ Or add Prometheus client to applications
3. ✅ Update metrics API to include request metrics

### Phase 4: Optimization (Week 4)
1. ✅ Create pre-built images with agents
2. ✅ Implement Prometheus target auto-registration
3. ✅ Add metrics retention policies
4. ✅ Set up alerting rules

---

## Configuration Updates

### Environment Variables

```bash
# Prometheus
PROMETHEUS_URL=http://localhost:9090
PROMETHEUS_TARGETS_DIR=/etc/prometheus/targets  # For file-based SD

# Metrics Collection
ENABLE_METRICS_COLLECTION=true
METRICS_COLLECTION_INTERVAL=30s
```

### Database Schema (Optional)

If storing metrics in our database:

```sql
CREATE TABLE metrics (
    id UUID PRIMARY KEY,
    resource_type VARCHAR(50),  -- 'service', 'database', 'volume'
    resource_id UUID,
    metric_name VARCHAR(100),
    value FLOAT,
    timestamp TIMESTAMP,
    labels JSONB
);

CREATE INDEX idx_metrics_resource ON metrics(resource_type, resource_id, timestamp);
```

---

## Best Practices

1. **Use Prometheus as Source of Truth**
   - Don't duplicate metrics in our database
   - Query Prometheus directly via API
   - Store only aggregated/derived metrics if needed

2. **Label Strategy**
   - Use consistent labels: `service_id`, `project_id`, `instance_id`
   - Enable easy filtering and aggregation
   - Match our resource IDs for easy correlation

3. **Scraping Interval**
   - Node Exporter: 15-30 seconds
   - Application metrics: 10-15 seconds
   - Balance between freshness and overhead

4. **Resource Limits**
   - Node Exporter: ~10-50MB RAM, minimal CPU
   - cAdvisor: ~30-100MB RAM, low CPU
   - Monitor agent resource usage

5. **Security**
   - Expose metrics only on private network
   - Use Prometheus authentication
   - Restrict access to metrics endpoints

---

## Alternative: OpenStack Telemetry + Node Exporter Hybrid

For comprehensive monitoring:

1. **OpenStack Telemetry** (Ceilometer/Gnocchi)
   - Infrastructure-level metrics (instance creation, deletion, etc.)
   - Billing and quota metrics
   - High-level resource usage

2. **Node Exporter** (on instances)
   - Detailed instance metrics (CPU, memory, disk, network)
   - Container-level metrics (via cAdvisor)
   - Application metrics

**Best of both worlds:**
- OpenStack Telemetry for infrastructure management
- Node Exporter for detailed application monitoring

---

## Conclusion

**Yes, deploying a metrics agent (Node Exporter) on instances is the best practice** because:

1. ✅ Industry standard approach
2. ✅ Comprehensive metrics coverage
3. ✅ Works with any infrastructure
4. ✅ Standard Prometheus ecosystem
5. ✅ Low overhead
6. ✅ Flexible and extensible

**Next Steps:**
1. Update instance creation to include Node Exporter
2. Configure Prometheus service discovery
3. Update metrics API queries to use actual Prometheus data
4. Test end-to-end metrics flow

