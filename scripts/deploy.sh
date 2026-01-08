#!/bin/bash
# Elastic Metal Server Deployment Script
# Usage: ./scripts/deploy.sh

set -e

echo "=========================================="
echo "Zyndra Deployment Script"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root or with sudo${NC}"
    exit 1
fi

echo -e "${GREEN}Step 1: Updating system...${NC}"
apt update && apt upgrade -y

echo -e "${GREEN}Step 2: Installing essential packages...${NC}"
apt install -y curl wget git build-essential ufw fail2ban

echo -e "${GREEN}Step 3: Installing Docker...${NC}"
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    systemctl enable docker
    systemctl start docker
    echo -e "${GREEN}Docker installed${NC}"
else
    echo -e "${YELLOW}Docker already installed${NC}"
fi

echo -e "${GREEN}Step 4: Installing Docker Compose...${NC}"
if ! command -v docker compose &> /dev/null; then
    apt install -y docker-compose-plugin
    echo -e "${GREEN}Docker Compose installed${NC}"
else
    echo -e "${YELLOW}Docker Compose already installed${NC}"
fi

echo -e "${GREEN}Step 5: Installing PostgreSQL...${NC}"
if ! command -v psql &> /dev/null; then
    apt install -y postgresql postgresql-contrib
    systemctl enable postgresql
    systemctl start postgresql
    echo -e "${GREEN}PostgreSQL installed${NC}"
else
    echo -e "${YELLOW}PostgreSQL already installed${NC}"
fi

echo -e "${GREEN}Step 6: Installing Caddy...${NC}"
if ! command -v caddy &> /dev/null; then
    apt install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main' > /etc/apt/sources.list.d/caddy-stable.list
    apt update
    apt install -y caddy
    systemctl enable caddy
    echo -e "${GREEN}Caddy installed${NC}"
else
    echo -e "${YELLOW}Caddy already installed${NC}"
fi

echo -e "${GREEN}Step 7: Configuring firewall...${NC}"
ufw --force enable
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw allow 8080/tcp  # Backend API (internal)
ufw allow 3000/tcp  # Frontend (internal)
echo -e "${GREEN}Firewall configured${NC}"

echo -e "${GREEN}Step 8: Creating deployment directory...${NC}"
mkdir -p /opt/zyndra
cd /opt/zyndra

echo ""
echo -e "${GREEN}=========================================="
echo "Installation Complete!"
echo "==========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Clone your repository: git clone <your-repo> /opt/zyndra"
echo "2. Configure environment variables"
echo "3. Run database migrations"
echo "4. Start services with docker-compose"
echo ""

