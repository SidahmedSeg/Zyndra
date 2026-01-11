#!/bin/bash
# Zyndra k3s Cluster Installation Script
# This script sets up a complete k3s cluster with all required components

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[ZYNDRA]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Configuration
DOMAIN=${DOMAIN:-"up.zyndra.app"}
EMAIL=${EMAIL:-"admin@zyndra.app"}
REGISTRY_URL=${REGISTRY_URL:-"registry.zyndra.app"}

log "Starting Zyndra k3s Cluster Installation"
log "Domain: $DOMAIN"
log "Email: $EMAIL"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root (sudo)"
fi

# 1. Install k3s
log "Step 1: Installing k3s..."
if command -v k3s &> /dev/null; then
    warn "k3s is already installed, skipping..."
else
    curl -sfL https://get.k3s.io | sh -s - \
        --disable traefik \
        --write-kubeconfig-mode 644
    
    # Wait for k3s to be ready
    log "Waiting for k3s to be ready..."
    sleep 10
    kubectl wait --for=condition=Ready nodes --all --timeout=120s
fi

# 2. Install Helm
log "Step 2: Installing Helm..."
if command -v helm &> /dev/null; then
    warn "Helm is already installed, skipping..."
else
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# 3. Add Helm repositories
log "Step 3: Adding Helm repositories..."
helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
helm repo add longhorn https://charts.longhorn.io 2>/dev/null || true
helm repo add traefik https://traefik.github.io/charts 2>/dev/null || true
helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
helm repo update

# 4. Install Traefik (as Ingress Controller)
log "Step 4: Installing Traefik Ingress Controller..."
kubectl create namespace traefik-system 2>/dev/null || true
helm upgrade --install traefik traefik/traefik \
    --namespace traefik-system \
    --set ingressClass.enabled=true \
    --set ingressClass.isDefaultClass=true \
    --set ports.web.redirectTo.port=websecure \
    --set ports.websecure.tls.enabled=true \
    --wait

# 5. Install cert-manager
log "Step 5: Installing cert-manager..."
kubectl create namespace cert-manager 2>/dev/null || true
helm upgrade --install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --set installCRDs=true \
    --wait

# Wait for cert-manager webhook
log "Waiting for cert-manager webhook..."
kubectl wait --for=condition=Available deployment/cert-manager-webhook \
    -n cert-manager --timeout=120s

# 6. Create ClusterIssuer for Let's Encrypt
log "Step 6: Creating Let's Encrypt ClusterIssuer..."
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ${EMAIL}
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: ${EMAIL}
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: traefik
EOF

# 7. Install Longhorn (Distributed Storage)
log "Step 7: Installing Longhorn..."
kubectl create namespace longhorn-system 2>/dev/null || true

# Install open-iscsi if not present (required for Longhorn)
if ! command -v iscsiadm &> /dev/null; then
    log "Installing open-iscsi (required for Longhorn)..."
    apt-get update && apt-get install -y open-iscsi
    systemctl enable iscsid
    systemctl start iscsid
fi

helm upgrade --install longhorn longhorn/longhorn \
    --namespace longhorn-system \
    --set defaultSettings.defaultDataPath="/var/lib/longhorn" \
    --set persistence.defaultClassReplicaCount=1 \
    --wait

# 8. Install Metrics Server
log "Step 8: Installing Metrics Server..."
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# For single-node clusters, we need to disable TLS verification
kubectl patch deployment metrics-server -n kube-system \
    --type='json' \
    -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]' 2>/dev/null || true

# 9. Create Zyndra namespace
log "Step 9: Creating Zyndra system namespace..."
kubectl create namespace zyndra-system 2>/dev/null || true

# 10. Create registry secret (if credentials provided)
if [ -n "$REGISTRY_USERNAME" ] && [ -n "$REGISTRY_PASSWORD" ]; then
    log "Step 10: Creating registry secret..."
    kubectl create secret docker-registry registry-creds \
        --docker-server=$REGISTRY_URL \
        --docker-username=$REGISTRY_USERNAME \
        --docker-password=$REGISTRY_PASSWORD \
        --namespace zyndra-system 2>/dev/null || true
else
    warn "Skipping registry secret (REGISTRY_USERNAME and REGISTRY_PASSWORD not set)"
fi

# Summary
log "============================================"
log "Zyndra k3s Cluster Installation Complete!"
log "============================================"
log ""
log "Installed Components:"
log "  - k3s (Kubernetes)"
log "  - Traefik (Ingress Controller)"
log "  - cert-manager (SSL Certificates)"
log "  - Longhorn (Distributed Storage)"
log "  - Metrics Server (CPU/Memory monitoring)"
log ""
log "ClusterIssuers:"
log "  - letsencrypt-prod (production certificates)"
log "  - letsencrypt-staging (testing certificates)"
log ""
log "Kubeconfig: /etc/rancher/k3s/k3s.yaml"
log ""
log "Next Steps:"
log "  1. Point *.${DOMAIN} to this server's IP"
log "  2. Set environment variables in Zyndra backend:"
log "     USE_K8S=true"
log "     K8S_BASE_DOMAIN=${DOMAIN}"
log "     K8S_INGRESS_CLASS=traefik"
log "     K8S_CERT_ISSUER=letsencrypt-prod"
log ""
log "To verify installation:"
log "  kubectl get pods -A"
log "  kubectl get clusterissuers"
log "============================================"

