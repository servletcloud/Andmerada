-- Leave down.sql empty if you want the rollback to have no effect
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
-- DROP TABLE example_table;
COMMIT;
