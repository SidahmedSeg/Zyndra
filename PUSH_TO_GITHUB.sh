#!/bin/bash
# Script to push Click-to-Deploy to GitHub
# Run this script after creating a GitHub repository

set -e

echo "🚀 Click-to-Deploy GitHub Push Script"
echo "======================================"
echo ""

# Check if git is initialized
if [ ! -d .git ]; then
    echo "📦 Initializing git repository..."
    git init
    echo "✅ Git initialized"
else
    echo "✅ Git already initialized"
fi

# Check if remote exists
if git remote | grep -q origin; then
    echo "⚠️  Remote 'origin' already exists"
    echo "Current remote:"
    git remote -v
    echo ""
    read -p "Do you want to update it? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter your GitHub username: " GITHUB_USER
        read -p "Enter your repository name (default: click-to-deploy): " REPO_NAME
        REPO_NAME=${REPO_NAME:-click-to-deploy}
        git remote set-url origin "https://github.com/${GITHUB_USER}/${REPO_NAME}.git"
        echo "✅ Remote updated"
    fi
else
    echo "📝 Setting up GitHub remote..."
    read -p "Enter your GitHub username: " GITHUB_USER
    read -p "Enter your repository name (default: click-to-deploy): " REPO_NAME
    REPO_NAME=${REPO_NAME:-click-to-deploy}
    git remote add origin "https://github.com/${GITHUB_USER}/${REPO_NAME}.git"
    echo "✅ Remote added: https://github.com/${GITHUB_USER}/${REPO_NAME}.git"
fi

echo ""
echo "📋 Checking files to commit..."
git add .

echo ""
echo "📝 Files staged. Review what will be committed:"
git status --short

echo ""
read -p "Continue with commit? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cancelled"
    exit 1
fi

# Commit
echo ""
echo "💾 Creating commit..."
git commit -m "Initial commit: Click-to-Deploy application" || {
    echo "⚠️  No changes to commit (or commit failed)"
    echo "This is OK if you've already committed everything"
}

# Check branch
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "main")
if [ -z "$CURRENT_BRANCH" ]; then
    echo "🌿 Creating main branch..."
    git checkout -b main
    CURRENT_BRANCH="main"
fi

echo ""
echo "🚀 Pushing to GitHub..."
echo "Branch: $CURRENT_BRANCH"
echo ""

# Push
if git push -u origin "$CURRENT_BRANCH" 2>&1; then
    echo ""
    echo "✅ Successfully pushed to GitHub!"
    echo ""
    echo "🔗 Repository: https://github.com/${GITHUB_USER}/${REPO_NAME}"
    echo ""
    echo "📝 Next steps:"
    echo "1. Go to https://railway.app"
    echo "2. Create new project"
    echo "3. Deploy from GitHub repo"
    echo "4. Select: ${GITHUB_USER}/${REPO_NAME}"
else
    echo ""
    echo "❌ Push failed!"
    echo ""
    echo "Common issues:"
    echo "1. Repository doesn't exist on GitHub - create it first at https://github.com/new"
    echo "2. Authentication required - GitHub may prompt for credentials"
    echo "3. Permission denied - check your GitHub access"
    echo ""
    echo "💡 Tip: Use GitHub CLI (gh) or set up SSH keys for easier authentication"
    exit 1
fi

