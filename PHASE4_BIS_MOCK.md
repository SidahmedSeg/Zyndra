# Phase 4 Bis: Mock OpenStack Integration

## Overview

Phase 4 Bis implements a **mocked OpenStack integration** that simulates all OpenStack operations. This allows us to:
- Build and test the complete deployment flow without needing the actual OpenStack service
- Develop the infrastructure provisioning logic independently
- Test error handling and edge cases easily
- Switch to real OpenStack integration later by simply changing a config flag

## Implementation

### Architecture

The implementation uses an **interface-based design**:

```go
type Client interface {
    CreateInstance(...)
    GetInstance(...)
    AllocateFloatingIP(...)
    // ... etc
}
```

Two implementations:
1. **MockClient** (`internal/infra/mock.go`) - Simulates OpenStack operations
2. **HTTPClient** (`internal/infra/http.go`) - Real HTTP calls (stubs for now)

### Switching Between Mock and Real

```go
// In config
UseMockInfra: true  // Use mock
UseMockInfra: false // Use real HTTP client (when implemented)
```

### Mock Client Features

#### ✅ Instance Operations
- `CreateInstance` - Creates mock instance, simulates async status change (building → active)
- `GetInstance` - Returns mock instance by ID
- `DeleteInstance` - Removes instance from mock store
- `WaitForInstanceStatus` - Polls until instance reaches target status

#### ✅ Network Operations
- `AllocateFloatingIP` - Generates mock floating IP
- `AttachFloatingIP` - Attaches floating IP to instance
- `CreateSecurityGroup` - Creates mock security group
- `CreateDNSRecord` - Creates mock DNS record

#### ✅ Container Operations
- `CreateContainer` - Creates mock container, simulates async status (creating → running)
- `GetContainerStatus` - Returns container status
- `StopContainer` - Stops container
- `DeleteContainer` - Removes container
- `WaitForContainerStatus` - Polls until container reaches target status

#### ✅ Volume Operations
- `CreateVolume` - Creates mock volume
- `AttachVolume` - Attaches volume to instance
- `DetachVolume` - Detaches volume
- `DeleteVolume` - Removes volume

### Mock Data Generation

- **IP Addresses**: Generated mock private IPs (10.0.x.x)
- **Floating IPs**: Generated mock public IPs (41.100.x.x)
- **IDs**: UUID-based for all resources
- **Status Transitions**: Simulated with goroutines and delays

### Thread Safety

All mock operations are protected with `sync.RWMutex` to ensure thread-safe concurrent access.

## Usage

### Configuration

```bash
# Use mock client (default)
USE_MOCK_INFRA=true

# Use real HTTP client (when ready)
USE_MOCK_INFRA=false
INFRA_SERVICE_URL=https://openstack-service.example.com
INFRA_SERVICE_API_KEY=your_api_key
```

### Creating a Client

```go
import "github.com/intelifox/click-deploy/internal/infra"

cfg := infra.Config{
    BaseURL:  config.InfraServiceURL,
    APIKey:   config.InfraServiceAPIKey,
    TenantID: project.OpenStackTenantID,
    UseMock:  config.UseMockInfra,
}

client := infra.NewClient(cfg)
```

### Example: Creating an Instance

```go
instance, err := client.CreateInstance(ctx, infra.CreateInstanceRequest{
    Name:      "my-service",
    FlavorID:  "small",
    ImageID:   "ubuntu-22.04",
    NetworkID: project.OpenStackNetworkID,
    SecurityGroups: []string{"default"},
})

// Mock client will:
// 1. Generate a UUID for instance ID
// 2. Set status to "building"
// 3. After 2 seconds, change status to "active"
// 4. Generate a mock IP address
```

## Benefits

1. **No External Dependencies** - Can develop without OpenStack service running
2. **Fast Development** - No network latency, instant responses
3. **Easy Testing** - Can simulate any scenario (errors, delays, etc.)
4. **Same Interface** - Real client will implement same interface
5. **Easy Switch** - Just change config flag when ready

## Next Steps

1. **Implement Workers** - Use mock client in provision and deploy workers
2. **Test Full Flow** - Test complete deployment flow with mocks
3. **Implement HTTP Client** - When OpenStack service is ready, implement real HTTP calls
4. **Switch to Real** - Change `USE_MOCK_INFRA=false` to use real service

## Files Created

- `internal/infra/client.go` - Interface and request/response types
- `internal/infra/mock.go` - Mock implementation
- `internal/infra/http.go` - HTTP client stubs (to be implemented)

## Status

✅ **Complete** - Mock client fully implemented with all operations
⏳ **Pending** - HTTP client implementation (Phase 4)
⏳ **Pending** - Worker integration using mock client

