#!/bin/bash

# Local development restart script for voice practice services

set -e

echo "🔄 Restarting Voice Practice Services"
echo "====================================="

# Function to check if script exists and is executable
check_script() {
    local script_name=$1
    if [ ! -f "$script_name" ]; then
        echo "❌ $script_name not found"
        exit 1
    fi
    if [ ! -x "$script_name" ]; then
        echo "🔧 Making $script_name executable..."
        chmod +x "$script_name"
    fi
}

# Check if required scripts exist
echo "🔍 Checking required scripts..."
check_script "stop-local.sh"
check_script "start-local.sh"

# Stop services first
echo ""
echo "🛑 Stopping existing services..."
./stop-local.sh

# Wait a moment to ensure services are fully stopped
echo ""
echo "⏳ Waiting for services to fully stop..."
sleep 3

# Check if any services are still running on our ports
check_port_clean() {
    local port=$1
    local service_name=$2
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "⚠️  $service_name still running on port $port"
        echo "   Waiting for it to stop..."
        sleep 2
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "❌ $service_name did not stop properly on port $port"
            echo "   You may need to manually kill the process:"
            echo "   sudo lsof -ti:$port | xargs kill -9"
            exit 1
        fi
    fi
}

echo "🔍 Verifying ports are free..."
check_port_clean 8080 "Gateway"
check_port_clean 50051 "Speech Service"

# Clean up any remaining PID files
echo "🧹 Cleaning up any remaining PID files..."
rm -f gateway.pid speech-service.pid

# Start services
echo ""
echo "🚀 Starting services..."
./start-local.sh

echo ""
echo "🎉 Restart completed successfully!"
echo ""
echo "📊 Quick Status Check:"
echo "  🌐 Gateway: http://localhost:8080"
echo "  🎙️  Speech Service: gRPC on localhost:50051"
echo ""
echo "🧪 Test the services:"
echo "  curl http://localhost:8080/health"
echo "  open http://localhost:8080"
echo ""
echo "📝 Monitor logs:"
echo "  tail -f gateway.log"
echo "  tail -f speech-service.log" 