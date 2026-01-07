# Railway Deployment Guide

Step-by-step guide to deploy Click-to-Deploy on Railway.

## Prerequisites

1. **GitHub Account** - Your code needs to be on GitHub
2. **Railway Account** - Sign up at https://railway.app (free to start)

## Step 1: Push Code to GitHub

If your code isn't on GitHub yet:

```bash
# 1. Initialize git (if not already done)
git init

# 2. Add all files
git add .

# 3. Commit
git commit -m "Initial commit: Click-to-Deploy application"

# 4. Create a new repository on GitHub
# Go to https://github.com/new
# Create a new repository (e.g., "click-to-deploy")

# 5. Add remote and push
git remote add origin https://github.com/YOUR_USERNAME/click-to-deploy.git
git branch -M main
git push -u origin main
```

**Important:** Make sure `.env` is in `.gitignore` (it should be already)

## Step 2: Sign Up for Railway

1. Go to https://railway.app
2. Click "Start a New Project"
3. Sign up with GitHub (recommended) or email
4. Authorize Railway to access your GitHub repositories

## Step 3: Create New Project

1. In Railway dashboard, click **"New Project"**
2. Select **"Deploy from GitHub repo"**
3. Choose your `click-to-deploy` repository
4. Railway will create a new project

## Step 4: Add PostgreSQL Database

1. In your Railway project, click **"+ New"**
2. Select **"Database"**
3. Choose **"PostgreSQL"**
4. Railway will automatically:
   - Create the database
   - Set `DATABASE_URL` environment variable
   - Link it to your service

## Step 5: Configure Your Service

1. Click on your service (the one that was auto-created from GitHub)
2. Go to **"Settings"** tab
3. Configure:

   **Build Settings:**
   - **Root Directory:** `/` (default)
   - **Dockerfile Path:** `Dockerfile` (should auto-detect)
   - **Build Command:** (leave empty, Dockerfile handles it)

   **Deploy Settings:**
   - **Start Command:** `./click-deploy` (from Dockerfile)
   - **Healthcheck Path:** `/health`

## Step 6: Set Environment Variables

Go to **"Variables"** tab and add these:

### Required Variables

```bash
# Server
PORT=8080
ENVIRONMENT=production

# Database (auto-set by Railway, but verify it exists)
# DATABASE_URL is automatically set by Railway when you add PostgreSQL

# Casdoor Authentication
CASDOOR_ENDPOINT=https://casdoor.example.com
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret

# OpenStack Infrastructure
INFRA_SERVICE_URL=https://openstack-service.example.com
INFRA_SERVICE_API_KEY=your_api_key
USE_MOCK_INFRA=true  # Set to false when ready for production

# Container Registry
REGISTRY_URL=https://registry.example.com
REGISTRY_USERNAME=admin
REGISTRY_PASSWORD=password

# BuildKit (if using)
BUILDKIT_ADDRESS=unix:///run/buildkit/buildkitd.sock
BUILD_DIR=/tmp/click-deploy-builds

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=https://YOUR_APP.railway.app/git/callback/github

# GitLab OAuth (if using)
GITLAB_CLIENT_ID=your_gitlab_client_id
GITLAB_CLIENT_SECRET=your_gitlab_client_secret
GITLAB_REDIRECT_URL=https://YOUR_APP.railway.app/git/callback/gitlab

# Webhook
WEBHOOK_SECRET=your_webhook_secret
BASE_URL=https://YOUR_APP.railway.app

# DNS (for database internal hostnames)
DNS_ZONE_ID=your_dns_zone_id

# Caddy (for custom domains)
CADDY_ADMIN_URL=http://localhost:2019

# Prometheus
PROMETHEUS_URL=http://localhost:9090
PROMETHEUS_TARGETS_DIR=/tmp/prometheus-targets

# Centrifugo (for real-time logs)
CENTRIFUGO_API_URL=http://localhost:8000
CENTRIFUGO_API_KEY=your_centrifugo_api_key

# Security (optional, defaults provided)
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# Database Connection Pool (optional)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300
```

