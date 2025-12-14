-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    source_website_url VARCHAR(255),
    source_repo_url VARCHAR(255) NOT NULL,
    source_branch VARCHAR(255) DEFAULT 'main',
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE languages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(50) NOT NULL,
    domain VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, code)
);

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    checksum_original VARCHAR(64),
    checksum_translated VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    last_synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(language_id, path)
);

CREATE TABLE configurations (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT
);

CREATE TABLE user_preferences (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    PRIMARY KEY (user_id, key)
);

CREATE TABLE ignored_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pattern TEXT NOT NULL
);

CREATE TABLE special_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    purpose VARCHAR(100) NOT NULL,
    pattern TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS special_paths;
DROP TABLE IF EXISTS ignored_paths;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS configurations;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
