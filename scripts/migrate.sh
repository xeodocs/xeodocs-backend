#!/bin/bash

# DB migration script
echo "Running database migrations..."

# Auth service migrations
psql $DATABASE_URL -c "
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
" || echo "Migration for auth failed or already exists"

echo "Migrations completed."
