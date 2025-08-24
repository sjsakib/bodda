# Bodda - AI Running & Cycling Coach

Bodda is an AI-powered coaching application that integrates with Strava to provide personalized training advice for runners and cyclists.

## Features

- **Strava Integration**: OAuth authentication and activity data access
- **AI Coaching**: Personalized advice powered by OpenAI with function calling
- **Conversation History**: Persistent chat sessions with full message history
- **Athlete Logbook**: Evolving profile that learns from each interaction
- **Real-time Streaming**: Live AI responses using Server-Sent Events
- **Responsive Design**: Modern web interface built with React and Tailwind CSS

## Tech Stack

**Backend:**
- Go 1.21+ with Gin web framework
- PostgreSQL database with migrations
- Strava API v3 integration
- OpenAI API with function calling
- JWT authentication
- Structured logging and metrics

**Frontend:**
- React 18 with TypeScript
- Vite build tool and hot reloading
- Tailwind CSS for styling
- React Router for navigation
- Server-Sent Events for real-time streaming

**Infrastructure:**
- Docker and Docker Compose
- Multi-stage Dockerfiles for development and production
- Nginx reverse proxy for production
- Environment-specific configurations

## Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone repository
git clone <repository-url>
cd bodda

# Set up environment
./scripts/setup-env.sh development

# Start development environment
make dev-docker
```

### Option 2: Manual Setup

```bash
# Set up environment
make setup-env

# Install dependencies
make install-deps

# Start development services
make dev
```

### Access Points

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/monitoring/health
- **Database Admin**: http://localhost:8081 (Adminer)

## Development

### Available Commands

```bash
# Development
make dev              # Start development environment (manual)
make dev-docker       # Start with Docker (hot reloading)
make setup-env        # Set up development environment

# Testing
make test             # Run all tests
make test-integration # Run integration tests
make test-coverage    # Run tests with coverage

# Database
make db-setup         # Set up development database
make db-reset         # Reset development database
make migrate          # Run database migrations
make seed             # Seed development data

# Code Quality
make lint             # Run linters
make format           # Format code

# Docker
make docker-up        # Start all Docker services
make docker-down      # Stop Docker services
make docker-logs      # View Docker logs
make docker-clean     # Clean Docker resources

# Production
make build-docker     # Build production Docker images
make deploy-prod      # Deploy to production
```

### Project Structure

```
bodda/
├── .kiro/specs/           # Feature specifications
├── frontend/              # React frontend
├── internal/              # Go backend packages
│   ├── config/           # Configuration management
│   ├── database/         # Database layer
│   ├── models/           # Data models
│   ├── monitoring/       # Logging and metrics
│   ├── server/           # HTTP server
│   └── services/         # Business logic
├── scripts/              # Development scripts
├── nginx/                # Nginx configuration
└── docker-compose*.yml   # Docker configurations
```

## Environment Configuration

### Required Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:password@host:port/database

# Authentication
JWT_SECRET=your-long-random-jwt-secret

# Strava API (get from https://developers.strava.com/)
STRAVA_CLIENT_ID=your-strava-client-id
STRAVA_CLIENT_SECRET=your-strava-client-secret
STRAVA_REDIRECT_URL=http://localhost:8080/auth/callback

# OpenAI API (get from https://platform.openai.com/)
OPENAI_API_KEY=your-openai-api-key

# Application URLs
FRONTEND_URL=http://localhost:5173
VITE_API_URL=http://localhost:8080
```

### Environment Files

- `.env.development` - Development configuration
- `.env.production` - Production configuration  
- `.env.test` - Test configuration
- `.env.example` - Template with all variables

## Deployment

### Production Deployment

```bash
# Set up production environment
make setup-env-prod

# Deploy with Docker Compose
make deploy-prod

# Or manually
docker-compose -f docker-compose.prod.yml up -d
```

### Health Monitoring

The application provides monitoring endpoints:

- **Health Check**: `GET /monitoring/health`
- **Application Metrics**: `GET /monitoring/metrics`  
- **System Metrics**: `GET /monitoring/system`

### Database Management

```bash
# Set up production database
./scripts/db-manage.sh setup-prod

# Create backup
./scripts/db-manage.sh backup bodda_prod

# Run migrations
./scripts/db-manage.sh migrate bodda_prod
```

## API Documentation

### Authentication Endpoints
- `GET /auth/strava` - Initiate Strava OAuth flow
- `GET /auth/callback` - Handle OAuth callback
- `GET /api/auth/check` - Check authentication status
- `POST /auth/logout` - Logout user

### Session Management
- `GET /api/sessions` - Get user's conversation sessions
- `POST /api/sessions` - Create new session
- `GET /api/sessions/:id/messages` - Get session messages

### Chat Interface
- `POST /api/sessions/:id/messages` - Send message to AI coach
- `GET /api/sessions/:id/stream` - Server-Sent Events for streaming responses

### Monitoring
- `GET /monitoring/health` - Application health status
- `GET /monitoring/metrics` - Application metrics
- `GET /monitoring/system` - System metrics

## Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User accounts and Strava authentication tokens
- `sessions` - Conversation sessions with titles and metadata
- `messages` - Chat messages with role (user/assistant) and timestamps
- `athlete_logbooks` - Evolving athlete profiles and coaching insights

## Architecture

### Backend Services
- **AuthService**: Handles Strava OAuth and JWT authentication
- **ChatService**: Manages conversation sessions and messages
- **StravaService**: Integrates with Strava API for activity data
- **AIService**: Processes messages through OpenAI with function calling
- **LogbookService**: Manages athlete profiles and training insights

### AI Function Calling
The AI coach has access to several tools:
- `get-athlete-profile` - Fetch Strava athlete profile
- `get-recent-activities` - Get recent training activities
- `get-activity-details` - Get detailed activity information
- `get-activity-streams` - Get activity time-series data
- `update-athlete-logbook` - Update athlete profile and insights

## Development Resources

- **[Development Guide](DEVELOPMENT.md)** - Detailed development workflow
- **[Deployment Guide](DEPLOYMENT.md)** - Production deployment instructions
- **[Feature Specs](.kiro/specs/)** - Implementation specifications

## Contributing

1. Check the [Development Guide](DEVELOPMENT.md) for setup instructions
2. Review existing [feature specifications](.kiro/specs/)
3. Follow the test-driven development approach
4. Ensure all tests pass: `make test`
5. Submit a pull request with clear description

### Code Quality Standards

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **TypeScript**: Use strict type checking and ESLint rules
- **Testing**: Maintain >80% test coverage
- **Documentation**: Update relevant docs with changes

## License

This project is licensed under the MIT License.