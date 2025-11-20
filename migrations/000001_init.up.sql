CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE teams (
    name TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    team_name TEXT NOT NULL REFERENCES teams(name),
    username TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE pull_requests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id),
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at TIMESTAMPTZ NULL
);

CREATE TABLE pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    slot SMALLINT NOT NULL CHECK (slot BETWEEN 1 AND 2),
    reviewer_id TEXT NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (pull_request_id, slot),
    UNIQUE (pull_request_id, reviewer_id)
);
