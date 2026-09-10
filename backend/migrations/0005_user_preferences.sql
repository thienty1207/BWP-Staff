CREATE TABLE user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    theme VARCHAR(20) NOT NULL DEFAULT 'dark',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_preferences_user_key UNIQUE (user_id),
    CONSTRAINT user_preferences_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT user_preferences_theme_check
        CHECK (theme IN ('dark', 'light', 'system')),
    CONSTRAINT user_preferences_language_not_blank CHECK (btrim(language) <> '')
);

