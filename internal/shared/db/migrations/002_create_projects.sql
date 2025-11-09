-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    doc_url TEXT NOT NULL,
    repo_url TEXT NOT NULL,
    languages JSONB,
    build_command TEXT,
    export_command TEXT,
    preview_command TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE projects;
