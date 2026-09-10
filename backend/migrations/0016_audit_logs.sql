CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT,
    action audit_action NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id BIGINT,
    metadata JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT audit_logs_actor_user_id_fkey
        FOREIGN KEY (actor_user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE SET NULL,
    CONSTRAINT audit_logs_entity_type_not_blank CHECK (btrim(entity_type) <> ''),
    CONSTRAINT audit_logs_metadata_object_check
        CHECK (metadata IS NULL OR jsonb_typeof(metadata) = 'object')
);
