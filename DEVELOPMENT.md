# Bodda Development Guide

This guide covers the development workflow, architecture, and best practices for contributing to the Bodda AI coaching application.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style and Standards](#code-style-and-standards)
- [Database Development](#database-development)
- [API Development](#api-development)
- [Frontend Development](#frontend-development)
- [Debugging](#debugging)

## Development Environment Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Git

### Quick Setup

```bash
# Clone repository
git clone <repository-url>
cd bodda

# Set up development environment
./scripts/setup-env.sh development

# Start development services
make dev
# OR
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

### Manual Setup

```bash
# Install Go dependencies
go mod tidy

# Install frontend dependencies
cd frontend && npm install && cd ..

# Start PostgreSQL
docker-compose up -d postgres

# Set up development database
./scripts/db-manage.sh setup-dev

# Start backend (with hot reloading)
air -c .air.toml

# Start frontend (in another terminal)
cd frontend && npm run dev
```

## Project Structure

```
bodda/
├── .kiro/                    # Kiro IDE specifications
│   └── specs/               # Feature specifications
├── frontend/                # React frontend application
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── services/       # API services
│   │   └── test/           # Frontend tests
│   └── dist/               # Built frontend assets
├── internal/               # Go backend packages
│   ├── config/            # Configuration management
│   ├── database/          # Database layer
│   ├── models/            # Data models
│   ├── monitoring/        # Logging and metrics
│   ├── server/            # HTTP server and routes
│   └── services/          # Business logic services
├── scripts/               # Development and deployment scripts
├── docker-compose*.yml    # Docker configurations
├── Dockerfile.backend     # Backend Docker configuration
└── main.go               # Application entry point
```

### Key Directories

- **`internal/`**: All Go backend code (follows Go project layout)
- **`frontend/src/`**: React application source code
- **`scripts/`**: Utility scripts for development and deployment
- **`.kiro/specs/`**: Feature specifications and implementation plans

## Development Workflow

### 1. Feature Development Process

```bash
# 1. Create feature branch
git checkout -b feature/your-feature-name

# 2. Implement feature following the spec
# - Check .kiro/specs/ for implementation plan
# - Follow test-driven development

# 3. Run tests
make test

# 4. Commit changes
git add .
git commit -m "feat: implement your feature"

# 5. Push and create PR
git push origin feature/your-feature-name
```

### 2. Daily Development Commands

```bash
# Start development environment
make dev

# Run all tests
make test

# Build application
make build

# Clean build artifacts
make clean

# View logs
make docker-logs

# Reset development database
./scripts/db-manage.sh reset-dev
```

### 3. Hot Reloading

- **Backend**: Uses [Air](https://github.com/cosmtrek/air) for automatic recompilation
- **Frontend**: Uses Vite's built-in hot module replacement
- **Database**: Changes require manual migration runs

## Testing

### Backend Testing

```bash
# Run all Go tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/services/...

# Run tests with verbose output
go test -v ./...
```

### Frontend Testing

```bash
cd frontend

# Run unit tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

### Integration Testing

```bash
# Set up test database
./scripts/db-manage.sh setup-test

# Run integration tests
TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/bodda_test?sslmode=disable" \
go test -tags=integration ./...
```

### Test Structure

#### Backend Tests

```go
// Example test structure
func TestServiceMethod(t *testing.T) {
    // Arrange
    service := setupTestService(t)
    
    // Act
    result, err := service.Method(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

#### Frontend Tests

```typescript
// Example component test
describe('Component', () => {
  it('should render correctly', () => {
    render(<Component />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });
});
```

## Code Style and Standards

### Go Code Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for common issues
go vet ./...
```

### TypeScript/React Standards

- Use TypeScript for type safety
- Follow React best practices
- Use ESLint and Prettier for code formatting

```bash
cd frontend

# Format code
npm run format

# Run linter
npm run lint

# Type check
npm run type-check
```

### Commit Message Format

```
type(scope): description

feat(auth): add Strava OAuth integration
fix(chat): resolve message streaming issue
docs(readme): update installation instructions
test(api): add integration tests for sessions
```

## Database Development

### Migration Workflow

```bash
# Create new migration (manual process)
# 1. Add migration to internal/database/migrations.go
# 2. Update the migration version
# 3. Test migration

# Run migrations
./scripts/db-manage.sh migrate bodda_dev

# Reset and re-run migrations
./scripts/db-manage.sh reset-dev
```

### Database Testing

```bash
# Use test database for development
export DATABASE_URL="postgres://postgres:password@localhost:5433/bodda_test?sslmode=disable"

# Run database tests
go test ./internal/database/...
```

### Schema Changes

1. **Add migration** to `internal/database/migrations.go`
2. **Update models** in `internal/models/`
3. **Update repositories** in `internal/database/`
4. **Add tests** for new functionality
5. **Update seed data** if necessary

## API Development

### Adding New Endpoints

1. **Define route** in `internal/server/server.go`
2. **Implement handler** in appropriate service
3. **Add middleware** if needed
4. **Write tests** for the endpoint
5. **Update API documentation**

### API Testing

```bash
# Test endpoints with curl
curl -X GET http://localhost:8080/monitoring/health

# Use httpie for better formatting
http GET localhost:8080/api/sessions Authorization:"Bearer $TOKEN"
```

### Authentication Testing

```bash
# Get test token (in development)
TOKEN=$(curl -s -X POST http://localhost:8080/auth/test-token | jq -r '.token')

# Use token in requests
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sessions
```

## Frontend Development

### Component Development

```bash
cd frontend

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### State Management

- Use React hooks for local state
- Use context for shared state
- Keep state close to where it's used

### API Integration

```typescript
// Use the custom useApi hook
const { data, loading, error } = useApi('/api/sessions');

// Handle loading and error states
if (loading) return <LoadingSpinner />;
if (error) return <ErrorMessage error={error} />;
```

## Debugging

### Backend Debugging

#### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start with debugger
dlv debug main.go

# Set breakpoints and debug
(dlv) break main.main
(dlv) continue
```

#### Logging

```go
// Use structured logging
import "bodda/internal/monitoring"

logger := monitoring.GetLogger()
logger.Info("Processing request", "user_id", userID, "action", "create_session")
```

#### Profiling

```bash
# Enable profiling
ENABLE_PROFILING=true go run main.go

# Analyze CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Analyze memory profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Frontend Debugging

#### Browser DevTools

- Use React Developer Tools extension
- Check Network tab for API calls
- Use Console for JavaScript debugging

#### Debugging Hooks

```typescript
// Add debug logging to custom hooks
const useApi = (url: string) => {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    console.log('API call:', url); // Debug log
    // ... rest of hook
  }, [url]);
  
  return { data, loading, error };
};
```

### Database Debugging

```bash
# Connect to development database
docker-compose exec postgres psql -U postgres -d bodda_dev

# Check query performance
EXPLAIN ANALYZE SELECT * FROM sessions WHERE user_id = 'some-id';

# Monitor database logs
docker-compose logs -f postgres
```

### Common Debugging Scenarios

#### 1. API Not Responding

```bash
# Check if backend is running
curl http://localhost:8080/monitoring/health

# Check backend logs
docker-compose logs backend

# Check database connection
docker-compose exec postgres psql -U postgres -c "SELECT 1;"
```

#### 2. Frontend Not Loading

```bash
# Check frontend logs
docker-compose logs frontend

# Check if API URL is correct
echo $VITE_API_URL

# Test API connectivity from frontend container
docker-compose exec frontend curl http://backend:8080/monitoring/health
```

#### 3. Database Issues

```bash
# Check database status
docker-compose ps postgres

# Reset database
./scripts/db-manage.sh reset-dev

# Check for migration issues
go run main.go -migrate
```

## Performance Optimization

### Backend Performance

```bash
# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Profile memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Check for goroutine leaks
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### Frontend Performance

```bash
# Analyze bundle size
cd frontend
npm run build
npm run analyze

# Check for unused dependencies
npx depcheck
```

### Database Performance

```sql
-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation 
FROM pg_stats 
WHERE tablename = 'sessions';
```

## Contributing Guidelines

1. **Follow the spec**: Check `.kiro/specs/` for implementation guidance
2. **Write tests**: Maintain test coverage above 80%
3. **Document changes**: Update relevant documentation
4. **Review code**: All changes require code review
5. **Test thoroughly**: Test in development environment before PR

### Pull Request Checklist

- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`go fmt`, `npm run format`)
- [ ] No linting errors
- [ ] Documentation updated
- [ ] Feature works in development environment
- [ ] Database migrations tested (if applicable)

## Useful Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://reactjs.org/docs/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Vite Documentation](https://vitejs.dev/guide/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)