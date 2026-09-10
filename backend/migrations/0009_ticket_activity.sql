CREATE TABLE ticket_activity (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    actor_user_id BIGINT,
    action VARCHAR(50) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ticket_activity_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_activity_actor_user_id_fkey
        FOREIGN KEY (actor_user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE SET NULL,
    CONSTRAINT ticket_activity_action_not_blank CHECK (btrim(action) <> ''),
    CONSTRAINT ticket_activity_metadata_object_check
        CHECK (metadata IS NULL OR jsonb_typeof(metadata) = 'object')
);
