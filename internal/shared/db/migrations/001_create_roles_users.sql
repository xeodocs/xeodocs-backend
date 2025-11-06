-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES ('admin', 'Full access') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name, description) VALUES ('editor', 'Can edit content') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name, description) VALUES ('viewer', 'Read-only access') ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Insert default admin user (password: 'tempadmin123' hashed with bcrypt)
-- Note: Change this password after first login
INSERT INTO users (username, password, role_id, created_at) VALUES (
    'admin',
    '$2a$10$EyXgHjDjZdGXkdQzI5atluZiAOkLncOyFMf0ftHaY/8kDvY0iCrpS', -- bcrypt hash for 'tempadmin123'
    1, -- admin role id
    CURRENT_TIMESTAMP
) ON CONFLICT (username) DO NOTHING;

-- +goose Down
DROP TABLE users;
DROP TABLE roles;