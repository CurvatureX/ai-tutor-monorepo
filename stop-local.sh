#!/bin/bash

# Local development stop script for voice practice services

echo "🛑 Stopping Voice Practice Services"
echo "================================="

# Function to stop service by PID file
stop_service() {
    local service_name=$1
    local pid_file=$2
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            echo "  Stopping $service_name (PID: $pid)..."
            kill "$pid"
            
            # Wait for graceful shutdown
            local count=0
            while kill -0 "$pid" 2>/dev/null && [ $count -lt 10 ]; do
                sleep 1
                ((count++))
            done
            
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                echo "    Force killing $service_name..."
                kill -9 "$pid" 2>/dev/null || true
            fi
            
            echo "  ✅ $service_name stopped"
        else
            echo "  ⚠️  $service_name was not running"
        fi
        rm -f "$pid_file"
    else
        echo "  ⚠️  No PID file found for $service_name"
    fi
}

# Stop Gateway
stop_service "Gateway" "gateway.pid"

# Stop Speech Service
stop_service "Speech Service" "speech-service.pid"

# Clean up log files (optional)
read -p "🗑️  Delete log files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f gateway.log speech-service.log
    echo "  ✅ Log files deleted"
fi

echo ""
echo "🎉 All services stopped successfully!"
echo ""
echo "💡 To restart services: ./start-local.sh"