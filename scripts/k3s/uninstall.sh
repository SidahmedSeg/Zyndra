#!/bin/bash
# Zyndra k3s Cluster Uninstall Script

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[ZYNDRA]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo)"
    exit 1
fi

log "Uninstalling Zyndra k3s Cluster..."

# Confirm
read -p "This will remove k3s and all data. Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

# Uninstall Helm releases
log "Removing Helm releases..."
helm uninstall longhorn -n longhorn-system 2>/dev/null || true
helm uninstall cert-manager -n cert-manager 2>/dev/null || true
helm uninstall traefik -n traefik-system 2>/dev/null || true

# Remove namespaces
log "Removing namespaces..."
kubectl delete namespace longhorn-system --timeout=60s 2>/dev/null || true
kubectl delete namespace cert-manager --timeout=60s 2>/dev/null || true
kubectl delete namespace traefik-system --timeout=60s 2>/dev/null || true
kubectl delete namespace zyndra-system --timeout=60s 2>/dev/null || true

# Uninstall k3s
log "Uninstalling k3s..."
/usr/local/bin/k3s-uninstall.sh 2>/dev/null || true

# Clean up data
log "Cleaning up data..."
rm -rf /var/lib/longhorn 2>/dev/null || true

log "Uninstallation complete!"

