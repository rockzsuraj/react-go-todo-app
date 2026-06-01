#!/bin/bash

# Development Docker setup with hot reload
echo "🚀 Starting development environment with hot reload..."

# Set environment variables for development
export ENV=development
export NODE_ENV=development

# Copy .env if it doesn't exist
if [ ! -f .env ]; then
    echo "📋 Creating .env file from template..."
    cp .env.example .env
fi

# Start services with hot reload
echo "🔥 Starting services with hot reload enabled..."
docker-compose up --build --force-recreate
