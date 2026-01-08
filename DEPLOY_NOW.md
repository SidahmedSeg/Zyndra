# Quick Deployment - Copy & Paste This!

Since I can't SSH directly (need private key), here's what to do:

## Step 1: SSH to Your Server

```bash
ssh ubuntu@151.115.100.18
```

## Step 2: Run This One Command

Copy and paste this entire command:

```bash
bash <(curl -s https://raw.githubusercontent.com/SidahmedSeg/Zyndra/main/scripts/full-deploy.sh)
```

That's it! The script will:
- ✅ Install all dependencies (Docker, Caddy, etc.)
- ✅ Clone your repository
- ✅ Configure everything
- ✅ Start all services
- ✅ Set up SSL automatically

## Alternative: If you prefer step-by-step

```bash
# 1. Update system
sudo apt update && sudo apt upgrade -y

# 2. Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh && sudo usermod -aG docker ubuntu && rm get-docker.sh
newgrp docker

# 3. Install Docker Compose
sudo apt install -y docker-compose-plugin

# 4. Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/caddy-stable-archive-keyring.gpg] https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main" | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy

# 5. Configure firewall
sudo ufw --force enable && sudo ufw allow 22/tcp && sudo ufw allow 80/tcp && sudo ufw allow 443/tcp

# 6. Clone repo
sudo mkdir -p /opt/zyndra && cd /opt/zyndra
sudo git clone https://github.com/SidahmedSeg/Zyndra.git .
sudo chown -R ubuntu:ubuntu /opt/zyndra && cd /opt/zyndra

# 7. Setup environment
cp env.production.template .env.production
DB_PASS=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
sed -i "s/REPLACE_WITH_STRONG_PASSWORD/$DB_PASS/g" .env.production
echo "PostgreSQL Password: $DB_PASS"

# 8. Configure Caddy
sudo cp Caddyfile.prod /etc/caddy/Caddyfile

# 9. Build and start
docker compose -f docker-compose.prod.yml --env-file .env.production build
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

# 10. Start Caddy
sudo systemctl start caddy && sudo systemctl enable caddy

# 11. Check status
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs backend | grep -i migration | tail -10
```

## After Deployment

1. **Configure DNS:**
   - `zyndra.armonika.cloud` → `151.115.100.18`
   - `api.zyndra.armonika.cloud` → `151.115.100.18`

2. **Wait 5-15 minutes** for DNS propagation

3. **Test:**
   - Frontend: `https://zyndra.armonika.cloud`
   - Backend: `https://api.zyndra.armonika.cloud/health`

## View Logs

```bash
cd /opt/zyndra
docker compose -f docker-compose.prod.yml logs -f
```

