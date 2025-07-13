CREATE TABLE IF NOT EXISTS request_history (
    id SERIAL PRIMARY KEY,
    endpoint TEXT NOT NULL,
    params JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL,
    error_message TEXT,
    response_time_ms INTEGER
);