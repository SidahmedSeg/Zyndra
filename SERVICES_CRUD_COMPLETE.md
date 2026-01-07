# ✅ Services CRUD Implementation Complete

## What Was Implemented

### 1. Store Layer (`internal/store/services.go`)
- ✅ `CreateService()` - Create new service
- ✅ `GetService()` - Get service by ID
- ✅ `ListServicesByProject()` - List all services in a project
- ✅ `UpdateService()` - Update service configuration
- ✅ `UpdateServicePosition()` - Update canvas position
- ✅ `DeleteService()` - Delete service (with cascade)
- ✅ `ServiceExists()` - Check if service exists
- ✅ `GetProjectIDForService()` - Get project ID for a service

### 2. API Handlers (`internal/api/services.go`)
- ✅ `ListServices()` - GET /projects/:id/services
- ✅ `CreateService()` - POST /projects/:id/services
- ✅ `GetService()` - GET /services/:id
- ✅ `UpdateService()` - PATCH /services/:id
- ✅ `UpdateServicePosition()` - PATCH /services/:id/position
- ✅ `DeleteService()` - DELETE /services/:id

### 3. Request Models (`internal/api/services_requests.go`)
- ✅ `CreateServiceRequest` - Request body for creating services
- ✅ `UpdateServiceRequest` - Request body for updating services
- ✅ `UpdateServicePositionRequest` - Request body for position updates

### 4. Validation
- ✅ Service name validation
- ✅ Service type validation (app, database, volume)
- ✅ Instance size validation (small, medium, large, xlarge)
- ✅ Port range validation (1-65535)
- ✅ Canvas position validation
- ✅ Project ownership verification

## API Endpoints

### GET /v1/click-deploy/projects/:id/services
**List all services in a project**

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "660e8400-e29b-41d4-a716-446655440001",
    "git_source_id": null,
    "name": "backend",
    "type": "app",
    "status": "pending",
    "instance_size": "medium",
    "port": 8080,
    "openstack_instance_id": null,
    "openstack_fip_id": null,
    "openstack_fip_address": null,
    "security_group_id": null,
    "subdomain": null,
    "generated_url": null,
    "current_image_tag": null,
    "canvas_x": 100,
    "canvas_y": 200,
    "created_at": "2026-01-06T12:00:00Z",
    "updated_at": "2026-01-06T12:00:00Z"
  }
]
```

### POST /v1/click-deploy/projects/:id/services
**Create a new service**

**Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "backend",
  "type": "app",
  "instance_size": "medium",
  "port": 8080,
  "git_source_id": "770e8400-e29b-41d4-a716-446655440002",
  "canvas_x": 100,
  "canvas_y": 200
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "project_id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "backend",
  "type": "app",
  "status": "pending",
  "instance_size": "medium",
  "port": 8080,
  "canvas_x": 100,
  "canvas_y": 200,
  ...
}
```

### GET /v1/click-deploy/services/:id
**Get a specific service**

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "project_id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "backend",
  "type": "app",
  "status": "live",
  ...
}
```

### PATCH /v1/click-deploy/services/:id
**Update a service**

**Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body (all fields optional):**
```json
{
  "name": "backend-api",
  "instance_size": "large",
  "port": 3000,
  "status": "live"
}
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "backend-api",
  "instance_size": "large",
  "port": 3000,
  "status": "live",
  "updated_at": "2026-01-06T13:00:00Z",
  ...
}
```

### PATCH /v1/click-deploy/services/:id/position
**Update canvas position**

**Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "x": 250,
  "y": 350
}
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "canvas_x": 250,
  "canvas_y": 350,
  "updated_at": "2026-01-06T13:00:00Z",
  ...
}
```

### DELETE /v1/click-deploy/services/:id
**Delete a service and all its resources**

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:** `204 No Content`

## Service Types

### App Service
- Type: `"app"`
- Represents an application deployed from Git repository
- Gets floating IP and public URL
- Can be linked to databases via environment variables

### Database Service
- Type: `"database"`
- Represents a managed database (PostgreSQL, MySQL, Redis)
- Gets internal hostname only (no public access)
- Auto-creates persistent volume

### Volume Service
- Type: `"volume"`
- Represents persistent storage
- Can be attached to app services or databases

## Service Statuses

- `pending` - Service created but not yet provisioned
- `provisioning` - Infrastructure being provisioned
- `building` - Container image being built
- `deploying` - Service being deployed
- `live` - Service is running and healthy
- `failed` - Deployment or health check failed
- `stopped` - Service manually stopped

## Instance Sizes

- `small` - 1 vCPU, 512MB RAM
- `medium` - 2 vCPU, 2GB RAM (default)
- `large` - 4 vCPU, 4GB RAM
- `xlarge` - 8 vCPU, 8GB RAM

## Features

### Canvas Position Management
- Services can be positioned on the canvas
- Position stored as `canvas_x` and `canvas_y` coordinates
- Updated via dedicated endpoint

### Project Scoping
- All services belong to a project
- Services are listed by project
- Project ownership verified on all operations

### Cascade Deletion
- Deleting a service automatically deletes:
  - All deployments
  - All deployment logs
  - All environment variables
  - All custom domains
  - Git source references

### Organization Isolation
- All operations verify service belongs to authenticated organization
- Users can only access/modify their organization's services
- Prevents cross-organization access

## Testing Examples

### Create Service
```bash
curl -X POST http://localhost:8080/v1/click-deploy/projects/{project-id}/services \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "backend",
    "type": "app",
    "instance_size": "medium",
    "port": 8080,
    "canvas_x": 100,
    "canvas_y": 200
  }'
```

### List Services
```bash
curl http://localhost:8080/v1/click-deploy/projects/{project-id}/services \
  -H "Authorization: Bearer <jwt-token>"
```

### Get Service
```bash
curl http://localhost:8080/v1/click-deploy/services/{service-id} \
  -H "Authorization: Bearer <jwt-token>"
```

### Update Service
```bash
curl -X PATCH http://localhost:8080/v1/click-deploy/services/{service-id} \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "backend-api",
    "instance_size": "large"
  }'
```

### Update Position
```bash
curl -X PATCH http://localhost:8080/v1/click-deploy/services/{service-id}/position \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "x": 250,
    "y": 350
  }'
```

### Delete Service
```bash
curl -X DELETE http://localhost:8080/v1/click-deploy/services/{service-id} \
  -H "Authorization: Bearer <jwt-token>"
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Validation error: name is required"
}
```

### 401 Unauthorized
```json
{
  "error": "Organization ID not found in token"
}
```

### 404 Not Found
```json
{
  "error": "Service not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to create service: <error details>"
}
```

## Security Features

- ✅ Organization isolation - users can only access their org's services
- ✅ Project ownership verification
- ✅ Authentication required for all operations
- ✅ Input validation and sanitization
- ✅ SQL injection prevention (parameterized queries)
- ✅ Cascade deletion prevents orphaned resources

## Next Steps

1. ✅ Services CRUD - **COMPLETE**
2. ⏳ Add comprehensive error handling
3. ⏳ Implement Deployments API
4. ⏳ Implement Environment Variables API
5. ⏳ Add request validation library (validator.v10)

---

**✅ Services CRUD is fully functional and ready for use!**

