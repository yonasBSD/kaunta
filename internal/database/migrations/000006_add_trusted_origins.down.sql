-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_trusted_origin_timestamp ON trusted_origin;
DROP FUNCTION IF EXISTS update_trusted_origin_timestamp();

-- Drop helper functions
DROP FUNCTION IF EXISTS get_trusted_origins();
DROP FUNCTION IF EXISTS is_trusted_origin(TEXT);

-- Drop index
DROP INDEX IF EXISTS idx_trusted_origin_active;

-- Drop table
DROP TABLE IF EXISTS trusted_origin;
