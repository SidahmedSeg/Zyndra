#!/bin/bash
# Helper script to push with GitHub token
# Usage: ./push_with_token.sh <github_repo_url> <access_token>

set -e

REPO_URL=$1
TOKEN=$2

if [ -z "$REPO_URL" ] || [ -z "$TOKEN" ]; then
    echo "Usage: ./push_with_token.sh <github_repo_url> <access_token>"
    echo "Example: ./push_with_token.sh https://github.com/username/repo.git ghp_xxxxxxxxxxxx"
    exit 1
fi

# Extract repo URL and add token
# Format: https://TOKEN@github.com/username/repo.git
PUSH_URL=$(echo "$REPO_URL" | sed "s|https://|https://${TOKEN}@|")

echo "🔗 Setting up remote..."
git remote remove origin 2>/dev/null || true
git remote add origin "$REPO_URL"

echo "🚀 Pushing to GitHub..."
git push -u origin main

echo "✅ Successfully pushed to GitHub!"
echo "🔗 Repository: $REPO_URL"

