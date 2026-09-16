-- Create user sessions table for revamped authentication system
CREATE TABLE public.mdl_mst_user_sessions (
    id serial4 NOT NULL,
    session_key varchar NOT NULL UNIQUE,
    user_id int8 NOT NULL,
    access_token_hash varchar NOT NULL,
    refresh_token_hash varchar NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'active',
    expires_at timestamptz NOT NULL,
    refresh_expires_at timestamptz NOT NULL,
    last_accessed_at timestamptz NOT NULL,
    ip_address varchar(45),
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT mdl_mst_user_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_mst_user_sessions_status_check CHECK (status IN ('active', 'revoked', 'expired'))
);

-- Create indexes for performance
CREATE INDEX idx_user_sessions_session_key ON mdl_mst_user_sessions(session_key);
CREATE INDEX idx_user_sessions_user_id ON mdl_mst_user_sessions(user_id);
CREATE INDEX idx_user_sessions_status ON mdl_mst_user_sessions(status);
CREATE INDEX idx_user_sessions_expires_at ON mdl_mst_user_sessions(expires_at);
CREATE INDEX idx_user_sessions_refresh_expires_at ON mdl_mst_user_sessions(refresh_expires_at);
CREATE INDEX idx_user_sessions_last_accessed_at ON mdl_mst_user_sessions(last_accessed_at);

-- Add comments to table and columns
COMMENT ON TABLE mdl_mst_user_sessions IS 'User session management for authentication system';
COMMENT ON COLUMN mdl_mst_user_sessions.session_key IS 'Unique session identifier';
COMMENT ON COLUMN mdl_mst_user_sessions.user_id IS 'Reference to staff user';
COMMENT ON COLUMN mdl_mst_user_sessions.access_token_hash IS 'Hashed access token for security';
COMMENT ON COLUMN mdl_mst_user_sessions.refresh_token_hash IS 'Hashed refresh token for security';
COMMENT ON COLUMN mdl_mst_user_sessions.status IS 'Session status: active, revoked, expired';
COMMENT ON COLUMN mdl_mst_user_sessions.expires_at IS 'Access token expiration time';
COMMENT ON COLUMN mdl_mst_user_sessions.refresh_expires_at IS 'Refresh token expiration time';
COMMENT ON COLUMN mdl_mst_user_sessions.last_accessed_at IS 'Last time session was accessed';
COMMENT ON COLUMN mdl_mst_user_sessions.ip_address IS 'Client IP address for security tracking';
COMMENT ON COLUMN mdl_mst_user_sessions.user_agent IS 'Client user agent for security tracking';
