#!/bin/bash
# Complete automated deployment script for Elastic Metal Server
# Run this script on the server: bash <(curl -s https://raw.githubusercontent.com/SidahmedSeg/Zyndra/main/scripts/full-deploy.sh)
# Or: wget https://raw.githubusercontent.com/SidahmedSeg/Zyndra/main/scripts/full-deploy.sh && bash full-deploy.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================="
echo "Zyndra Complete Deployment Script"
echo "==========================================${NC}"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then 
    echo -e "${YELLOW}Note: Some commands require sudo. You may be prompted for password.${NC}"
    SUDO="sudo"
else
    SUDO=""
fi

# Configuration
DEPLOY_DIR="/opt/zyndra"
REPO_URL="https://github.com/SidahmedSeg/Zyndra.git"
FRONTEND_DOMAIN="zyndra.armonika.cloud"
BACKEND_DOMAIN="api.zyndra.armonika.cloud"
WEBHOOK_SECRET="ed7f219ca3afd5838ab10186dec58a9cc65ce34277ed47ca1364138910bc1bd1"

echo -e "${GREEN}Step 1: Updating system packages...${NC}"
$SUDO apt update && $SUDO apt upgrade -y

echo -e "${GREEN}Step 2: Installing essential packages...${NC}"
$SUDO apt install -y curl wget git build-essential ufw fail2ban openssl

echo -e "${GREEN}Step 3: Installing Docker...${NC}"
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    $SUDO sh /tmp/get-docker.sh
    rm /tmp/get-docker.sh
    $SUDO systemctl enable docker
    $SUDO systemctl start docker
    # Add current user to docker group
    if [ -n "$SUDO_USER" ]; then
        $SUDO usermod -aG docker $SUDO_USER
    else
        $SUDO usermod -aG docker $USER
    fi
    echo -e "${GREEN}Docker installed${NC}"
else
    echo -e "${YELLOW}Docker already installed${NC}"
fi

echo -e "${GREEN}Step 4: Installing Docker Compose...${NC}"
if ! command -v docker compose &> /dev/null; then
    $SUDO apt install -y docker-compose-plugin
    echo -e "${GREEN}Docker Compose installed${NC}"
else
    echo -e "${YELLOW}Docker Compose already installed${NC}"
fi

echo -e "${GREEN}Step 5: Installing Caddy...${NC}"
if ! command -v caddy &> /dev/null; then
    $SUDO apt install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | $SUDO gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main' | $SUDO tee /etc/apt/sources.list.d/caddy-stable.list
    $SUDO apt update
    $SUDO apt install -y caddy
    $SUDO systemctl enable caddy
    echo -e "${GREEN}Caddy installed${NC}"
else
    echo -e "${YELLOW}Caddy already installed${NC}"
fi

echo -e "${GREEN}Step 6: Configuring firewall...${NC}"
$SUDO ufw --force enable
$SUDO ufw allow 22/tcp    # SSH
$SUDO ufw allow 80/tcp    # HTTP
$SUDO ufw allow 443/tcp   # HTTPS
echo -e "${GREEN}Firewall configured${NC}"

echo -e "${GREEN}Step 7: Creating deployment directory...${NC}"
$SUDO mkdir -p $DEPLOY_DIR
cd $DEPLOY_DIR

echo -e "${GREEN}Step 8: Cloning repository...${NC}"
if [ -d ".git" ]; then
    echo -e "${YELLOW}Repository already exists, pulling latest...${NC}"
    $SUDO git pull
else
    $SUDO git clone $REPO_URL .
fi

# Change ownership
if [ -n "$SUDO_USER" ]; then
    $SUDO chown -R $SUDO_USER:$SUDO_USER $DEPLOY_DIR
    cd $DEPLOY_DIR
else
    $SUDO chown -R $USER:$USER $DEPLOY_DIR
    cd $DEPLOY_DIR
fi

echo -e "${GREEN}Step 9: Creating environment file...${NC}"
if [ ! -f .env.production ]; then
    cp env.production.template .env.production
    
    # Generate PostgreSQL password
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    sed -i "s/REPLACE_WITH_STRONG_PASSWORD/$DB_PASSWORD/g" .env.production
    
    # Set domains
    sed -i "s|BASE_URL=.*|BASE_URL=https://$BACKEND_DOMAIN|g" .env.production
    sed -i "s|NEXT_PUBLIC_API_URL=.*|NEXT_PUBLIC_API_URL=https://$BACKEND_DOMAIN|g" .env.production
    sed -i "s|WEBHOOK_SECRET=.*|WEBHOOK_SECRET=$WEBHOOK_SECRET|g" .env.production
    sed -i "s|CORS_ORIGINS=.*|CORS_ORIGINS=https://$FRONTEND_DOMAIN|g" .env.production
    
    echo -e "${GREEN}Environment file created with generated password${NC}"
    echo -e "${YELLOW}PostgreSQL password: $DB_PASSWORD${NC}"
    echo -e "${YELLOW}Save this password!${NC}"
else
    echo -e "${YELLOW}Environment file already exists, skipping...${NC}"
fi

echo -e "${GREEN}Step 10: Configuring Caddy...${NC}"
$SUDO cp Caddyfile.prod /etc/caddy/Caddyfile
$SUDO caddy validate --config /etc/caddy/Caddyfile || echo -e "${YELLOW}Caddy validation warning (may be OK)${NC}"

echo -e "${GREEN}Step 11: Building Docker images...${NC}"
docker compose -f docker-compose.prod.yml --env-file .env.production build

echo -e "${GREEN}Step 12: Starting services...${NC}"
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

echo -e "${GREEN}Step 13: Waiting for services to start...${NC}"
sleep 15

echo -e "${GREEN}Step 14: Starting Caddy...${NC}"
$SUDO systemctl start caddy
$SUDO systemctl enable caddy

echo -e "${GREEN}Step 15: Checking service status...${NC}"
docker compose -f docker-compose.prod.yml ps

echo ""
echo -e "${BLUE}=========================================="
echo "Deployment Complete!"
echo "==========================================${NC}"
echo ""
echo -e "${GREEN}Services Status:${NC}"
docker compose -f docker-compose.prod.yml ps
echo ""
echo -e "${GREEN}Backend Migration Status:${NC}"
docker compose -f docker-compose.prod.yml logs backend | grep -i migration | tail -10 || echo "Check logs manually"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Configure DNS:"
echo "   - $FRONTEND_DOMAIN → 151.115.100.18"
echo "   - $BACKEND_DOMAIN → 151.115.100.18"
echo ""
echo "2. Wait 5-15 minutes for DNS propagation"
echo ""
echo "3. Verify deployment:"
echo "   curl http://localhost:8080/health"
echo "   curl http://localhost:3000"
echo ""
echo "4. After DNS propagates:"
echo "   https://$FRONTEND_DOMAIN"
echo "   https://$BACKEND_DOMAIN/health"
echo ""
echo -e "${GREEN}View logs:${NC}"
echo "   docker compose -f docker-compose.prod.yml logs -f"
echo ""

