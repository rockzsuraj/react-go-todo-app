# Production Deployment Guide

## Prerequisites
1. **Kubernetes Cluster** (EKS/GKE/AKS)
2. **GitHub Container Registry Access**
3. **Domain Name** for ingress

## Step 1: Setup GitHub Secrets
```bash
# In your GitHub repository settings, add these secrets:
KUBE_CONFIG          # Base64 encoded kubeconfig
API_URL              # https://api.yourdomain.com
SUPABASE_URL         # Your Supabase URL
SUPABASE_ANON_KEY    # Your Supabase anon key
```

## Step 2: Push Images to Registry
```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and push images
make push-images API_URL=https://api.yourdomain.com SUPABASE_URL=your-url SUPABASE_ANON_KEY=your-key
```

## Step 3: Deploy to Production Cluster
```bash
# Connect to your production cluster
kubectl config use-context your-production-cluster

# Deploy application
make deploy

# Check status
make status
```

## Step 4: Setup Ingress & SSL
Update `k8s/ingress.yaml` with your domain:
```yaml
spec:
  tls:
  - hosts:
    - yourdomain.com
    secretName: todo-tls
  rules:
  - host: yourdomain.com
```

## Step 5: Monitor Deployment
```bash
# Watch rollout
kubectl rollout status deployment/todo-frontend -n todo-app
kubectl rollout status deployment/todo-backend -n todo-app

# Check logs
kubectl logs -f deployment/todo-frontend -n todo-app
kubectl logs -f deployment/todo-backend -n todo-app
```

## CI/CD Pipeline
Push to `main` branch triggers automatic deployment via GitHub Actions.

## Rollback
```bash
make rollback
```