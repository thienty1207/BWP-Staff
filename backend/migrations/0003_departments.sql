CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT departments_code_key UNIQUE (code),
    CONSTRAINT departments_name_key UNIQUE (name),
    CONSTRAINT departments_code_not_blank CHECK (btrim(code) <> ''),
    CONSTRAINT departments_name_not_blank CHECK (btrim(name) <> '')
);

