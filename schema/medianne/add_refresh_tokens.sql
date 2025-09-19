-- Add refresh tokens table for JWT refresh token functionality
-- This migration adds support for refresh tokens to enable secure token renewal

CREATE TABLE mdl_refresh_tokens (
    id serial4 NOT NULL,
    token varchar(255) UNIQUE NOT NULL,
    user_id int8 NOT NULL,
    institution_id int8 NOT NULL,
    device_id varchar(255),
    user_agent text,
    ip_address varchar(45),
    is_revoked bool DEFAULT false NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    CONSTRAINT mdl_refresh_tokens_pkey PRIMARY KEY (id)
);

-- Indexes for performance
CREATE INDEX idx_refresh_tokens_token ON mdl_refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON mdl_refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON mdl_refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON mdl_refresh_tokens(is_revoked);

-- Foreign key constraints
ALTER TABLE mdl_refresh_tokens 
ADD CONSTRAINT fk_refresh_tokens_user_id 
FOREIGN KEY (user_id) REFERENCES mdl_mst_staff(id);

ALTER TABLE mdl_refresh_tokens 
ADD CONSTRAINT fk_refresh_tokens_institution_id 
FOREIGN KEY (institution_id) REFERENCES mdl_mst_institution(id);

-- Add comment to the table
COMMENT ON TABLE mdl_refresh_tokens IS 'Stores refresh tokens for JWT authentication system';
COMMENT ON COLUMN mdl_refresh_tokens.token IS 'Unique refresh token string';
COMMENT ON COLUMN mdl_refresh_tokens.user_id IS 'Reference to staff user who owns the token';
COMMENT ON COLUMN mdl_refresh_tokens.institution_id IS 'Reference to institution for data isolation';
COMMENT ON COLUMN mdl_refresh_tokens.device_id IS 'Device identifier for single device login support';
COMMENT ON COLUMN mdl_refresh_tokens.is_revoked IS 'Flag indicating if token has been revoked';
COMMENT ON COLUMN mdl_refresh_tokens.expires_at IS 'Token expiration timestamp';

-- Add unique constraint to ensure only one active token per user per device
-- This enforces single device login at the database level
CREATE UNIQUE INDEX idx_refresh_tokens_user_device_unique 
ON mdl_refresh_tokens(user_id, device_id) 
WHERE is_revoked = false AND expires_at > NOW();
