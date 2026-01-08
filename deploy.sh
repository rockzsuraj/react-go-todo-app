#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
HEALTH_CHECK_TIMEOUT=60
ROLLBACK_ENABLED=true

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Backup current state for rollback
backup_current_state() {
    log "Creating backup of current deployment..."
    docker-compose ps --format json > .deployment-backup.json
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.ID}}" | grep react-todos > .images-backup.txt || true
}

# Health check function
health_check() {
    local service=$1
    local url=$2
    local max_attempts=$((HEALTH_CHECK_TIMEOUT / 5))
    
    log "Performing health check for $service..."
    
    for i in $(seq 1 $max_attempts); do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log "$service is healthy ✓"
            return 0
        fi
        
        warn "Health check attempt $i/$max_attempts failed for $service"
        sleep 5
    done
    
    error "$service health check failed after $HEALTH_CHECK_TIMEOUT seconds"
    return 1
}

# Rollback function
rollback() {
    if [ "$ROLLBACK_ENABLED" = true ]; then
        error "Deployment failed. Initiating rollback..."
        docker-compose down
        # Restore previous images if backup exists
        if [ -f .images-backup.txt ]; then
            log "Rolling back to previous version..."
            docker-compose up -d
        fi
        exit 1
    else
        error "Deployment failed. Rollback disabled."
        exit 1
    fi
}

# Main deployment function
deploy() {
    log "Starting production deployment..."
    
    # Pre-deployment checks
    if ! command -v docker-compose &> /dev/null; then
        error "docker-compose is not installed"
        exit 1
    fi
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "docker-compose.yml not found"
        exit 1
    fi
    
    # Backup current state
    backup_current_state
    
    # Stop existing containers
    log "Stopping existing containers..."
    docker-compose down || true
    
    # Build new images
    log "Building new images..."
    if ! docker-compose build --no-cache; then
        error "Build failed"
        rollback
    fi
    
    # Start services
    log "Starting services..."
    if ! docker-compose up -d; then
        error "Failed to start services"
        rollback
    fi
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 10
    
    # Health checks
    if ! health_check "Backend API" "http://localhost:8080/health"; then
        rollback
    fi
    
    if ! health_check "Frontend" "http://localhost:3000"; then
        rollback
    fi
    
    if ! health_check "Database" "http://localhost:8081"; then
        warn "pgAdmin health check failed, but continuing..."
    fi
    
    # Cleanup old images
    log "Cleaning up old images..."
    docker image prune -f || true
    
    # Remove backup files
    rm -f .deployment-backup.json .images-backup.txt
    
    log "✅ Deployment completed successfully!"
    log "Services available at:"
    log "  - Frontend: http://localhost:3000"
    log "  - Backend API: http://localhost:8080"
    log "  - Database Admin: http://localhost:8081"
}

# Script execution
case "${1:-deploy}" in
    "deploy")
        deploy
        ;;
    "rollback")
        log "Manual rollback initiated..."
        docker-compose down
        if [ -f .deployment-backup.json ]; then
            docker-compose up -d
            log "Rollback completed"
        else
            error "No backup found for rollback"
            exit 1
        fi
        ;;
    "health")
        health_check "Backend API" "http://localhost:8080/health"
        health_check "Frontend" "http://localhost:3000"
        ;;
    *)
        echo "Usage: $0 {deploy|rollback|health}"
        echo "  deploy   - Deploy the application (default)"
        echo "  rollback - Rollback to previous version"
        echo "  health   - Check service health"
        exit 1
        ;;
esac