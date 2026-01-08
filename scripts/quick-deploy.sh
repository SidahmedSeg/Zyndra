#!/bin/bash
# Quick deployment script for elastic metal server
# Run this after initial setup

set -e

echo "=========================================="
echo "Zyndra Quick Deployment"
echo "=========================================="
echo ""

cd /opt/zyndra || exit 1

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo "ERROR: .env.production not found!"
    echo "Please create it from env.production.template first:"
    echo "  cp env.production.template .env.production"
    echo "  nano .env.production  # Edit passwords"
    exit 1
fi

# Generate PostgreSQL password if not set
if ! grep -q "POSTGRES_PASSWORD=.*[A-Za-z0-9]" .env.production || grep -q "REPLACE_WITH_STRONG_PASSWORD" .env.production; then
    echo "Generating PostgreSQL password..."
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    sed -i "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=$DB_PASSWORD/" .env.production
    sed -i "s|postgresql://zyndra:.*@postgres|postgresql://zyndra:$DB_PASSWORD@postgres|" .env.production
    echo "Generated password saved to .env.production"
fi

echo "Building and starting services..."
docker compose -f docker-compose.prod.yml --env-file .env.production build

echo "Starting services..."
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

echo ""
echo "Waiting for services to start..."
sleep 10

echo ""
echo "Checking service status..."
docker compose -f docker-compose.prod.yml ps

echo ""
echo "Checking backend logs for migrations..."
sleep 5
docker compose -f docker-compose.prod.yml logs backend | grep -i migration | tail -20

echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Configure DNS:"
echo "   - zyndra.armonika.cloud → 151.115.100.18"
echo "   - api.zyndra.armonika.cloud → 151.115.100.18"
echo ""
echo "2. Configure Caddy:"
echo "   sudo cp Caddyfile.prod /etc/caddy/Caddyfile"
echo "   sudo systemctl restart caddy"
echo ""
echo "3. Check logs:"
echo "   docker compose -f docker-compose.prod.yml logs -f"
echo ""
echo "4. Verify services:"
echo "   curl http://localhost:8080/health"
echo "   curl http://localhost:3000"
echo ""

