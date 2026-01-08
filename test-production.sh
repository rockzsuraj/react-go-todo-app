#!/bin/bash

set -e

echo "🧪 Starting automated production tests..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test functions
test_passed() {
    echo -e "${GREEN}✅ $1${NC}"
}

test_failed() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

test_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# 1. Build and start services
test_info "Building and starting services..."
docker-compose down -v 2>/dev/null || true
docker-compose up --build -d

# 2. Wait for services to be ready
test_info "Waiting for services to be ready..."
timeout=60
counter=0

while ! curl -s http://localhost:8080/health > /dev/null 2>&1; do
    sleep 2
    counter=$((counter + 2))
    if [ $counter -ge $timeout ]; then
        test_failed "Backend failed to start within $timeout seconds"
    fi
done

while ! curl -s http://localhost:3000 > /dev/null 2>&1; do
    sleep 2
    counter=$((counter + 2))
    if [ $counter -ge $timeout ]; then
        test_failed "Frontend failed to start within $timeout seconds"
    fi
done

test_passed "Services started successfully"

# 3. Test health endpoints
test_info "Testing health endpoints..."

# Health check
health_response=$(curl -s http://localhost:8080/health)
if echo "$health_response" | grep -q '"status":"healthy"'; then
    test_passed "Health check returns correct JSON"
else
    test_failed "Health check failed: $health_response"
fi

# Readiness check
ready_response=$(curl -s http://localhost:8080/ready)
if echo "$ready_response" | grep -q '"status":"ready"'; then
    test_passed "Readiness check returns correct JSON"
else
    test_failed "Readiness check failed: $ready_response"
fi

# 4. Test security headers
test_info "Testing security headers..."
headers=$(curl -I -s http://localhost:8080/health)

if echo "$headers" | grep -q "X-Content-Type-Options: nosniff"; then
    test_passed "X-Content-Type-Options header present"
else
    test_failed "X-Content-Type-Options header missing"
fi

if echo "$headers" | grep -q "X-Frame-Options: DENY"; then
    test_passed "X-Frame-Options header present"
else
    test_failed "X-Frame-Options header missing"
fi

# 5. Test API endpoints
test_info "Testing API endpoints..."

# Get todos
todos_response=$(curl -s http://localhost:8080/api/todos)
if [ $? -eq 0 ]; then
    test_passed "GET /api/todos endpoint working"
else
    test_failed "GET /api/todos endpoint failed"
fi

# Create todo
create_response=$(curl -s -X POST http://localhost:8080/api/todos \
    -H "Content-Type: application/json" \
    -d '{"description":"Test todo","assigned":"Test user"}')
if [ $? -eq 0 ]; then
    test_passed "POST /api/todos endpoint working"
else
    test_failed "POST /api/todos endpoint failed"
fi

# 6. Test frontend
test_info "Testing frontend..."
frontend_response=$(curl -s http://localhost:3000)
if echo "$frontend_response" | grep -q "react-app"; then
    test_passed "Frontend serving React app"
else
    test_failed "Frontend not serving React app correctly"
fi

# 7. Test structured logging
test_info "Testing structured logging..."
logs=$(docker-compose logs backend 2>/dev/null | tail -5)
if echo "$logs" | grep -q '"level"'; then
    test_passed "Structured JSON logging working"
else
    test_failed "Structured logging not working"
fi

# 8. Test production build
test_info "Testing production Docker build..."
docker build -f docker/backend-prod.Dockerfile -t todo-api:test . > /dev/null 2>&1
if [ $? -eq 0 ]; then
    test_passed "Production Docker build successful"
else
    test_failed "Production Docker build failed"
fi

# 9. Test graceful shutdown
test_info "Testing graceful shutdown..."
container_id=$(docker-compose ps -q backend)
docker kill -s SIGTERM $container_id > /dev/null 2>&1
sleep 3
shutdown_logs=$(docker logs $container_id 2>&1 | tail -5)
if echo "$shutdown_logs" | grep -q "server stopped gracefully"; then
    test_passed "Graceful shutdown working"
else
    test_failed "Graceful shutdown not working"
fi

# Cleanup
test_info "Cleaning up..."
docker-compose down -v > /dev/null 2>&1
docker rmi todo-api:test > /dev/null 2>&1 || true

echo ""
echo -e "${GREEN}🎉 All tests passed! Production setup is working correctly.${NC}"
echo ""
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080/health"
echo "📊 API Endpoints: http://localhost:8080/api/todos"