INSERT INTO _table_name_ (
    id,
    name,
    applied_at,
    sql_up,
    sql_down,
    sql_up_sha256,
    sql_down_sha256,
    duration_ms,
    rollback_blocked,
    meta
) VALUES (
    @id,
    @name,
    @applied_at,
    @sql_up,
    @sql_down,
    @sql_up_sha256,
    @sql_down_sha256,
    @duration_ms,
    @rollback_blocked,
    @meta
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    applied_at = EXCLUDED.applied_at,
    sql_up = EXCLUDED.sql_up,
    sql_down = EXCLUDED.sql_down,
    sql_up_sha256 = EXCLUDED.sql_up_sha256,
    sql_down_sha256 = EXCLUDED.sql_down_sha256,
    duration_ms = EXCLUDED.duration_ms,
    rollback_blocked = EXCLUDED.rollback_blocked,
    meta = EXCLUDED.meta;
