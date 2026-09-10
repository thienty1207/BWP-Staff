CREATE TABLE locations (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT locations_code_key UNIQUE (code),
    CONSTRAINT locations_name_not_blank CHECK (btrim(name) <> '')
);

