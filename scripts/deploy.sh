#!/bin/bash

set -e

NAMESPACE="todo-app"
FRONTEND_IMAGE=""
BACKEND_IMAGE=""

deploy() {
    echo "🚀 Deploying to Kubernetes..."
    
    if [[ -n "$FRONTEND_IMAGE" ]]; then
        sed -i.bak "s|FRONTEND_IMAGE|$FRONTEND_IMAGE|g" k8s/frontend-deployment.yaml
    fi
    
    if [[ -n "$BACKEND_IMAGE" ]]; then
        sed -i.bak "s|BACKEND_IMAGE|$BACKEND_IMAGE|g" k8s/backend-deployment.yaml
    fi
    
    kubectl apply -f k8s/
    
    echo "⏳ Waiting for deployments..."
    kubectl rollout status deployment/todo-frontend -n $NAMESPACE --timeout=300s
    kubectl rollout status deployment/todo-backend -n $NAMESPACE --timeout=300s
    
    # Restore original files
    [[ -f k8s/frontend-deployment.yaml.bak ]] && mv k8s/frontend-deployment.yaml.bak k8s/frontend-deployment.yaml
    [[ -f k8s/backend-deployment.yaml.bak ]] && mv k8s/backend-deployment.yaml.bak k8s/backend-deployment.yaml
    
    echo "✅ Deployment completed successfully!"
}

rollback() {
    echo "🔄 Rolling back deployments..."
    kubectl rollout undo deployment/todo-frontend -n $NAMESPACE
    kubectl rollout undo deployment/todo-backend -n $NAMESPACE
    
    kubectl rollout status deployment/todo-frontend -n $NAMESPACE
    kubectl rollout status deployment/todo-backend -n $NAMESPACE
    echo "✅ Rollback completed!"
}

status() {
    echo "📊 Deployment Status:"
    kubectl get pods -n $NAMESPACE
    echo ""
    kubectl get services -n $NAMESPACE
    echo ""
    kubectl get ingress -n $NAMESPACE
}

# Parse arguments
COMMAND="$1"
for arg in "$@"; do
    case $arg in
        --frontend-image=*)
            FRONTEND_IMAGE="${arg#*=}"
            ;;
        --backend-image=*)
            BACKEND_IMAGE="${arg#*=}"
            ;;
    esac
done

case $COMMAND in
    deploy)
        deploy
        ;;
    rollback)
        rollback
        ;;
    status)
        status
        ;;
    *)
        echo "Usage: $0 [deploy|rollback|status] [--frontend-image=IMAGE] [--backend-image=IMAGE]"
        exit 1
        ;;
esac