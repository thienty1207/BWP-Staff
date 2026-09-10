CREATE TABLE announcements (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT announcements_author_id_fkey
        FOREIGN KEY (author_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT announcements_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT announcements_published_state_check
        CHECK ((is_published AND published_at IS NOT NULL) OR (NOT is_published AND published_at IS NULL))
);

