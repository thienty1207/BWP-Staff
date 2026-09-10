CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    employee_code VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    password_hash TEXT NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    department_id BIGINT NOT NULL,
    role user_role NOT NULL DEFAULT 'staff',
    avatar_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_employee_code_key UNIQUE (employee_code),
    CONSTRAINT users_department_id_fkey
        FOREIGN KEY (department_id)
        REFERENCES departments (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT users_username_not_blank CHECK (btrim(username) <> ''),
    CONSTRAINT users_employee_code_not_blank CHECK (btrim(employee_code) <> ''),
    CONSTRAINT users_password_hash_not_blank CHECK (btrim(password_hash) <> ''),
    CONSTRAINT users_full_name_not_blank CHECK (btrim(full_name) <> '')
);

