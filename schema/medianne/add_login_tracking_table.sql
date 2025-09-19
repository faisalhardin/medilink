-- Add login tracking table for comprehensive login monitoring
-- This migration adds a table to track all login attempts and device sessions

CREATE TABLE mdl_mst_login (
    id serial4 NOT NULL,
    user_id int8 NOT NULL,
    institution_id int8 NOT NULL,
    device_id varchar(255) NOT NULL,
    user_agent text,
    ip_address varchar(45) NOT NULL,
    login_type varchar(50) NOT NULL, -- 'google_oauth', 'refresh_token', 'manual'
    session_id varchar(255), -- Reference to refresh token or session
    status varchar(20) NOT NULL, -- 'success', 'failed', 'revoked'
    failure_reason varchar(255), -- Reason for failed login
    login_at timestamptz NOT NULL,
    logout_at timestamptz NULL,
    expires_at timestamptz NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT mdl_mst_login_pkey PRIMARY KEY (id)
);

-- Indexes for performance and queries
CREATE INDEX idx_login_user_id ON mdl_mst_login(user_id);
CREATE INDEX idx_login_device_id ON mdl_mst_login(device_id);
CREATE INDEX idx_login_institution_id ON mdl_mst_login(institution_id);
CREATE INDEX idx_login_login_at ON mdl_mst_login(login_at);
CREATE INDEX idx_login_status ON mdl_mst_login(status);
CREATE INDEX idx_login_session_id ON mdl_mst_login(session_id);

-- Composite indexes for common queries
CREATE INDEX idx_login_user_device ON mdl_mst_login(user_id, device_id);
CREATE INDEX idx_login_user_status ON mdl_mst_login(user_id, status);
CREATE INDEX idx_login_device_status ON mdl_mst_login(device_id, status);

-- Add comments
COMMENT ON TABLE mdl_mst_login IS 'Tracks all login attempts and device sessions for auditing and monitoring';
COMMENT ON COLUMN mdl_mst_login.user_id IS 'Reference to staff user who logged in';
COMMENT ON COLUMN mdl_mst_login.institution_id IS 'Reference to institution for data isolation';
COMMENT ON COLUMN mdl_mst_login.device_id IS 'Device identifier for tracking sessions';
COMMENT ON COLUMN mdl_mst_login.user_agent IS 'Browser/client information';
COMMENT ON COLUMN mdl_mst_login.ip_address IS 'IP address of the login attempt';
COMMENT ON COLUMN mdl_mst_login.login_type IS 'Type of login (google_oauth, refresh_token, manual)';
COMMENT ON COLUMN mdl_mst_login.session_id IS 'Reference to refresh token or session identifier';
COMMENT ON COLUMN mdl_mst_login.status IS 'Status of the login attempt (success, failed, revoked)';
COMMENT ON COLUMN mdl_mst_login.failure_reason IS 'Reason for failed login attempts';
COMMENT ON COLUMN mdl_mst_login.login_at IS 'Timestamp when login occurred';
COMMENT ON COLUMN mdl_mst_login.logout_at IS 'Timestamp when logout occurred (NULL if still active)';
COMMENT ON COLUMN mdl_mst_login.expires_at IS 'Timestamp when session expires';
