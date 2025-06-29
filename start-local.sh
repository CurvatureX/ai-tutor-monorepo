#!/bin/bash

# Local development startup script for voice practice services

set -e

echo "🚀 Starting Voice Practice Services Locally"
echo "=========================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Creating from template..."
    cp .env.local .env
    echo "📝 Please edit .env file with your API keys before running services"
    echo "   Required: ASR_ACCESS_KEY, ASR_APP_KEY, LLM_API_KEY, TTS_APP_ID, TTS_TOKEN"
    echo ""
fi

# Function to check if port is available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "❌ Port $1 is already in use"
        exit 1
    fi
}

# Check required ports
echo "🔍 Checking ports..."
check_port 8080
check_port 50051

# Build services
echo "📦 Building services..."

# Build Speech Service
echo "  Building Speech Service..."
cd services/speech-service
if ! go build -o speech-service ./cmd; then
    echo "❌ Failed to build Speech Service"
    exit 1
fi
cd ../..

# Build Gateway
echo "  Building Gateway..."
cd gateway
if ! go build -o gateway .; then
    echo "❌ Failed to build Gateway"
    exit 1
fi
cd ..

echo "✅ Build completed successfully"

# Start services
echo "🚀 Starting services..."

# Start Speech Service in background
echo "  Starting Speech Service on port 50051..."
cd services/speech-service
./speech-service > ../../speech-service.log 2>&1 &
SPEECH_PID=$!
echo $SPEECH_PID > ../../speech-service.pid
cd ../..

# Wait a moment for Speech Service to start
sleep 2

# Check if Speech Service is running
if ! kill -0 $SPEECH_PID 2>/dev/null; then
    echo "❌ Speech Service failed to start. Check speech-service.log"
    exit 1
fi

# Start Gateway in background
echo "  Starting Gateway on port 8080..."
cd gateway
./gateway > ../gateway.log 2>&1 &
GATEWAY_PID=$!
echo $GATEWAY_PID > ../gateway.pid
cd ..

# Wait a moment for Gateway to start
sleep 2

# Check if Gateway is running
if ! kill -0 $GATEWAY_PID 2>/dev/null; then
    echo "❌ Gateway failed to start. Check gateway.log"
    # Clean up Speech Service
    kill $SPEECH_PID 2>/dev/null || true
    exit 1
fi

echo ""
echo "🎉 Services started successfully!"
echo ""
echo "📋 Service Status:"
echo "  🎙️  Speech Service: http://localhost:50051 (PID: $SPEECH_PID)"
echo "  🌐 Gateway: http://localhost:8080 (PID: $GATEWAY_PID)"
echo ""
echo "🧪 Test URLs:"
echo "  • WebSocket Test: http://localhost:8080"
echo "  • Health Check: http://localhost:8080/health"
echo "  • Readiness Check: http://localhost:8080/ready"
echo ""
echo "📝 Logs:"
echo "  • Gateway: tail -f gateway.log"
echo "  • Speech Service: tail -f speech-service.log"
echo ""
echo "🛑 To stop services: ./stop-local.sh"
echo ""
echo "🔧 If you need to configure API keys, edit the .env file and restart services"