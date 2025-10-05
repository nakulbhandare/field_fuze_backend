# FieldFuze Backend Docker Deployment

This directory contains all the Docker configuration files for deploying the FieldFuze Backend with comprehensive monitoring, logging, and tracing.

## 🏗️ Architecture

### Services Stack
- **Backend**: Go application with horizontal scaling (2-5 replicas)
- **Nginx**: Reverse proxy with load balancing and SSL termination
- **Prometheus**: Metrics collection and monitoring
- **Grafana**: Dashboards and visualization
- **Loki**: Log aggregation and search
- **Promtail**: Log shipping agent
- **Jaeger**: Distributed tracing
- **Node Exporter**: System metrics
- **cAdvisor**: Container metrics

### Single URL Access
All services are accessible through a single URL via Nginx reverse proxy:
- `http://localhost:9080/` - Main application
- `http://localhost:9080/api/` - API endpoints
- `http://localhost:9080/grafana/` - Monitoring dashboards
- `http://localhost:9080/jaeger/` - Distributed tracing
- `http://localhost:9080/prometheus/` - Metrics (optional)

## 🚀 Quick Start

### 1. Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit with your configuration
nano .env
```

### 2. Deploy Development Environment
```bash
# Using the deployment script (recommended)
./scripts/deploy.sh dev

# Or manually
docker-compose up -d
```

### 3. Deploy Staging Environment
```bash
./scripts/deploy.sh staging
```

### 4. Deploy Production Environment
```bash
./scripts/deploy.sh prod
```

## 📊 Monitoring & Observability

### Grafana Dashboards
- **Backend Monitoring**: API metrics, response times, error rates
- **Infrastructure Monitoring**: CPU, memory, disk, network usage
- **Custom Dashboards**: Available in `docker/grafana/dashboards/`

### Prometheus Metrics
- Application metrics from Go backend
- System metrics from Node Exporter
- Container metrics from cAdvisor
- Custom alert rules in `docker/prometheus/rules/`

### Jaeger Tracing
- Distributed tracing for API requests
- Performance bottleneck identification
- Request flow visualization

### Loki Logging
- Centralized log aggregation
- Real-time log search and filtering
- Integration with Grafana for log visualization

## ⚡ Scaling

### Horizontal Scaling
```bash
# Scale backend to 5 replicas
./scripts/scale.sh backend 5

# Scale nginx to 3 replicas
./scripts/scale.sh nginx 3
```

### Vertical Scaling
Edit resource limits in docker-compose files:
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1G
    reservations:
      cpus: '1.0'
      memory: 512M
```

## 🔧 Configuration Files

### Core Configuration
- `docker-compose.yml` - Base configuration
- `docker-compose.override.yml` - Development overrides
- `docker-compose.prod.yml` - Production configuration

### Service Configuration
- `nginx/nginx.conf` - Reverse proxy and load balancing
- `prometheus/prometheus.yml` - Metrics collection
- `grafana/provisioning/` - Dashboard and datasource provisioning
- `loki/loki.yml` - Log aggregation configuration
- `jaeger/jaeger-config.yml` - Tracing configuration

## 🛡️ Security Features

### Resource Limits
- CPU and memory limits for all containers
- Automatic restart policies
- Health checks for critical services

### Network Security
- Isolated Docker network
- Rate limiting on API endpoints
- Security headers in Nginx

### Monitoring Alerts
- High CPU/memory usage alerts
- Service down notifications
- HTTP error rate monitoring
- Container resource exhaustion alerts

## 📈 Performance Optimization

### Caching
- Nginx response caching
- Prometheus metric caching
- Grafana dashboard caching

### Load Balancing
- Least connection algorithm
- Health check integration
- Automatic failover

### Resource Management
- Optimized Docker images (multi-stage builds)
- Resource reservations and limits
- Efficient log rotation

## 🔍 Troubleshooting

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend

# Real-time monitoring
docker stats
```

### Health Checks
```bash
# Check service health
curl http://localhost:9080/health

# Check individual services
docker-compose ps
```

### Service Management
```bash
# Restart specific service
docker-compose restart backend

# Rebuild and restart
docker-compose up -d --build backend

# Stop all services
docker-compose down
```

## 🌐 Environment-Specific URLs

### Development (Port 9080)
- Application: http://localhost:9080/
- Grafana: http://localhost:9080/grafana/ (admin/admin123)
- Jaeger: http://localhost:9080/jaeger/
- API Docs: http://localhost:9080/api/v1/auth/docs/

### Production (Standard Ports)
- Application: http://your-domain.com/
- Grafana: http://your-domain.com/grafana/
- Jaeger: http://your-domain.com/jaeger/
- API Docs: http://your-domain.com/api/v1/auth/docs/

## 📝 Best Practices

1. **Always use environment files** for configuration
2. **Monitor resource usage** regularly
3. **Set up proper alerts** for production
4. **Use SSL certificates** in production
5. **Regular backup** of Grafana dashboards
6. **Keep images updated** for security
7. **Test scaling** before production deployment

## 🔄 CI/CD Integration

This setup is designed to work with CI/CD pipelines:
- Environment-specific configurations
- Automated deployment scripts
- Health check endpoints
- Graceful shutdown handling