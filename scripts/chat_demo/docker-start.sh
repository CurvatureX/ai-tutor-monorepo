#!/bin/bash

# AI Tutor Docker Startup Script
# This script helps you start the AI Tutor services using Docker Compose

set -e

echo "🚀 AI Tutor Docker Startup Script"
echo "=================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ docker-compose is not installed. Please install it and try again."
    exit 1
fi

# Navigate to the scripts directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "📁 Working directory: $SCRIPT_DIR"

# Check if .env file exists, if not, suggest creating one
if [ ! -f .env ]; then
    echo "⚠️  No .env file found. Using default configuration."
    echo "💡 Tip: Copy .env.example to .env and configure your API keys for full functionality."
    echo ""
fi

# Stop any existing containers
echo "🛑 Stopping any existing containers..."
docker-compose down --remove-orphans

# Build and start services
echo "🔨 Building and starting services..."
docker-compose up --build -d

echo ""
echo "⏳ Waiting for services to be healthy..."

# Wait for conversation service to be healthy
echo "🔍 Checking conversation service health..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose exec -T conversation-service curl -f http://localhost:8000/v1/models -H "Authorization: Bearer dev-token" > /dev/null 2>&1; then
        echo "✅ Conversation service is healthy!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "⏳ Attempt $attempt/$max_attempts - waiting for conversation service..."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Conversation service failed to start properly."
    echo "📋 Checking logs..."
    docker-compose logs conversation-service
    exit 1
fi

# Wait for chat demo to be healthy
echo "🔍 Checking chat demo health..."
max_attempts=15
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose exec -T chat-demo curl -f http://localhost:3001 > /dev/null 2>&1; then
        echo "✅ Chat demo is healthy!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "⏳ Attempt $attempt/$max_attempts - waiting for chat demo..."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Chat demo failed to start properly."
    echo "📋 Checking logs..."
    docker-compose logs chat-demo
    exit 1
fi

echo ""
echo "🎉 AI Tutor services are ready!"
echo "================================"
echo "📍 Chat Demo:           http://localhost:3001"
echo "📍 Conversation API:    http://localhost:8000"
echo "📍 API Documentation:   http://localhost:8000/docs"
echo ""
echo "🔑 Authentication: Use 'dev-token' in the chat demo"
echo "🤖 Available Models: Gemini 2.5 Flash, DeepSeek Chat, Doubao Seed, Local Tutor"
echo ""
echo "📋 Useful Commands:"
echo "   docker-compose logs -f           # View logs"
echo "   docker-compose down              # Stop services"
echo "   docker-compose restart           # Restart services"
echo "   docker-compose ps                # Check service status"
echo ""
echo "🌐 Open http://localhost:3001 in your browser to start chatting!" 