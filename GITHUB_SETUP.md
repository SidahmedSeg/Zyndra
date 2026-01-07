# GitHub Setup for Railway Deployment

Quick guide to push your code to GitHub before deploying to Railway.

## Step 1: Initialize Git Repository

If you haven't initialized git yet:

```bash
# Initialize git repository
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Click-to-Deploy application"
```

## Step 2: Create GitHub Repository

1. Go to https://github.com/new
2. Repository name: `click-to-deploy` (or your preferred name)
3. Description: "Platform-as-a-Service (PaaS) for OpenStack"
4. Choose **Public** or **Private**
5. **Don't** initialize with README, .gitignore, or license (we already have these)
6. Click **"Create repository"**

## Step 3: Push to GitHub

```bash
# Add GitHub remote (replace YOUR_USERNAME with your GitHub username)
git remote add origin https://github.com/YOUR_USERNAME/click-to-deploy.git

# Rename branch to main (if needed)
git branch -M main

# Push to GitHub
git push -u origin main
```

## Step 4: Verify

1. Go to your GitHub repository
2. Verify all files are there
3. Check that `.env` is **NOT** in the repository (it's in `.gitignore`)

## Step 5: Ready for Railway!

Now you can follow the Railway setup guide:
- Quick start: [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md)
- Full guide: [docs/RAILWAY_SETUP.md](./docs/RAILWAY_SETUP.md)

## Important: What's Excluded

The `.gitignore` file ensures these are **NOT** pushed to GitHub:
- `.env` files (environment variables)
- Binary files (`click-deploy`, `server`)
- Build artifacts
- Database files
- Node modules
- IDE files

**Never commit:**
- `.env` files with secrets
- Database passwords
- API keys
- OAuth secrets

## Troubleshooting

### "Repository not found"

- Check your GitHub username is correct
- Verify repository exists on GitHub
- Ensure you have push access

### "Permission denied"

- Use HTTPS with personal access token, or
- Set up SSH keys: https://docs.github.com/en/authentication/connecting-to-github-with-ssh

### "Large files" error

- Check `.gitignore` is working
- Remove large files: `git rm --cached large-file`
- Commit again

