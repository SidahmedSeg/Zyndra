# Hosting Provider Recommendations

This document provides recommendations for hosting Click-to-Deploy, starting with small-scale deployments and scaling options.

## 🏆 Top Recommendations for Starting Small

### 1. **Railway** (Recommended for Quick Start) ⭐

**Best for:** Getting started quickly with minimal configuration

**Pros:**
- ✅ One-click PostgreSQL database
- ✅ Automatic Docker builds from Git
- ✅ Built-in environment variable management
- ✅ Free tier available ($5 credit/month)
- ✅ Automatic HTTPS/SSL
- ✅ Simple deployment process
- ✅ Good for Next.js frontend too

**Cons:**
- ⚠️ Can get expensive at scale
- ⚠️ Less control over infrastructure

**Setup:**
1. Connect GitHub repository
2. Add PostgreSQL service
3. Set environment variables
4. Deploy!

**Cost:** ~$5-20/month for small deployments

**Link:** https://railway.app

---

### 2. **Render** (Great Balance) ⭐

**Best for:** Professional deployments with good developer experience

**Pros:**
- ✅ Free tier available (with limitations)
- ✅ Managed PostgreSQL included
- ✅ Docker support
- ✅ Automatic SSL certificates
- ✅ Easy environment variable management
- ✅ Good documentation
- ✅ Supports both backend and frontend

**Cons:**
- ⚠️ Free tier spins down after inactivity
- ⚠️ Can be slower than dedicated VPS

**Setup:**
1. Connect GitHub repository
2. Create PostgreSQL database
3. Create Web Service (Docker)
4. Configure environment variables

**Cost:** ~$7-25/month for small deployments

**Link:** https://render.com

---

### 3. **DigitalOcean App Platform** (Best for Growth)

**Best for:** When you need more control and predictable pricing

**Pros:**
- ✅ Managed PostgreSQL available
- ✅ Docker support
- ✅ Automatic scaling
- ✅ Good performance
- ✅ Predictable pricing
- ✅ Great documentation
- ✅ Can add managed databases easily

**Cons:**
- ⚠️ Slightly more complex setup
- ⚠️ No free tier (but affordable)

**Setup:**
1. Connect GitHub repository
2. Create App (Docker)
3. Add managed PostgreSQL database
4. Configure environment variables

**Cost:** ~$12-30/month for small deployments

**Link:** https://www.digitalocean.com/products/app-platform

---

### 4. **Fly.io** (Best for Docker)

**Best for:** Global distribution and Docker-first approach

**Pros:**
- ✅ Excellent Docker support
- ✅ Global edge deployment
- ✅ Free tier (3 shared VMs)
- ✅ Great for containerized apps
- ✅ Fast deployments
- ✅ Built-in health checks

**Cons:**
- ⚠️ PostgreSQL requires separate setup (or use Supabase)
- ⚠️ Learning curve for flyctl CLI

**Setup:**
1. Install `flyctl`
2. Run `fly launch`
3. Configure PostgreSQL (external or Supabase)
4. Deploy with `fly deploy`

**Cost:** Free tier available, ~$5-15/month for small deployments

**Link:** https://fly.io

---

## 🎯 Recommendation by Use Case

### **Just Starting / MVP**
→ **Railway** or **Render**
- Fastest to get running
- Minimal configuration
- Free tier to test

### **Production with Growth Plans**
→ **DigitalOcean App Platform**
- Better for scaling
- More predictable costs
- Professional features

### **Docker-First / Global Distribution**
→ **Fly.io**
- Best Docker experience
- Global edge network
- Great for containerized apps

### **Budget-Conscious / Self-Hosting**
→ **Hetzner Cloud** or **Vultr**
- VPS with Docker Compose
- Full control
- Very affordable (~$5-10/month)

---

## 📋 Setup Guide for Railway (Recommended)

### Step 1: Create Railway Account
1. Go to https://railway.app
2. Sign up with GitHub

### Step 2: Create New Project
1. Click "New Project"
2. Select "Deploy from GitHub repo"
3. Choose your Click2Deploy repository

### Step 3: Add PostgreSQL
1. Click "+ New" → "Database" → "PostgreSQL"
2. Railway automatically creates database and sets `DATABASE_URL`

### Step 4: Configure Environment Variables
Add these in Railway dashboard:

```bash
# Server
PORT=8080
ENVIRONMENT=production

# Database (auto-set by Railway, but verify)
DATABASE_URL=${{Postgres.DATABASE_URL}}

# Casdoor
CASDOOR_ENDPOINT=https://casdoor.example.com
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret

# OpenStack (or use mock)
USE_MOCK_INFRA=true  # Set to false when ready
INFRA_SERVICE_URL=https://openstack-service.example.com
INFRA_SERVICE_API_KEY=your_api_key

# Registry
REGISTRY_URL=https://registry.example.com
REGISTRY_USERNAME=admin
REGISTRY_PASSWORD=password

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=https://your-app.railway.app/git/callback/github

# Webhook
WEBHOOK_SECRET=your_webhook_secret
BASE_URL=https://your-app.railway.app

# Other services...
```

### Step 5: Configure Build Settings
- **Root Directory:** `/` (or leave default)
- **Dockerfile Path:** `Dockerfile`
- **Start Command:** (auto-detected from Dockerfile)

