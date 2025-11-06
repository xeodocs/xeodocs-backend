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

-- Insert default admin user (password: 'tempadmin123' hashed with bcrypt)
-- Note: Change this password after first login
INSERT INTO users (username, password, role, created_at) VALUES (
    'admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash for 'tempadmin123'
    'admin',
    CURRENT_TIMESTAMP
) ON CONFLICT (username) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE username = 'admin';
DROP TABLE roles;
