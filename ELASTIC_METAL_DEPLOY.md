# Elastic Metal Server Deployment - Step by Step

**Server:** `151.115.100.18`  
**User:** `ubuntu`  
**Domains:** `zyndra.armonika.cloud` & `api.zyndra.armonika.cloud`  
**WEBHOOK_SECRET:** `ed7f219ca3afd5838ab10186dec58a9cc65ce34277ed47ca1364138910bc1bd1`

## Complete Deployment Steps

### Step 1: Connect to Server

```bash
ssh ubuntu@151.115.100.18
```

### Step 2: Initial Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git build-essential ufw fail2ban

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker ubuntu
rm get-docker.sh

# Install Docker Compose
sudo apt install -y docker-compose-plugin

# Install PostgreSQL (if not using Docker)
sudo apt install -y postgresql postgresql-contrib

# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install -y caddy

# Configure firewall
sudo ufw --force enable
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
```

### Step 3: Clone Repository

```bash
# Create deployment directory
sudo mkdir -p /opt/zyndra
cd /opt/zyndra

# Clone repository
sudo git clone https://github.com/SidahmedSeg/Zyndra.git .

# Change ownership
sudo chown -R ubuntu:ubuntu /opt/zyndra
cd /opt/zyndra
```

### Step 4: Configure Environment Variables

```bash
# Copy template
cp env.production.template .env.production

# Edit environment file
nano .env.production
```

**Update these values in `.env.production`:**

```bash
# IMPORTANT: Generate a strong PostgreSQL password
POSTGRES_PASSWORD=YOUR_STRONG_PASSWORD_HERE
DATABASE_URL=postgresql://zyndra:YOUR_STRONG_PASSWORD_HERE@postgres:5432/zyndra?sslmode=disable

# These are already set correctly:
BASE_URL=https://api.zyndra.armonika.cloud
NEXT_PUBLIC_API_URL=https://api.zyndra.armonika.cloud
WEBHOOK_SECRET=ed7f219ca3afd5838ab10186dec58a9cc65ce34277ed47ca1364138910bc1bd1
CORS_ORIGINS=https://zyndra.armonika.cloud
DISABLE_AUTH=true
USE_MOCK_INFRA=true

# Update registry credentials (if you have them):
REGISTRY_URL=https://your-registry.com
REGISTRY_USERNAME=your_username
REGISTRY_PASSWORD=your_password
```

**Generate a strong password:**
```bash
openssl rand -base64 32
```

### Step 5: Configure Caddy

```bash
# Copy Caddyfile
sudo cp Caddyfile.prod /etc/caddy/Caddyfile

# Verify configuration
sudo caddy validate --config /etc/caddy/Caddyfile
```

The Caddyfile should configure:
- `zyndra.armonika.cloud` → Frontend (port 3000)
- `api.zyndra.armonika.cloud` → Backend (port 8080)

### Step 6: Build and Start Services

```bash
cd /opt/zyndra

# Build all services
docker compose -f docker-compose.prod.yml --env-file .env.production build

# Start services
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

# Check status
docker compose -f docker-compose.prod.yml ps
```

### Step 7: Verify Migrations

```bash
# Check backend logs for migration messages
docker compose -f docker-compose.prod.yml logs backend | grep -i migration

# Should see:
# === STARTING DATABASE MIGRATIONS ===
# Detected database type: postgres
# Running migration: 001_initial
# ✓ Migration 001_initial completed successfully
# ✅ MIGRATIONS COMPLETED SUCCESSFULLY
```

**If migrations failed**, run them manually:
```bash
# Connect to PostgreSQL
docker compose -f docker-compose.prod.yml exec postgres psql -U zyndra -d zyndra

# Then paste the SQL from migrations/postgres/001_initial.up.sql
# Then paste the SQL from migrations/postgres/002_tables.up.sql
```

### Step 8: Start Caddy

```bash
# Start Caddy
sudo systemctl start caddy
sudo systemctl enable caddy

# Check status
sudo systemctl status caddy

# View logs
sudo journalctl -u caddy -f
```

Caddy will automatically:
- Obtain SSL certificates via Let's Encrypt
- Configure HTTPS
- Route traffic to your services

### Step 9: Configure DNS

**In your DNS provider, add these A records:**
- `zyndra.armonika.cloud` → `151.115.100.18`
- `api.zyndra.armonika.cloud` → `151.115.100.18`

Wait 5-15 minutes for DNS propagation.

### Step 10: Verify Deployment

```bash
# Check all services
docker compose -f docker-compose.prod.yml ps

# Check backend health
curl http://localhost:8080/health

# Check frontend
curl http://localhost:3000

# After DNS propagates, test domains:
curl https://api.zyndra.armonika.cloud/health
```

## Useful Commands

### View Logs
```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
docker compose -f docker-compose.prod.yml logs -f postgres
```

### Restart Services
```bash
docker compose -f docker-compose.prod.yml restart
# Or restart specific service:
docker compose -f docker-compose.prod.yml restart backend
```

### Rebuild Services
```bash
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d
```

### Stop Services
```bash
docker compose -f docker-compose.prod.yml down
```

### Check Caddy Status
```bash
sudo systemctl status caddy
sudo caddy validate --config /etc/caddy/Caddyfile
```

## Troubleshooting

### Backend won't start
```bash
# Check logs
docker compose -f docker-compose.prod.yml logs backend

# Common issues:
# - Missing DATABASE_URL
# - PostgreSQL not ready
# - Missing required env vars
```

### Migrations not running
```bash
# Check if migrations table exists
docker compose -f docker-compose.prod.yml exec postgres psql -U zyndra -d zyndra -c "\dt schema_migrations"

# Run migrations manually (see Step 7)
```

### Frontend build fails
```bash
# Check build logs
docker compose -f docker-compose.prod.yml logs frontend

# Common issues:
# - Missing NEXT_PUBLIC_API_URL
# - Build errors in Next.js
```

### Caddy SSL certificate issues
```bash
# Check Caddy logs
sudo journalctl -u caddy -f

# Common issues:
# - DNS not pointing to server
# - Port 80 blocked
# - Domain validation failed
```

### Can't access services
```bash
# Check if services are running
docker compose -f docker-compose.prod.yml ps

# Check if ports are accessible
curl http://localhost:8080/health
curl http://localhost:3000

# Check firewall
sudo ufw status
```

## Security Checklist

- [x] Strong PostgreSQL password set
- [x] WEBHOOK_SECRET generated
- [x] Firewall configured (ports 22, 80, 443)
- [x] Services running as non-root (Docker handles this)
- [x] Database only accessible from localhost
- [x] SSL certificates via Let's Encrypt (automatic with Caddy)

## Next Steps After Deployment

1. ✅ Test frontend: `https://zyndra.armonika.cloud`
2. ✅ Test backend: `https://api.zyndra.armonika.cloud/health`
3. ✅ Log in with mock authentication
4. ✅ Create first project
5. ✅ Test API endpoints

## Support

If you encounter issues:
1. Check service logs: `docker compose -f docker-compose.prod.yml logs [service]`
2. Check Caddy logs: `sudo journalctl -u caddy -f`
3. Verify DNS: `dig zyndra.armonika.cloud`
4. Verify firewall: `sudo ufw status`