### Step 6: Deploy
1. Railway automatically builds and deploys
2. Check logs in Railway dashboard
3. Access your app at `https://your-app.railway.app`

### Step 7: Run Migrations
```bash
# Using Railway CLI
railway run migrate -path migrations/postgres -database "$DATABASE_URL" up

# Or via Railway dashboard → Database → Query
```

---

## 📋 Setup Guide for Render

### Step 1: Create Render Account
1. Go to https://render.com
2. Sign up with GitHub

### Step 2: Create PostgreSQL Database
1. Dashboard → "New +" → "PostgreSQL"
2. Choose plan (Free tier available)
3. Note the connection string

### Step 3: Create Web Service
1. Dashboard → "New +" → "Web Service"
2. Connect GitHub repository
3. Configure:
   - **Name:** click-deploy-api
   - **Environment:** Docker
   - **Region:** Choose closest
   - **Branch:** main
   - **Root Directory:** `/`
   - **Dockerfile Path:** `Dockerfile`

### Step 4: Environment Variables
Add all required variables (see Railway guide above)

### Step 5: Deploy
1. Click "Create Web Service"
2. Render builds and deploys automatically
3. Access at `https://your-app.onrender.com`

---

## 📋 Setup Guide for DigitalOcean App Platform

### Step 1: Create DigitalOcean Account
1. Go to https://www.digitalocean.com
2. Sign up (get $200 credit)

### Step 2: Create App
1. Go to App Platform
2. "Create App" → "GitHub"
3. Select repository

### Step 3: Configure App
1. **Source:** GitHub repository
2. **Type:** Docker
3. **Dockerfile Path:** `Dockerfile`
4. **Run Command:** (auto-detected)

### Step 4: Add Database
1. "Add Resource" → "Database" → "PostgreSQL"
2. Choose plan (Basic $15/month minimum)
3. Database automatically linked

### Step 5: Environment Variables
Add all required variables

### Step 6: Deploy
1. Review and create
2. DigitalOcean builds and deploys
3. Access at `https://your-app.ondigitalocean.app`

---

## 🔧 VPS Option (Hetzner/Vultr) - For Full Control

If you prefer self-hosting on a VPS:

### Recommended Setup:
- **VPS:** Hetzner CPX11 (€4.51/month) or Vultr ($6/month)
- **OS:** Ubuntu 22.04 LTS
- **Deployment:** Docker Compose

### Setup Steps:
```bash
# 1. SSH into VPS
ssh root@your-server-ip

# 2. Install Docker & Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 3. Clone repository
git clone <your-repo-url>
cd Click2Deploy

# 4. Create .env file
cp .env.example .env
nano .env  # Edit with your values

# 5. Start services
docker-compose up -d

# 6. Run migrations
docker-compose exec api migrate -path migrations/postgres \
  -database "postgres://clickdeploy:password@postgres:5432/clickdeploy?sslmode=disable" up

# 7. Set up reverse proxy (Nginx/Caddy) for HTTPS
```

**Pros:**
- Full control
- Very affordable
- Can host multiple services
- Good for learning

**Cons:**
- Manual setup required
- You manage updates/security
- No automatic scaling

---

## 💰 Cost Comparison (Small Deployment)

| Provider | Starting Cost | Database | Notes |
|----------|--------------|----------|-------|
| **Railway** | $5-20/mo | Included | Free tier available |
| **Render** | $7-25/mo | Included | Free tier (spins down) |
| **DigitalOcean** | $12-30/mo | $15/mo extra | Most predictable |
| **Fly.io** | $5-15/mo | External | Free tier available |
| **Hetzner VPS** | €4.51/mo | Self-hosted | Full control |
| **Vultr VPS** | $6/mo | Self-hosted | Full control |

---

## 🚀 My Recommendation: Start with Railway

**Why Railway for starting:**
1. ✅ **Fastest setup** - Can be running in 10 minutes
2. ✅ **All-in-one** - Database included, no separate setup
3. ✅ **Free tier** - Test before committing
4. ✅ **Automatic HTTPS** - No SSL configuration needed
5. ✅ **Git-based** - Auto-deploy on push
6. ✅ **Good for Next.js** - Can host frontend too

**Migration path:**
- Start on Railway (MVP/testing)
- Move to DigitalOcean App Platform (production)
- Or scale to Kubernetes (enterprise)

---

## 📝 Next Steps After Choosing Provider

1. **Set up the provider** (follow guide above)
2. **Configure environment variables**
3. **Run database migrations**
4. **Test health endpoint**
5. **Set up monitoring** (Prometheus metrics)
6. **Configure custom domain** (if needed)
7. **Set up CI/CD** (GitHub Actions for auto-deploy)

---

## 🔗 Useful Links

- Railway: https://railway.app
- Render: https://render.com
- DigitalOcean: https://www.digitalocean.com/products/app-platform
- Fly.io: https://fly.io
- Hetzner: https://www.hetzner.com/cloud
- Vultr: https://www.vultr.com

---

## 💡 Tips for Production

1. **Always use managed databases** - Don't self-host PostgreSQL in production initially
2. **Enable backups** - Most providers offer automatic backups
3. **Set up monitoring** - Use Prometheus metrics endpoint
4. **Use environment-specific configs** - Separate dev/staging/prod
5. **Enable rate limiting** - Already configured in the app
6. **Use HTTPS** - Most providers handle this automatically
7. **Set up alerts** - Monitor health checks and errors

