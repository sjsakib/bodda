#!/bin/bash

# Database management script for Bodda application
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}

# Function to wait for database to be ready
wait_for_db() {
    local db_name=$1
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for database $db_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name -c "SELECT 1;" >/dev/null 2>&1; then
            echo "Database $db_name is ready!"
            return 0
        fi
        
        echo "Attempt $attempt/$max_attempts: Database not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: Database $db_name is not ready after $max_attempts attempts"
    return 1
}

# Function to create database if it doesn't exist
create_db() {
    local db_name=$1
    echo "Creating database $db_name if it doesn't exist..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "
        SELECT 'CREATE DATABASE $db_name'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db_name')\\gexec
    " || echo "Database $db_name already exists or creation failed"
}

# Function to run migrations
run_migrations() {
    local db_name=$1
    echo "Running migrations for $db_name..."
    
    # Run the Go migration command
    cd "$PROJECT_ROOT"
    DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$db_name?sslmode=disable" go run main.go -migrate || echo "Migrations completed or already up to date"
}

# Function to seed database
seed_db() {
    local db_name=$1
    local seed_file=$2
    
    if [ -f "$SCRIPT_DIR/$seed_file" ]; then
        echo "Seeding database $db_name with $seed_file..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name -f "$SCRIPT_DIR/$seed_file"
    else
        echo "Seed file $seed_file not found, skipping seeding"
    fi
}

# Function to reset database
reset_db() {
    local db_name=$1
    echo "Resetting database $db_name..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "
        DROP DATABASE IF EXISTS $db_name;
        CREATE DATABASE $db_name;
    "
    
    # Enable UUID extension
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name -c "
        CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";
    "
}

# Function to backup database
backup_db() {
    local db_name=$1
    local backup_file="$PROJECT_ROOT/backups/${db_name}_$(date +%Y%m%d_%H%M%S).sql"
    
    mkdir -p "$PROJECT_ROOT/backups"
    echo "Backing up database $db_name to $backup_file..."
    
    PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name > "$backup_file"
    echo "Backup completed: $backup_file"
}

# Main command handling
case "${1:-help}" in
    "setup-dev")
        echo "Setting up development database..."
        create_db "bodda_dev"
        wait_for_db "bodda_dev"
        run_migrations "bodda_dev"
        seed_db "bodda_dev" "seed-dev-data.sql"
        echo "Development database setup complete!"
        ;;
    
    "setup-test")
        echo "Setting up test database..."
        create_db "bodda_test"
        wait_for_db "bodda_test"
        run_migrations "bodda_test"
        seed_db "bodda_test" "seed-test-data.sql"
        echo "Test database setup complete!"
        ;;
    
    "reset-dev")
        reset_db "bodda_dev"
        run_migrations "bodda_dev"
        seed_db "bodda_dev" "seed-dev-data.sql"
        echo "Development database reset complete!"
        ;;
    
    "reset-test")
        reset_db "bodda_test"
        run_migrations "bodda_test"
        seed_db "bodda_test" "seed-test-data.sql"
        echo "Test database reset complete!"
        ;;
    
    "migrate")
        db_name=${2:-bodda_dev}
        run_migrations "$db_name"
        ;;
    
    "seed")
        db_name=${2:-bodda_dev}
        seed_file=${3:-seed-dev-data.sql}
        seed_db "$db_name" "$seed_file"
        ;;
    
    "backup")
        db_name=${2:-bodda_dev}
        backup_db "$db_name"
        ;;
    
    "wait")
        db_name=${2:-bodda_dev}
        wait_for_db "$db_name"
        ;;
    
    "help"|*)
        echo "Database management script for Bodda"
        echo ""
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  setup-dev          Set up development database with migrations and seed data"
        echo "  setup-test         Set up test database with migrations and test data"
        echo "  reset-dev          Reset and recreate development database"
        echo "  reset-test         Reset and recreate test database"
        echo "  migrate [db_name]  Run migrations on specified database (default: bodda_dev)"
        echo "  seed [db_name] [seed_file]  Seed database with specified file"
        echo "  backup [db_name]   Create backup of specified database"
        echo "  wait [db_name]     Wait for database to be ready"
        echo "  help               Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  DB_HOST            Database host (default: localhost)"
        echo "  DB_PORT            Database port (default: 5432)"
        echo "  DB_USER            Database user (default: postgres)"
        echo "  DB_PASSWORD        Database password (default: password)"
        ;;
esac