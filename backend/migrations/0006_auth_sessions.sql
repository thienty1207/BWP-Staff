CREATE TABLE auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_token_hash TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT auth_sessions_token_hash_key UNIQUE (session_token_hash),
    CONSTRAINT auth_sessions_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT auth_sessions_token_hash_not_blank CHECK (btrim(session_token_hash) <> ''),
    CONSTRAINT auth_sessions_expiry_after_creation CHECK (expires_at > created_at)
);
