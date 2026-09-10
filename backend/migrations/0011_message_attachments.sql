CREATE TABLE message_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    storage_key TEXT NOT NULL,
    public_url TEXT,
    mime_type VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    width INTEGER,
    height INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT message_attachments_message_id_fkey
        FOREIGN KEY (message_id)
        REFERENCES ticket_messages (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT message_attachments_file_size_check CHECK (file_size_bytes >= 0),
    CONSTRAINT message_attachments_width_check CHECK (width IS NULL OR width >= 0),
    CONSTRAINT message_attachments_height_check CHECK (height IS NULL OR height >= 0),
    CONSTRAINT message_attachments_file_name_not_blank CHECK (btrim(file_name) <> ''),
    CONSTRAINT message_attachments_storage_key_not_blank CHECK (btrim(storage_key) <> ''),
    CONSTRAINT message_attachments_mime_type_not_blank CHECK (btrim(mime_type) <> '')
);

