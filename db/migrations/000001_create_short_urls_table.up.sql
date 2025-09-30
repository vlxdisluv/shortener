CREATE TABLE short_urls (
    hash TEXT PRIMARY KEY,
    original TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    views BIGINT NOT NULL DEFAULT 0
);