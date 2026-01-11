#!/bin/bash
# Zyndra k3s Cluster Verification Script

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check() {
    if eval "$2" &>/dev/null; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1"
        return 1
    fi
}

warn_check() {
    if eval "$2" &>/dev/null; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${YELLOW}⚠${NC} $1 (optional)"
    fi
}

echo "Zyndra k3s Cluster Health Check"
echo "================================"
echo ""

# Core components
echo "Core Components:"
check "k3s installed" "command -v k3s"
check "kubectl available" "kubectl version --client"
check "Helm installed" "command -v helm"
echo ""

# Kubernetes nodes
echo "Kubernetes Cluster:"
check "Node ready" "kubectl get nodes | grep -q Ready"
check "API server responding" "kubectl cluster-info"
echo ""

# Namespaces
echo "Required Namespaces:"
check "traefik-system" "kubectl get namespace traefik-system"
check "cert-manager" "kubectl get namespace cert-manager"
check "longhorn-system" "kubectl get namespace longhorn-system"
check "zyndra-system" "kubectl get namespace zyndra-system"
echo ""

# Deployments
echo "Core Deployments:"
check "Traefik" "kubectl get deployment traefik -n traefik-system"
check "cert-manager" "kubectl get deployment cert-manager -n cert-manager"
check "cert-manager-webhook" "kubectl get deployment cert-manager-webhook -n cert-manager"
check "Longhorn manager" "kubectl get deployment longhorn-manager -n longhorn-system"
warn_check "Metrics server" "kubectl get deployment metrics-server -n kube-system"
echo ""

# ClusterIssuers
echo "SSL Certificate Issuers:"
check "letsencrypt-prod" "kubectl get clusterissuer letsencrypt-prod"
warn_check "letsencrypt-staging" "kubectl get clusterissuer letsencrypt-staging"
echo ""

# Storage
echo "Storage:"
check "Longhorn StorageClass" "kubectl get storageclass longhorn"
echo ""

# Metrics
echo "Metrics:"
if kubectl top nodes &>/dev/null; then
    echo -e "${GREEN}✓${NC} Metrics API available"
    echo ""
    echo "Node Resources:"
    kubectl top nodes
else
    echo -e "${YELLOW}⚠${NC} Metrics not available yet (may take a few minutes)"
fi
echo ""

echo "================================"
echo "Verification complete!"

