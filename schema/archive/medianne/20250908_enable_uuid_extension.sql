-- Enable UUID extension for uuid_generate_v4() function
-- This extension is required for tables that use uuid_generate_v4() as default values

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

COMMENT ON EXTENSION "uuid-ossp" IS 'Provides functions to generate UUIDs (universally unique identifiers)';

