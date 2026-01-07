# Click-to-Deploy

Platform-as-a-Service (PaaS) module for OpenStack-based cloud platform.

## Quick Start

### Option 1: Railway Deployment (Recommended for Production)

Deploy to Railway in 5 minutes:

```bash
# 1. Push to GitHub (if not already)
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/YOUR_USERNAME/click-to-deploy.git
git push -u origin main

# 2. Follow Railway setup guide
# See RAILWAY_QUICKSTART.md for step-by-step instructions
```

See [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md) for the quick start guide, or [docs/RAILWAY_SETUP.md](./docs/RAILWAY_SETUP.md) for detailed instructions.

### Option 2: Docker Compose (Local Development)

The fastest way to get started locally:

```bash
# 1. Create environment file
cp .env.example .env
# Edit .env with your configuration

# 2. Start all services
docker-compose up -d

# 3. Run database migrations
docker-compose exec api migrate -path migrations/postgres \
  -database "postgres://clickdeploy:devpassword@postgres:5432/clickdeploy?sslmode=disable" up

# 4. Check health
curl http://localhost:8080/health
```

See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed deployment instructions.

### Option 2: Local Development

#### Prerequisites

- Go 1.22+
- PostgreSQL (or use Docker)
- Docker (optional, for local database)

#### Setup

1. **Install dependencies:**
   ```bash
   make install-deps
   ```

2. **Set up database:**
   ```bash
   # Using Docker
   docker run -d \
     --name click-deploy-db \
     -e POSTGRES_USER=clickdeploy \
     -e POSTGRES_PASSWORD=devpassword \
     -e POSTGRES_DB=clickdeploy \
     -p 5432:5432 \
     postgres:15
   ```

3. **Create .env file:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run migrations:**
   ```bash
   export DATABASE_URL="postgres://clickdeploy:devpassword@localhost:5432/clickdeploy?sslmode=disable"
   make migrate-up
   ```

5. **Start server:**
   ```bash
   make run
   ```

6. **Test:**
   ```bash
   curl http://localhost:8080/health
   ```

## Project Structure

```
click-deploy/
├── cmd/server/          # Application entry point
├── internal/            # Internal packages
│   ├── api/            # HTTP handlers
│   ├── auth/           # Authentication
│   ├── config/          # Configuration
│   ├── store/           # Database layer
│   └── ...
├── migrations/          # Database migrations
└── web/                 # Frontend (React)
```

## Development

See [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) for the full development roadmap.

## Documentation

- [Deployment Guide](./DEPLOYMENT.md) - Production deployment instructions
- [Development Plan](./DEVELOPMENT_PLAN.md) - Full development roadmap
- [Quick Start Guide](./QUICK_START.md) - Detailed setup instructions
- [Project Status](./PROJECT_STATUS.md) - Current project status and progress
- [Specification Analysis](./SPECIFICATION_ANALYSIS.md) - Technical analysis

## Production Deployment

For production deployment, see [DEPLOYMENT.md](./DEPLOYMENT.md) which covers:

- Docker Compose setup
- Docker container deployment
- Binary deployment
- Environment configuration
- Database migrations
- Health checks and monitoring
- Scaling strategies
- Production checklist

