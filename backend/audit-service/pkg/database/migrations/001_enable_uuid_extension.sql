-- +migrate Up
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +migrate Down
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
