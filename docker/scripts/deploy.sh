#!/bin/bash

# FieldFuze Backend Deployment Script
# Usage: ./docker/scripts/deploy.sh [environment]
# Environments: dev, staging, prod

set -e

ENVIRONMENT=${1:-dev}
PROJECT_NAME="fieldfuze-backend"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
DOCKER_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$DOCKER_DIR")"

echo "🚀 Deploying FieldFuze Backend - Environment: $ENVIRONMENT"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install it and try again."
    exit 1
fi

# Change to docker directory
cd "$DOCKER_DIR"

# Validate environment
case $ENVIRONMENT in
    dev|development)
        COMPOSE_FILES="-f docker-compose.yml"
        ENV_FILE=".env"
        print_status "Deploying development environment"
        ;;
    staging)
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.override.yml"
        ENV_FILE=".env.staging"
        print_status "Deploying staging environment"
        ;;
    prod|production)
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"
        ENV_FILE=".env.production"
        print_status "Deploying production environment"
        ;;
    *)
        print_error "Invalid environment: $ENVIRONMENT"
        print_error "Valid environments: dev, staging, prod"
        exit 1
        ;;
esac

# Check if environment file exists
if [ ! -f "$ENV_FILE" ]; then
    print_warning "Environment file $ENV_FILE not found."
    if [ -f ".env.example" ]; then
        print_status "Copying .env.example to $ENV_FILE"
        cp .env.example "$ENV_FILE"
        print_warning "Please update $ENV_FILE with your actual configuration values."
    else
        print_error "No .env.example found. Please create $ENV_FILE manually."
        exit 1
    fi
fi

# Load environment variables
set -a
source "$ENV_FILE"
set +a

# Build and deploy
print_status "Building Docker images..."
docker-compose $COMPOSE_FILES build

print_status "Starting services..."
docker-compose $COMPOSE_FILES up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
sleep 30

# Check service health
print_status "Checking service health..."

# Check backend health
if curl -f http://localhost:${NGINX_HTTP_PORT:-9080}/health > /dev/null 2>&1; then
    print_status "✅ Backend is healthy"
else
    print_warning "⚠️  Backend health check failed"
fi

# Check Grafana
if curl -f http://localhost:${NGINX_HTTP_PORT:-9080}/grafana/api/health > /dev/null 2>&1; then
    print_status "✅ Grafana is healthy"
else
    print_warning "⚠️  Grafana health check failed"
fi

# Check Jaeger
if curl -f http://localhost:${NGINX_HTTP_PORT:-9080}/jaeger > /dev/null 2>&1; then
    print_status "✅ Jaeger is healthy"
else
    print_warning "⚠️  Jaeger health check failed"
fi

print_status "Deployment completed!"
print_status ""
print_status "🌐 Access URLs:"
print_status "   Main Application: http://localhost:${NGINX_HTTP_PORT:-9080}/"
print_status "   API Documentation: http://localhost:${NGINX_HTTP_PORT:-9080}/api/v1/auth/docs/"
print_status "   Grafana Dashboard: http://localhost:${NGINX_HTTP_PORT:-9080}/grafana/"
print_status "   Jaeger Tracing: http://localhost:${NGINX_HTTP_PORT:-9080}/jaeger/"
print_status "   Prometheus: http://localhost:${NGINX_HTTP_PORT:-9080}/prometheus/"
print_status ""
print_status "📊 Monitoring:"
print_status "   Grafana Username: admin"
print_status "   Grafana Password: ${GRAFANA_PASSWORD:-admin123}"
print_status ""
print_status "🔧 Management Commands:"
print_status "   View logs: docker-compose $COMPOSE_FILES logs -f"
print_status "   Scale backend: docker-compose $COMPOSE_FILES up -d --scale backend=5"
print_status "   Stop services: docker-compose $COMPOSE_FILES down"
print_status "   Restart services: docker-compose $COMPOSE_FILES restart"