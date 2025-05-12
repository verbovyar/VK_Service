-- +goose Up
CREATE TABLE IF NOT EXISTS messages (
    id         SERIAL      PRIMARY KEY,
    subject    TEXT        NOT NULL,
    data       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS messages;
