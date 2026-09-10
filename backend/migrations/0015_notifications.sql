CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    ticket_id BIGINT,
    announcement_id BIGINT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notifications_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT notifications_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE SET NULL,
    CONSTRAINT notifications_announcement_id_fkey
        FOREIGN KEY (announcement_id)
        REFERENCES announcements (id)
        ON UPDATE RESTRICT
        ON DELETE SET NULL,
    CONSTRAINT notifications_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT notifications_read_state_check
        CHECK ((is_read AND read_at IS NOT NULL) OR (NOT is_read AND read_at IS NULL))
);
