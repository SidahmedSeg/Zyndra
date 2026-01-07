# Railway Quick Start - 5 Minutes

Get Click-to-Deploy running on Railway in 5 minutes!

## ✅ Prerequisites Checklist

- [ ] Code is pushed to GitHub
- [ ] Railway account created (https://railway.app)
- [ ] GitHub OAuth app created (for Git integration)

## 🚀 Quick Deploy Steps

### 1. Push to GitHub (if not already)

```bash
git add .
git commit -m "Ready for Railway deployment"
git push origin main
```

### 2. Create Railway Project

1. Go to https://railway.app
2. Click **"New Project"**
3. Select **"Deploy from GitHub repo"**
4. Choose your repository

### 3. Add PostgreSQL

1. Click **"+ New"** → **"Database"** → **"PostgreSQL"**
2. Done! Railway sets `DATABASE_URL` automatically

### 4. Set Minimum Required Variables

Go to your service → **"Variables"** tab, add:

```bash
# Replace with your actual values
CASDOOR_ENDPOINT=https://casdoor.example.com
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret

# For testing, use mock OpenStack
USE_MOCK_INFRA=true

# GitHub OAuth (update after first deploy with your Railway URL)
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=https://YOUR_APP.railway.app/git/callback/github

# Webhook
WEBHOOK_SECRET=generate-a-random-secret
BASE_URL=https://YOUR_APP.railway.app
```

### 5. Deploy & Run Migrations

1. Railway auto-deploys on push, or click **"Deploy"**
2. Wait for build to complete
3. Run migrations:

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login and link
railway login
railway link

# Run migrations
railway run migrate -path migrations/postgres -database "$DATABASE_URL" up
```

### 6. Test

```bash
curl https://YOUR_APP.railway.app/health
# Should return: OK
```

## 🎉 Done!

Your app is now live at `https://YOUR_APP.railway.app`

## 📝 Next Steps

1. Update `GITHUB_REDIRECT_URL` with your actual Railway URL
2. Add remaining environment variables (see full guide)
3. Test API endpoints
4. Configure custom domain (optional)

## 📚 Full Guide

See [docs/RAILWAY_SETUP.md](./docs/RAILWAY_SETUP.md) for detailed instructions.

