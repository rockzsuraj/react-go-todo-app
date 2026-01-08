#!/bin/bash

echo "🚀 Services are starting..."
echo ""
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080/health"
echo "📊 API Endpoints: http://localhost:8080/api/todos"
echo ""
echo "⏳ Waiting for services to be ready..."

# Wait for services to be healthy
while ! curl -s http://localhost:8080/health > /dev/null 2>&1; do
  sleep 2
done

while ! curl -s http://localhost:3000 > /dev/null 2>&1; do
  sleep 2
done

echo ""
echo "✅ All services are ready!"
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080/health"