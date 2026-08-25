CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_user_id BIGINT NOT NULL UNIQUE,
    telegram_username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    flag_hash VARCHAR(255) NOT NULL,
    initial_points INTEGER NOT NULL CHECK (initial_points >= 0),
    minimum_points INTEGER NOT NULL CHECK (minimum_points >= 0),
    current_points INTEGER NOT NULL CHECK (current_points >= 0),
    decay INTEGER NOT NULL CHECK (decay >= 0),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tasks_points_order CHECK (minimum_points <= current_points AND current_points <= initial_points),
    CONSTRAINT tasks_dates_order CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_tasks_starts_at ON tasks (starts_at);
CREATE INDEX idx_tasks_is_active ON tasks (is_active);

CREATE TABLE task_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    storage_path TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    telegram_file_id TEXT,
    mime_type VARCHAR(255),
    file_size BIGINT NOT NULL CHECK (file_size >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_files_task_id ON task_files (task_id);

CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    points_awarded INTEGER NOT NULL DEFAULT 0 CHECK (points_awarded >= 0),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_user_id ON submissions (user_id);
CREATE INDEX idx_submissions_task_id ON submissions (task_id);
CREATE INDEX idx_submissions_submitted_at ON submissions (submitted_at);
CREATE INDEX idx_submissions_is_correct ON submissions (is_correct);
CREATE UNIQUE INDEX uq_submissions_correct_solution
    ON submissions (user_id, task_id)
    WHERE is_correct = TRUE;

CREATE TABLE ratings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    total_points INTEGER NOT NULL DEFAULT 0 CHECK (total_points >= 0),
    solved_tasks_count INTEGER NOT NULL DEFAULT 0 CHECK (solved_tasks_count >= 0),
    last_solved_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ratings_total_points ON ratings (total_points DESC);
