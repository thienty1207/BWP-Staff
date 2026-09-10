CREATE TABLE ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ticket_messages_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_messages_sender_id_fkey
        FOREIGN KEY (sender_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);
