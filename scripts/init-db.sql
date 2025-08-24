-- Database initialization script
-- This script runs when the PostgreSQL container starts for the first time

-- Create the main database if it doesn't exist
SELECT 'CREATE DATABASE bodda'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bodda')\gexec

-- Create test database if it doesn't exist
SELECT 'CREATE DATABASE bodda_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bodda_test')\gexec

-- Create development database if it doesn't exist
SELECT 'CREATE DATABASE bodda_dev'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bodda_dev')\gexec

-- Enable UUID extension for all databases
\c bodda;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c bodda_test;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c bodda_dev;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";