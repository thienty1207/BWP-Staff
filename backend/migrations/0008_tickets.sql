CREATE TABLE tickets (
    id BIGSERIAL PRIMARY KEY,
    requester_id BIGINT NOT NULL,
    department_id BIGINT NOT NULL,
    location_id BIGINT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status ticket_status NOT NULL DEFAULT 'pending',
    accepted_by BIGINT,
    accepted_at TIMESTAMPTZ,
    assigned_to BIGINT,
    assigned_at TIMESTAMPTZ,
    closed_by BIGINT,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tickets_requester_id_fkey
        FOREIGN KEY (requester_id)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_department_id_fkey
        FOREIGN KEY (department_id)
        REFERENCES departments (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_location_id_fkey
        FOREIGN KEY (location_id)
        REFERENCES locations (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_accepted_by_fkey
        FOREIGN KEY (accepted_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_assigned_to_fkey
        FOREIGN KEY (assigned_to)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_closed_by_fkey
        FOREIGN KEY (closed_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT tickets_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT tickets_acceptance_pair_check CHECK ((accepted_by IS NULL) = (accepted_at IS NULL)),
    CONSTRAINT tickets_assignment_pair_check CHECK ((assigned_to IS NULL) = (assigned_at IS NULL)),
    CONSTRAINT tickets_closure_pair_check CHECK ((closed_by IS NULL) = (closed_at IS NULL)),
    CONSTRAINT tickets_pending_acceptance_check
        CHECK (status <> 'pending' OR (accepted_by IS NULL AND accepted_at IS NULL)),
    CONSTRAINT tickets_accepted_acceptance_check
        CHECK (status <> 'accepted' OR (accepted_by IS NOT NULL AND accepted_at IS NOT NULL)),
    CONSTRAINT tickets_closed_acceptance_check
        CHECK (status <> 'closed' OR (accepted_by IS NOT NULL AND accepted_at IS NOT NULL)),
    CONSTRAINT tickets_closed_closure_check
        CHECK (status <> 'closed' OR (closed_by IS NOT NULL AND closed_at IS NOT NULL))
);

