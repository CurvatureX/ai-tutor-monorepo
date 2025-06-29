#!/bin/bash

# AI Tutor Docker Stop Script
# This script helps you stop the AI Tutor services

set -e

echo "🛑 AI Tutor Docker Stop Script"
echo "==============================="

# Navigate to the scripts directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "📁 Working directory: $SCRIPT_DIR"

# Stop and remove containers
echo "🛑 Stopping AI Tutor services..."
docker-compose down --remove-orphans

# Optional: Remove volumes (uncomment if you want to reset data)
# echo "🗑️  Removing volumes..."
# docker-compose down -v

echo ""
echo "✅ AI Tutor services stopped successfully!"
echo ""
echo "📋 Other useful commands:"
echo "   docker system prune              # Clean up Docker system"
echo "   docker-compose down -v           # Stop and remove volumes"
echo "   docker-compose logs <service>    # View logs for specific service" 