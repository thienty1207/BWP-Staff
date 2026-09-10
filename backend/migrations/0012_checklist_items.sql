CREATE TABLE checklist_items (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_by BIGINT,
    completed_at TIMESTAMPTZ,
    assigned_to BIGINT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT checklist_items_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES tickets (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT checklist_items_completed_by_fkey
        FOREIGN KEY (completed_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT checklist_items_assigned_to_fkey
        FOREIGN KEY (assigned_to)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT checklist_items_created_by_fkey
        FOREIGN KEY (created_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT checklist_items_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT checklist_items_sort_order_check CHECK (sort_order >= 0),
    CONSTRAINT checklist_items_completion_pair_check
        CHECK ((completed_by IS NULL) = (completed_at IS NULL)),
    CONSTRAINT checklist_items_completed_state_check
        CHECK (NOT is_completed OR (completed_by IS NOT NULL AND completed_at IS NOT NULL))
);
