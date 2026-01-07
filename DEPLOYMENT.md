# Click-to-Deploy Deployment Guide

This guide covers deploying Click-to-Deploy in production environments.

> 💡 **New to hosting?** See [HOSTING_RECOMMENDATIONS.md](./docs/HOSTING_RECOMMENDATIONS.md) for provider recommendations and detailed setup guides.

## Quick Start with Docker Compose

The easiest way to get started is using Docker Compose:

### 1. Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 2GB RAM available

### 2. Setup

```bash
# Clone the repository
git clone <repository-url>
cd Click2Deploy

# Create environment file
cp .env.example .env
# Edit .env with your configuration values

# Start services
docker-compose up -d

# Run database migrations
docker-compose exec api migrate -path migrations/postgres \
  -database "postgres://clickdeploy:devpassword@postgres:5432/clickdeploy?sslmode=disable" up

# Check health
curl http://localhost:8080/health
```

### 3. Verify Deployment

```bash
# Check logs
docker-compose logs -f api

# Check health endpoint
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics
```

## Production Deployment

### Option 1: Docker Container

#### Build the Image

```bash
# Build the Docker image
docker build -t click-deploy:latest .

# Or use Makefile
make docker-build
```

#### Run the Container

```bash
docker run -d \
  --name click-deploy \
  -p 8080:8080 \
  --env-file .env \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /tmp/click-deploy-builds:/tmp/click-deploy-builds \
  -v /tmp/prometheus-targets:/tmp/prometheus-targets \
  click-deploy:latest
```

### Option 2: Binary Deployment

#### Build Binary

```bash
# Build for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags='-w -s' \
  -o click-deploy \
  ./cmd/server

# Or use Makefile
make build
```

#### Run Binary

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/clickdeploy?sslmode=disable"
export CASDOOR_ENDPOINT="https://casdoor.example.com"
# ... other environment variables

# Run the server
./click-deploy
```

### Option 3: Kubernetes (Future)

Kubernetes deployment manifests will be added in a future update.

## Environment Configuration

### Required Variables

See `.env.example` for all available configuration options. Minimum required:

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable

# Authentication
CASDOOR_ENDPOINT=https://casdoor.example.com
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret

# OpenStack (or use mock for testing)
INFRA_SERVICE_URL=https://openstack-service.example.com
INFRA_SERVICE_API_KEY=your_api_key
USE_MOCK_INFRA=false  # Set to true for development
```

### Security Considerations

1. **Never commit `.env` files** - Use `.env.example` as a template
2. **Use secrets management** - In production, use:
   - Kubernetes Secrets
   - Docker Secrets
   - HashiCorp Vault
   - AWS Secrets Manager
   - Or similar solutions

3. **Database Security**:
   - Use SSL/TLS for database connections (`sslmode=require`)
   - Use strong passwords
   - Restrict database access to application servers only

4. **API Security**:
   - Enable rate limiting (configured via `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_WINDOW`)
   - Use HTTPS in production (via reverse proxy like Nginx/Caddy)
   - Keep security headers enabled (default)

## Database Migrations

### Using Docker Compose

```bash
# Run migrations
docker-compose exec api migrate -path migrations/postgres \
  -database "postgres://clickdeploy:devpassword@postgres:5432/clickdeploy?sslmode=disable" up

# Rollback last migration
docker-compose exec api migrate -path migrations/postgres \
  -database "postgres://clickdeploy:devpassword@postgres:5432/clickdeploy?sslmode=disable" down
```

### Using Binary

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations/postgres -database "$DATABASE_URL" up

# Rollback
migrate -path migrations/postgres -database "$DATABASE_URL" down
```

## Health Checks

The application exposes a health check endpoint:

```bash
curl http://localhost:8080/health
# Should return: OK
```

Use this endpoint for:
- Load balancer health checks
- Container orchestration (Docker, Kubernetes)
- Monitoring systems

## Monitoring

### Prometheus Metrics

Metrics are exposed at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

Configure Prometheus to scrape this endpoint for monitoring.

### Logs

Application logs are written to stdout/stderr. In production:

- Use a log aggregation service (ELK, Loki, CloudWatch, etc.)
- Configure log rotation
- Set appropriate log levels

## Scaling

### Horizontal Scaling

The application is stateless and can be scaled horizontally:

1. **Database**: Use connection pooling (configured via `DB_MAX_OPEN_CONNS`)
2. **Load Balancer**: Place behind a load balancer (Nginx, HAProxy, AWS ALB)
3. **Multiple Instances**: Run multiple containers/instances behind the load balancer

### Vertical Scaling

Adjust resource limits based on:
- Number of concurrent deployments
- Database size
- Build workload

Recommended minimum:
- **CPU**: 2 cores
- **Memory**: 2GB RAM
- **Storage**: 10GB for builds

## Troubleshooting

### Container Won't Start

1. Check logs: `docker-compose logs api`
2. Verify environment variables: `docker-compose exec api env | grep DATABASE`
3. Check database connectivity
4. Verify all required environment variables are set

### Database Connection Issues

1. Verify `DATABASE_URL` is correct
2. Check database is accessible from container
3. Ensure database user has proper permissions
4. Check firewall rules

### Build Failures

1. Verify BuildKit is accessible (if using Docker socket)
2. Check `BUILD_DIR` has write permissions
3. Verify registry credentials are correct
4. Check network connectivity to registry

## Production Checklist

Before deploying to production:

- [ ] Update all environment variables in `.env`
- [ ] Set `USE_MOCK_INFRA=false` (if using real OpenStack)
- [ ] Configure proper database credentials
- [ ] Set up SSL/TLS certificates
- [ ] Configure reverse proxy (Nginx/Caddy) for HTTPS
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure log aggregation
- [ ] Set up backups for database
- [ ] Configure rate limiting appropriately
- [ ] Review security headers configuration
- [ ] Test health check endpoint
- [ ] Test metrics endpoint
- [ ] Run database migrations
- [ ] Test deployment flow end-to-end

## Support

For issues or questions:
- Check [PROJECT_STATUS.md](./PROJECT_STATUS.md) for current status
- Review [QUICK_START.md](./QUICK_START.md) for development setup
- See [docs/](./docs/) for detailed documentation

