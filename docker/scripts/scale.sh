#!/bin/bash

# FieldFuze Backend Auto-Scaling Script
# Usage: ./scripts/scale.sh [service] [replicas]

set -e

SERVICE=${1:-backend}
REPLICAS=${2:-3}

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate inputs
if ! [[ "$REPLICAS" =~ ^[0-9]+$ ]] || [ "$REPLICAS" -lt 1 ] || [ "$REPLICAS" -gt 10 ]; then
    print_error "Invalid replica count: $REPLICAS (must be between 1-10)"
    exit 1
fi

# Check current status
print_status "Current service status:"
docker-compose ps

# Scale the service
print_status "Scaling $SERVICE to $REPLICAS replicas..."
docker-compose up -d --scale $SERVICE=$REPLICAS

# Wait for scaling to complete
print_status "Waiting for scaling to complete..."
sleep 10

# Show new status
print_status "New service status:"
docker-compose ps

# Check resource usage
print_status "Resource usage:"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

print_status "✅ Scaling completed successfully!"