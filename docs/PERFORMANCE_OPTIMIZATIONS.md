# Performance Optimizations

## Overview

This document describes the performance optimizations implemented in Click to Deploy to ensure fast response times and efficient resource usage.

---

## 1. Database Connection Pooling

### Implementation

Database connection pooling is configured in `internal/store/db.go` with optimized settings:

```go
// Default pool configuration
MaxOpenConns:    25  // Maximum open connections
MaxIdleConns:    5  // Idle connections kept ready
ConnMaxLifetime: 300 // 5 minutes - connection reuse limit
ConnMaxIdleTime: 600 // 10 minutes - idle connection timeout
```

### Configuration

Environment variables allow customization:

```bash
DB_MAX_OPEN_CONNS=25      # Maximum open connections
DB_MAX_IDLE_CONNS=5       # Maximum idle connections
DB_CONN_MAX_LIFETIME=300  # Connection lifetime in seconds
```

### Benefits

- **Prevents Connection Exhaustion**: Limits concurrent connections to prevent database overload
- **Faster Response Times**: Maintains a pool of ready connections, avoiding connection setup overhead
- **Resource Efficiency**: Closes idle connections to free database resources
- **Stale Connection Prevention**: Rotates connections periodically to avoid using stale connections

### Best Practices

- **MaxOpenConns**: Set to ~75% of PostgreSQL's `max_connections` setting
- **MaxIdleConns**: Keep small (5-10) to maintain responsiveness without wasting resources
- **ConnMaxLifetime**: 5 minutes is optimal for most deployments
- **ConnMaxIdleTime**: 10 minutes balances resource usage and connection readiness

---

## 2. Response Compression

### Implementation

Gzip compression middleware is implemented in `internal/api/compression.go`:

- Automatically compresses HTTP responses when client supports gzip
- Reduces bandwidth usage for large JSON payloads
- Improves response times for clients on slow connections

### Usage

The compression middleware is automatically applied to all routes:

```go
r.Use(api.CompressionMiddleware)
```

### Benefits

- **Reduced Bandwidth**: JSON responses can be compressed by 70-90%
- **Faster Transfers**: Especially beneficial for large deployment logs and metrics data
- **Better User Experience**: Faster page loads and API responses

### When Compression is Applied

- Client includes `Accept-Encoding: gzip` header
- Response is not already compressed
- Response size is significant (middleware handles this automatically)

---

## 3. Query Optimization

### Indexes

The database schema includes indexes on frequently queried columns:

**Projects:**
- `idx_projects_org` on `casdoor_org_id` - Fast organization-based queries

**Services:**
- `idx_services_project` on `project_id` - Fast project service listing
- `idx_services_status` on `status` - Fast status filtering

**Deployments:**
- `idx_deployments_service` on `service_id` - Fast deployment history
- `idx_deployments_status` on `status` - Fast status filtering
- `idx_deployments_created` on `created_at` - Fast chronological queries

**Jobs:**
- `idx_jobs_status_created` on `(status, created_at)` - Fast job queue queries with SKIP LOCKED

**Databases:**
- `idx_databases_service` on `service_id` - Fast service database lookup
- `idx_databases_hostname` on `internal_hostname` - Fast hostname resolution

### Query Patterns

**Efficient Patterns:**
- Use indexed columns in WHERE clauses
- Use LIMIT for pagination
- Use SKIP LOCKED for job queue (prevents blocking)

**Avoid:**
- Full table scans on large tables
- N+1 query patterns (use JOINs or batch queries)
- Unnecessary ORDER BY on large result sets

---

## 4. Job Queue Optimization

### SKIP LOCKED Pattern

The job queue uses PostgreSQL's `SKIP LOCKED` feature for efficient concurrent processing:

```sql
SELECT * FROM jobs 
WHERE status = 'queued' 
ORDER BY created_at ASC 
FOR UPDATE SKIP LOCKED 
LIMIT 1
```

### Benefits

- **No Blocking**: Workers don't wait for each other
- **High Concurrency**: Multiple workers can process jobs simultaneously
- **Efficient**: No polling or external queue dependencies

---

## 5. API Response Optimization

### Pagination

List endpoints support pagination to limit response sizes:

