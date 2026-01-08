#!/bin/bash

echo "🚀 Production Deployment Script"
echo "================================"

# Check prerequisites
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ Set GITHUB_TOKEN environment variable"
    exit 1
fi

if [ -z "$KUBE_CONFIG" ]; then
    echo "❌ Set KUBE_CONFIG environment variable (base64 encoded)"
    exit 1
fi

# Setup kubeconfig
echo "$KUBE_CONFIG" | base64 -d > prod-kubeconfig.yaml
export KUBECONFIG=prod-kubeconfig.yaml

# Login to registry
echo "$GITHUB_TOKEN" | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin

# Build and push images
echo "📦 Building and pushing images..."
make push-images \
    API_URL=${API_URL:-https://api.yourdomain.com} \
    SUPABASE_URL=${SUPABASE_URL} \
    SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}

# Deploy to production
echo "🚀 Deploying to production..."
make deploy

# Check status
echo "📊 Checking deployment status..."
make status

echo "✅ Production deployment complete!"
echo "🌐 Access your app at: https://yourdomain.com"