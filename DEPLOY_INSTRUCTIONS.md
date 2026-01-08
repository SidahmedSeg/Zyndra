# Deployment Instructions for Elastic Metal Server

**Server:** `151.115.100.18`  
**User:** `ubuntu`  
**Frontend:** `zyndra.armonika.cloud`  
**Backend:** `api.zyndra.armonika.cloud`  
**WEBHOOK_SECRET:** `ed7f219ca3afd5838ab10186dec58a9cc65ce34277ed47ca1364138910bc1bd1`

## Step 1: Connect to Server

```bash
ssh ubuntu@151.115.100.18
```

## Step 2: Check OS and Update

```bash
cat /etc/os-release
sudo apt update && sudo apt upgrade -y
```

## Step 3: Run Automated Setup

```bash
# Clone the repository
cd /opt
sudo git clone https://github.com/SidahmedSeg/Zyndra.git zyndra
cd zyndra

# Make deployment script executable
sudo chmod +x scripts/deploy.sh

# Run automated setup (installs Docker, PostgreSQL, Caddy, etc.)
sudo ./scripts/deploy.sh
```

## Step 4: Configure Environment Variables

```bash
cd /opt/zyndra

# Copy example env file
sudo cp env.production.example .env.production

# Edit with your values
sudo nano .env.production
```

**Use these values:**
```bash
DATABASE_URL=postgresql://zyndra:CHANGE_THIS_PASSWORD@postgres:5432/zyndra?sslmode=disable
POSTGRES_DB=zyndra
POSTGRES_USER=zyndra
POSTGRES_PASSWORD=YOUR_STRONG_DB_PASSWORD_HERE
BASE_URL=https://api.zyndra.armonika.cloud
NEXT_PUBLIC_API_URL=https://api.zyndra.armonika.cloud
DISABLE_AUTH=true
REGISTRY_URL=https://registry.example.com
REGISTRY_USERNAME=your_registry_username
REGISTRY_PASSWORD=your_registry_password
WEBHOOK_SECRET=ed7f219ca3afd5838ab10186dec58a9cc65ce34277ed47ca1364138910bc1bd1
CORS_ORIGINS=https://zyndra.armonika.cloud
USE_MOCK_INFRA=true
```

**IMPORTANT:** Replace `CHANGE_THIS_PASSWORD` and `YOUR_STRONG_DB_PASSWORD_HERE` with a strong password!

## Step 5: Configure Caddy

```bash
cd /opt/zyndra

# Copy Caddyfile to Caddy config directory
sudo cp Caddyfile.prod /etc/caddy/Caddyfile

# Edit Caddyfile to match your domains (if needed)
sudo nano /etc/caddy/Caddyfile
```

The Caddyfile should have:
- `zyndra.armonika.cloud` → frontend (port 3000)
- `api.zyndra.armonika.cloud` → backend (port 8080)

## Step 6: Start Services with Docker Compose

```bash
cd /opt/zyndra

# Build and start all services
sudo docker compose -f docker-compose.prod.yml --env-file .env.production up -d

# Check logs
sudo docker compose -f docker-compose.prod.yml logs -f
```

## Step 7: Run Database Migrations

The backend should automatically run migrations on startup. To verify:

```bash
# Check backend logs for migration messages
sudo docker compose -f docker-compose.prod.yml logs backend | grep -i migration

# Or run migrations manually if needed
sudo docker compose -f docker-compose.prod.yml exec backend /app/click-deploy --migrate
```

If migrations don't run automatically, connect to PostgreSQL and run them manually:

```bash
# Connect to PostgreSQL container
sudo docker compose -f docker-compose.prod.yml exec postgres psql -U zyndra -d zyndra

# Then run the SQL from migrations/postgres/001_initial.up.sql and 002_tables.up.sql
```

## Step 8: Configure DNS

Point your domains to the server IP:

**DNS Records (A records):**
- `zyndra.armonika.cloud` → `151.115.100.18`
- `api.zyndra.armonika.cloud` → `151.115.100.18`

Wait for DNS propagation (usually 5-15 minutes).

## Step 9: Start Caddy

```bash
# Start and enable Caddy
sudo systemctl start caddy
sudo systemctl enable caddy

# Check Caddy status
sudo systemctl status caddy

# View Caddy logs
sudo journalctl -u caddy -f
```

Caddy will automatically:
- Obtain SSL certificates from Let's Encrypt
- Set up HTTPS for both domains
- Route traffic to your services

## Step 10: Verify Deployment

1. **Check all services are running:**
   ```bash
   sudo docker compose -f docker-compose.prod.yml ps
   ```

2. **Check backend health:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Check frontend:**
   ```bash
   curl http://localhost:3000
   ```

4. **After DNS propagates, test domains:**
   - Frontend: `https://zyndra.armonika.cloud`
   - Backend: `https://api.zyndra.armonika.cloud/health`

## Troubleshooting

### View logs:
```bash
# All services
sudo docker compose -f docker-compose.prod.yml logs -f

# Specific service
sudo docker compose -f docker-compose.prod.yml logs -f backend
sudo docker compose -f docker-compose.prod.yml logs -f frontend
sudo docker compose -f docker-compose.prod.yml logs -f postgres
```

### Restart services:
```bash
sudo docker compose -f docker-compose.prod.yml restart
```

### Rebuild services:
```bash
sudo docker compose -f docker-compose.prod.yml build --no-cache
sudo docker compose -f docker-compose.prod.yml up -d
```

### Check Caddy SSL status:
```bash
sudo caddy validate --config /etc/caddy/Caddyfile
```

## Next Steps After Deployment

1. Test the frontend at `https://zyndra.armonika.cloud`
2. Test the backend API at `https://api.zyndra.armonika.cloud/health`
3. Log in with mock authentication
4. Create your first project

## Security Checklist

- [ ] Changed default PostgreSQL password
- [ ] Set strong WEBHOOK_SECRET (done: using generated secret)
- [ ] Caddy SSL certificates obtained (automatic)
- [ ] Firewall configured (ports 22, 80, 443 open)
- [ ] Services running as non-root users (done in Docker)
- [ ] Database only accessible from localhost (done in docker-compose)

