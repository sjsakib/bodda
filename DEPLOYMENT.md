# Bodda Deployment Guide

This guide covers deployment options for the Bodda AI coaching application, from local development to production environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Production Deployment](#production-deployment)
- [Environment Configuration](#environment-configuration)
- [Database Management](#database-management)
- [Monitoring and Logging](#monitoring-and-logging)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Docker & Docker Compose**: For containerized deployment
- **Go 1.21+**: For backend development
- **Node.js 18+**: For frontend development
- **PostgreSQL 15+**: For database (or use Docker)
- **Git**: For version control

### Required API Keys

- **Strava API**: Create an application at [Strava Developers](https://developers.strava.com/)
- **OpenAI API**: Get your API key from [OpenAI Platform](https://platform.openai.com/)

## Development Setup

### Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd bodda
   ```

2. **Set up environment**
   ```bash
   # Copy and configure environment file
   cp .env.example .env
   # Edit .env with your API keys and configuration
   
   # Or use the setup script
   ./scripts/setup-env.sh development
   ```

3. **Start development environment**
   ```bash
   # Start all services with hot reloading
   docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
   
   # Or use the Makefile
   make docker-up
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Database Admin (Adminer): http://localhost:8081

### Manual Development Setup

1. **Start PostgreSQL**
   ```bash
   docker-compose up -d postgres
   ```

2. **Set up database**
   ```bash
   ./scripts/db-manage.sh setup-dev
   ```

3. **Start backend**
   ```bash
   go run main.go
   ```

4. **Start frontend** (in another terminal)
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Production Deployment

### Docker Production Deployment

1. **Prepare environment**
   ```bash
   # Create production environment file
   cp .env.production .env
   # Configure all required environment variables
   ```

2. **Deploy with Docker Compose**
   ```bash
   # Build and start production services
   docker-compose -f docker-compose.prod.yml up -d
   
   # Check service status
   docker-compose -f docker-compose.prod.yml ps
   ```

3. **Set up SSL (recommended)**
   ```bash
   # Create SSL certificates directory
   mkdir -p nginx/ssl
   
   # Add your SSL certificates
   # nginx/ssl/cert.pem
   # nginx/ssl/key.pem
   ```

### Cloud Deployment Options

#### Option 1: Docker-based Cloud Deployment

**AWS ECS / Google Cloud Run / Azure Container Instances**

1. Build and push images:
   ```bash
   # Build production images
   docker build -f Dockerfile.backend --target production -t bodda-backend .
   docker build -f frontend/Dockerfile --target production -t bodda-frontend ./frontend
   
   # Tag and push to your registry
   docker tag bodda-backend your-registry/bodda-backend:latest
   docker tag bodda-frontend your-registry/bodda-frontend:latest
   docker push your-registry/bodda-backend:latest
   docker push your-registry/bodda-frontend:latest
   ```

2. Deploy using your cloud provider's container service

#### Option 2: Kubernetes Deployment

1. **Create Kubernetes manifests** (example):
   ```yaml
   # k8s/namespace.yaml
   apiVersion: v1
   kind: Namespace
   metadata:
     name: bodda
   ```

2. **Deploy to cluster**:
   ```bash
   kubectl apply -f k8s/
   ```

## Environment Configuration

### Environment Variables

#### Required Variables

```bash
# Database
DATABASE_URL=postgres://user:password@host:port/database?sslmode=require

# Authentication
JWT_SECRET=your-long-random-jwt-secret

# Strava API
STRAVA_CLIENT_ID=your-strava-client-id
STRAVA_CLIENT_SECRET=your-strava-client-secret
STRAVA_REDIRECT_URL=https://yourdomain.com/auth/callback

# OpenAI API
OPENAI_API_KEY=your-openai-api-key

# Application URLs
FRONTEND_URL=https://yourdomain.com
VITE_API_URL=https://api.yourdomain.com
```

#### Optional Variables

```bash
# Logging
LOG_LEVEL=info  # debug, info, warn, error

# Performance
ENABLE_PROFILING=false
ENABLE_METRICS=true

# Ports
PORT=8080
FRONTEND_PORT=3000
```

### Environment-Specific Configurations

- **Development**: Use `.env.development` for local development
- **Testing**: Use `.env.test` for automated testing
- **Production**: Use `.env.production` or environment variables

## Database Management

### Database Operations

```bash
# Set up development database
./scripts/db-manage.sh setup-dev

# Set up test database
./scripts/db-manage.sh setup-test

# Reset development database
./scripts/db-manage.sh reset-dev

# Run migrations
./scripts/db-manage.sh migrate bodda_prod

# Create backup
./scripts/db-manage.sh backup bodda_prod
```

### Production Database Setup

1. **Create production database**
   ```sql
   CREATE DATABASE bodda_prod;
   CREATE USER bodda_user WITH PASSWORD 'secure_password';
   GRANT ALL PRIVILEGES ON DATABASE bodda_prod TO bodda_user;
   ```

2. **Run migrations**
   ```bash
   DATABASE_URL="postgres://bodda_user:secure_password@host:5432/bodda_prod?sslmode=require" \
   go run main.go -migrate
   ```

### Database Backup Strategy

```bash
# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump $DATABASE_URL > "$BACKUP_DIR/bodda_backup_$DATE.sql"

# Keep only last 7 days of backups
find $BACKUP_DIR -name "bodda_backup_*.sql" -mtime +7 -delete
```

## Monitoring and Logging

### Health Checks

The application provides several monitoring endpoints:

- **Health Check**: `GET /monitoring/health`
- **Application Metrics**: `GET /monitoring/metrics`
- **System Metrics**: `GET /monitoring/system`

### Logging Configuration

```bash
# Development: Structured text logs
LOG_LEVEL=debug

# Production: JSON logs for log aggregation
LOG_LEVEL=warn
```

### Production Monitoring Setup

1. **Log Aggregation** (ELK Stack, Fluentd, etc.)
   ```yaml
   # docker-compose.prod.yml addition
   logging:
     driver: "json-file"
     options:
       max-size: "10m"
       max-file: "3"
   ```

2. **Metrics Collection** (Prometheus, DataDog, etc.)
   - Use `/monitoring/metrics` endpoint
   - Set up alerts for error rates and response times

3. **Uptime Monitoring**
   - Monitor `/monitoring/health` endpoint
   - Set up alerts for service availability

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database connectivity
docker-compose exec postgres psql -U postgres -d bodda_dev -c "SELECT 1;"

# Check database logs
docker-compose logs postgres
```

#### 2. Frontend Build Issues

```bash
# Clear node modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install

# Check for TypeScript errors
npm run type-check
```

#### 3. Backend Issues

```bash
# Check Go module dependencies
go mod tidy
go mod verify

# Run tests
go test ./...

# Check for race conditions
go test -race ./...
```

#### 4. Docker Issues

```bash
# Rebuild containers
docker-compose build --no-cache

# Check container logs
docker-compose logs backend
docker-compose logs frontend

# Clean up Docker resources
docker system prune -a
```

### Performance Optimization

#### Database Optimization

```sql
-- Add indexes for better query performance
CREATE INDEX CONCURRENTLY idx_sessions_user_created 
ON sessions(user_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_messages_session_created 
ON messages(session_id, created_at);
```

#### Application Optimization

```bash
# Enable Go profiling in development
ENABLE_PROFILING=true go run main.go

# Analyze memory usage
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Security Considerations

1. **Environment Variables**: Never commit secrets to version control
2. **Database**: Use SSL connections in production
3. **API Keys**: Rotate keys regularly
4. **HTTPS**: Always use HTTPS in production
5. **CORS**: Configure CORS properly for your domain

### Backup and Recovery

#### Backup Strategy

```bash
# Daily automated backup
0 2 * * * /path/to/scripts/db-manage.sh backup bodda_prod

# Weekly full system backup
0 3 * * 0 tar -czf /backups/bodda_full_$(date +%Y%m%d).tar.gz /app
```

#### Recovery Process

```bash
# Restore from backup
psql $DATABASE_URL < /backups/bodda_backup_20240101_020000.sql

# Verify data integrity
./scripts/db-manage.sh wait bodda_prod
```

## Support

For additional support:

1. Check the application logs: `docker-compose logs`
2. Review the health check endpoint: `/monitoring/health`
3. Consult the troubleshooting section above
4. Check the project's issue tracker

## Version Information

- **Application Version**: Check `/monitoring/health` endpoint
- **Database Schema Version**: Managed by Go migrations
- **Docker Images**: Tagged with semantic versioning