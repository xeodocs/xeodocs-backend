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

-- +goose Down
DROP TABLE roles;
