ALTER TABLE tickets
    ADD COLUMN priority BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN due_at TIMESTAMPTZ;

CREATE TABLE ticket_assigned_departments (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    department_id BIGINT NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ticket_assigned_departments_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_assigned_departments_department_id_fkey
        FOREIGN KEY (department_id)
        REFERENCES departments (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_ticket_assigned_departments_ticket_department
    ON ticket_assigned_departments (ticket_id, department_id);

CREATE INDEX idx_ticket_assigned_departments_department_ticket
    ON ticket_assigned_departments (department_id, ticket_id);

CREATE TABLE ticket_assigned_users (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ticket_assigned_users_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_assigned_users_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_ticket_assigned_users_ticket_user
    ON ticket_assigned_users (ticket_id, user_id);

CREATE INDEX idx_ticket_assigned_users_user_ticket
    ON ticket_assigned_users (user_id, ticket_id);

CREATE TABLE ticket_attachments (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    uploaded_by BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    storage_key TEXT NOT NULL,
    public_url TEXT,
    mime_type VARCHAR(150) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    width INTEGER,
    height INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ticket_attachments_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_attachments_uploaded_by_fkey
        FOREIGN KEY (uploaded_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT ticket_attachments_file_size_check CHECK (file_size_bytes >= 0),
    CONSTRAINT ticket_attachments_width_check CHECK (width IS NULL OR width >= 0),
    CONSTRAINT ticket_attachments_height_check CHECK (height IS NULL OR height >= 0),
    CONSTRAINT ticket_attachments_file_name_not_blank CHECK (btrim(file_name) <> ''),
    CONSTRAINT ticket_attachments_storage_key_not_blank CHECK (btrim(storage_key) <> ''),
    CONSTRAINT ticket_attachments_mime_type_not_blank CHECK (btrim(mime_type) <> '')
);

CREATE INDEX idx_ticket_attachments_ticket_created
    ON ticket_attachments (ticket_id, created_at ASC, id ASC);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM tickets
        WHERE (assigned_to IS NULL) <> (assigned_at IS NULL)
    ) THEN
        RAISE EXCEPTION 'legacy ticket assignment has mismatched assigned_to and assigned_at';
    END IF;
END;
$$;

INSERT INTO ticket_assigned_users (ticket_id, user_id, assigned_at)
SELECT id, assigned_to, assigned_at
FROM tickets
WHERE assigned_to IS NOT NULL;

DROP INDEX IF EXISTS idx_tickets_assigned_status;

ALTER TABLE tickets
    DROP CONSTRAINT tickets_assignment_pair_check,
    DROP CONSTRAINT tickets_assigned_to_fkey,
    DROP COLUMN assigned_to,
    DROP COLUMN assigned_at;
