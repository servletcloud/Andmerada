CREATE TABLE IF NOT EXISTS "_table_name_" (
    id BIGINT PRIMARY KEY, -- The unique migration ID (e.g., a timestamp like 20241225112129)
    name TEXT NOT NULL CHECK (char_length(name) <= 255),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    sql_up TEXT NOT NULL CHECK (char_length(sql_up) <= 1000000),
    sql_down TEXT CHECK (char_length(sql_down) <= 1000000),
    sql_up_sha256 TEXT NOT NULL,
    sql_down_sha256 TEXT,
    duration_ms BIGINT NOT NULL,
    rollback_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    meta JSONB NOT NULL
);