**Important:** Replace `YOUR_APP.railway.app` with your actual Railway domain (you'll see it after first deployment).

## Step 7: Deploy

1. Railway automatically deploys when you push to GitHub
2. Or click **"Deploy"** button in Railway dashboard
3. Watch the build logs in real-time
4. Wait for deployment to complete

## Step 8: Run Database Migrations

After first deployment, run migrations:

### Option 1: Using Railway CLI

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Link to your project
railway link

# Run migrations
railway run migrate -path migrations/postgres -database "$DATABASE_URL" up
```

### Option 2: Using Railway Dashboard

1. Go to your service
2. Click **"Deployments"** tab
3. Click on the latest deployment
4. Click **"View Logs"**
5. Use the **"Shell"** option (if available) to run commands

### Option 3: Using Railway Database Query

1. Go to your PostgreSQL service
2. Click **"Query"** tab
3. Manually run SQL from `migrations/postgres/001_initial.up.sql`
4. Then run `migrations/postgres/002_tables.up.sql`

## Step 9: Verify Deployment

1. **Check Health:**
   ```bash
   curl https://YOUR_APP.railway.app/health
   # Should return: OK
   ```

2. **Check Metrics:**
   ```bash
   curl https://YOUR_APP.railway.app/metrics
   ```

3. **Check Logs:**
   - Go to Railway dashboard → Your service → **"Deployments"** → **"View Logs"**

## Step 10: Configure Custom Domain (Optional)

1. Go to your service → **"Settings"** → **"Networking"**
2. Click **"Generate Domain"** (if not auto-generated)
3. Or add your custom domain:
   - Click **"Custom Domain"**
   - Enter your domain (e.g., `api.yourdomain.com`)
   - Railway will provide DNS instructions
   - Update your DNS records as instructed
   - Railway automatically provisions SSL certificate

## Troubleshooting

### Build Fails

1. **Check Dockerfile:**
   - Ensure Dockerfile is in root directory
   - Verify all paths are correct

2. **Check Build Logs:**
   - Railway dashboard → Service → Deployments → View Logs
   - Look for error messages

3. **Common Issues:**
   - Missing dependencies in `go.mod`
   - Incorrect Dockerfile paths
   - Build timeout (increase in settings)

### Application Won't Start

1. **Check Environment Variables:**
   - Verify all required variables are set
   - Check `DATABASE_URL` is correctly linked

2. **Check Application Logs:**
   - Railway dashboard → Service → Deployments → View Logs
   - Look for startup errors

3. **Common Issues:**
   - Database connection failures
   - Missing environment variables
   - Port conflicts (should be 8080)

### Database Connection Issues

1. **Verify DATABASE_URL:**
   - Railway automatically sets this when you add PostgreSQL
   - Check in Variables tab
   - Format: `postgres://user:password@host:port/dbname`

2. **Check Database Status:**
   - Go to PostgreSQL service
   - Verify it's running
   - Check connection string

3. **Run Migrations:**
   - Ensure migrations are run before using the app

### Health Check Fails

1. **Verify Health Endpoint:**
   - Check `/health` endpoint is accessible
   - Should return "OK"

2. **Check Port:**
   - Ensure application listens on `PORT` environment variable
   - Railway sets this automatically

## Railway-Specific Tips

1. **Auto-Deploy:**
   - Railway automatically deploys on every push to main branch
   - You can disable this in Settings → Source

2. **Environment Variables:**
   - Use Railway's variable reference: `${{Postgres.DATABASE_URL}}`
   - Or just use the auto-linked `DATABASE_URL`

3. **Logs:**
   - Access logs in real-time from Railway dashboard
   - Logs are retained for 7 days (free tier)

4. **Scaling:**
   - Railway auto-scales based on traffic
   - You can set manual limits in Settings

5. **Costs:**
   - Free tier: $5 credit/month
   - Pay-as-you-go after that
   - Monitor usage in dashboard

## Next Steps

After successful deployment:

1. ✅ Test all API endpoints
2. ✅ Set up monitoring (Prometheus metrics)
3. ✅ Configure custom domain (if needed)
4. ✅ Set up CI/CD (optional, Railway auto-deploys)
5. ✅ Configure backups for database
6. ✅ Set up alerts

## Support

- Railway Docs: https://docs.railway.app
- Railway Discord: https://discord.gg/railway
- Railway Status: https://status.railway.app

