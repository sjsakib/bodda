#!/bin/bash

# Environment setup script for Bodda application
set -e

ENVIRONMENT=${1:-development}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Setting up environment: $ENVIRONMENT"

# Function to copy environment file
setup_env_file() {
    local env_file=".env.$ENVIRONMENT"
    local target_file=".env"
    
    if [ -f "$PROJECT_ROOT/$env_file" ]; then
        echo "Copying $env_file to $target_file"
        cp "$PROJECT_ROOT/$env_file" "$PROJECT_ROOT/$target_file"
    else
        echo "Warning: $env_file not found, using .env.example as template"
        if [ -f "$PROJECT_ROOT/.env.example" ]; then
            cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/$target_file"
            echo "Please edit .env file with your specific configuration"
        else
            echo "Error: No environment template found"
            exit 1
        fi
    fi
}

# Function to validate required environment variables
validate_env() {
    local required_vars=(
        "DATABASE_URL"
        "JWT_SECRET"
        "STRAVA_CLIENT_ID"
        "STRAVA_CLIENT_SECRET"
        "OPENAI_API_KEY"
    )
    
    echo "Validating environment variables..."
    source "$PROJECT_ROOT/.env"
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ] || [[ "${!var}" == *"your-"* ]] || [[ "${!var}" == *"test-"* ]]; then
            echo "Warning: $var is not properly configured"
        else
            echo "✓ $var is configured"
        fi
    done
}

# Function to generate JWT secret if needed
generate_jwt_secret() {
    if [ "$ENVIRONMENT" != "test" ]; then
        echo "Generating JWT secret..."
        JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
        sed -i.bak "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$PROJECT_ROOT/.env"
        rm -f "$PROJECT_ROOT/.env.bak"
        echo "✓ JWT secret generated"
    fi
}

# Main setup process
main() {
    cd "$PROJECT_ROOT"
    
    case $ENVIRONMENT in
        development|dev)
            setup_env_file
            generate_jwt_secret
            validate_env
            echo "Development environment ready!"
            echo "Run 'make dev' or 'docker-compose -f docker-compose.yml -f docker-compose.dev.yml up' to start"
            ;;
        production|prod)
            echo "Production environment setup"
            echo "Make sure to set all environment variables in your deployment system"
            validate_env
            ;;
        test)
            setup_env_file
            validate_env
            echo "Test environment ready!"
            ;;
        *)
            echo "Usage: $0 [development|production|test]"
            exit 1
            ;;
    esac
}

main "$@"