-- +goose Up
ALTER TABLE projects ADD COLUMN last_repo_check_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE projects ADD COLUMN last_commit_hash VARCHAR(40);

-- +goose Down
ALTER TABLE projects DROP COLUMN last_commit_hash;
ALTER TABLE projects DROP COLUMN last_repo_check_at;
