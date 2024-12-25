-- Andmerada does not provide automatic transaction management.
-- You must wrap your migration logic in explicit `BEGIN` and `COMMIT` statements if a transaction
-- is required.
--
-- Use transactions for operations that need atomicity to ensure changes are applied or rolled back
-- as a unit (e.g., creating multiple interdependent objects).
BEGIN;

-- Set timeouts to avoid long-running or blocking queries
-- SET LOCAL statement_timeout = '5min';
-- SET LOCAL lock_timeout = '10s';
-- SET LOCAL lock_wait_timeout = '30s';
-- SET LOCAL idle_in_transaction_session_timeout = '60s';
--
-- Improve performance for index creation
-- SET LOCAL work_mem = '128MB';
-- SET LOCAL maintenance_work_mem = '256MB';
--
-- Tag this session for easier traceability
-- SET LOCAL application_name = 'andmerada_migration';
--
-- CREATE TABLE example_table (
--     id BIGSERIAL CONSTRAINT pk_example_table PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMPTZ DEFAULT NOW ()
-- );
--
COMMIT;