- `GET /projects` - Paginated project listing
- `GET /services/{id}/deployments` - Paginated deployment history
- `GET /deployments/{id}/logs` - Paginated log retrieval

### Field Selection

Consider adding field selection in the future:
- `GET /services?fields=id,name,status` - Only return requested fields

---

## 6. Caching Strategy (Future)

### Potential Caching Layers

1. **Project/Service Metadata Cache**
   - Cache frequently accessed project and service details
   - Invalidate on updates
   - Use Redis or in-memory cache

2. **Git Repository Cache**
   - Cache repository listings and branch information
   - TTL: 5-10 minutes
   - Invalidate on webhook events

3. **Metrics Cache**
   - Cache Prometheus query results
   - TTL: 30 seconds (metrics update frequently)
   - Reduce load on Prometheus

### Implementation Notes

- Use cache-aside pattern
- Implement cache invalidation on writes
- Monitor cache hit rates
- Use TTL-based expiration

---

## 7. Database Query Optimization

### Prepared Statements

All queries use prepared statements (via `database/sql`):
- Prevents SQL injection
- Allows query plan caching
- Improves performance for repeated queries

### Batch Operations

Where possible, use batch operations:
- `INSERT ... VALUES (...), (...), (...)` for multiple rows
- `UPDATE ... WHERE id IN (...)` for bulk updates

### Connection Reuse

Connection pooling ensures connections are reused:
- Avoids connection setup overhead
- Maintains connection state
- Reduces database load

---

## 8. Monitoring and Metrics

### Performance Metrics

Track the following metrics:
- Database connection pool usage
- Query execution times
- API response times
- Compression ratios
- Cache hit rates (when implemented)

### Prometheus Metrics

Existing metrics help identify performance bottlenecks:
- `click_deploy_service_request_duration_seconds` - API response times
- `click_deploy_service_requests_total` - Request volume
- Database connection pool metrics (to be added)

---

## 9. Best Practices

### Do's

✅ Use connection pooling (already implemented)
✅ Enable response compression (already implemented)
✅ Use indexes on frequently queried columns
✅ Implement pagination for list endpoints
✅ Use SKIP LOCKED for job queue
✅ Monitor database connection pool usage
✅ Use prepared statements

### Don'ts

❌ Don't create unbounded queries (always use LIMIT)
❌ Don't fetch unnecessary data (select only needed columns)
❌ Don't use N+1 query patterns
❌ Don't ignore database connection limits
❌ Don't skip indexes on foreign keys
❌ Don't compress already compressed content

---

## 10. Configuration Recommendations

### Production Settings

```bash
# Database Connection Pool
DB_MAX_OPEN_CONNS=25      # Adjust based on PostgreSQL max_connections
DB_MAX_IDLE_CONNS=5       # Keep small for efficiency
DB_CONN_MAX_LIFETIME=300  # 5 minutes

# PostgreSQL Configuration (recommended)
max_connections = 100      # Allow room for other connections
shared_buffers = 256MB     # Adjust based on available RAM
effective_cache_size = 1GB # Adjust based on available RAM
```

### Development Settings

```bash
# Smaller pool for development
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=2
DB_CONN_MAX_LIFETIME=300
```

---

## 11. Future Optimizations

### Planned

1. **Response Caching**
   - Cache GET responses for immutable resources
   - Implement ETag support
   - Add cache-control headers

2. **Database Query Result Caching**
   - Cache frequently accessed data
   - Use Redis for distributed caching
   - Implement cache invalidation strategies

3. **Connection Pool Monitoring**
   - Add Prometheus metrics for pool usage
   - Alert on connection exhaustion
   - Auto-scale pool size based on load

4. **Query Performance Monitoring**
   - Log slow queries
   - Track query execution times
   - Identify optimization opportunities

5. **Batch API Endpoints**
   - Support bulk operations
   - Reduce API round trips
   - Improve efficiency for bulk updates

---

## Summary

Current optimizations provide:
- ✅ Efficient database connection management
- ✅ Reduced bandwidth usage via compression
- ✅ Fast job queue processing
- ✅ Optimized query patterns with indexes

Future optimizations will add:
- ⏳ Response caching
- ⏳ Query result caching
- ⏳ Performance monitoring dashboards
- ⏳ Advanced query optimization

