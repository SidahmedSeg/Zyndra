# Elastic Metal Server Deployment Guide

## Prerequisites Information Needed

Before starting deployment, provide:

1. **Server Access:**
   - Server IP address
   - SSH username (usually `root` or `ubuntu`)
   - SSH key or password access method

2. **Domain Configuration:**
   - Domain for frontend (e.g., `zyndra.armonika.cloud`)
   - Domain for backend API (e.g., `api.zyndra.armonika.cloud`)
   - DNS access (to point domains to server IP)

3. **Database:**
   - Will you use PostgreSQL on the same server or external?
   - If external, provide connection string
   - If same server, we'll install PostgreSQL

4. **Environment Variables:**
   - `DATABASE_URL` (PostgreSQL connection string)
   - `BASE_URL` (backend URL, e.g., `https://api.zyndra.armonika.cloud`)
   - `REGISTRY_URL`, `REGISTRY_USERNAME`, `REGISTRY_PASSWORD`
   - `WEBHOOK_SECRET` (generate a random secret)
   - `DISABLE_AUTH=true` (for now, using mock auth)

## Quick Start Checklist

### What I Need From You:
- [ ] Server IP address: `_____________`
- [ ] SSH username: `_____________`
- [ ] SSH key path or password: `_____________`
- [ ] Frontend domain: `_____________`
- [ ] Backend API domain: `_____________`
- [ ] Database: Same server or external?
- [ ] OS version: `_____________` (run `cat /etc/os-release`)

### What We'll Set Up:
- [ ] Server initialization (update, security)
- [ ] Install Docker & Docker Compose
- [ ] Install PostgreSQL (if on same server)
- [ ] Install Caddy (reverse proxy & SSL)
- [ ] Install Prometheus (metrics)
- [ ] Configure firewall
- [ ] Deploy backend service
- [ ] Deploy frontend service
- [ ] Configure DNS
- [ ] Set up SSL certificates
- [ ] Configure systemd services
- [ ] Set up monitoring

## Deployment Architecture

```
Internet
    ↓
[Caddy Reverse Proxy] (Port 80/443)
    ├──→ Backend API (Port 8080)
    │     ├── Go Application
    │     └── PostgreSQL
    └──→ Frontend (Port 3000)
          └── Next.js Application
```

## Next Steps

Once you provide the information above, I'll:
1. Create deployment scripts
2. Generate configuration files
3. Provide step-by-step commands to run
4. Set up everything automatically

