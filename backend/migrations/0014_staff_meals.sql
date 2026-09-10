CREATE TABLE staff_meals (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255),
    image_storage_key TEXT NOT NULL,
    image_url TEXT,
    valid_from DATE,
    valid_to DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    uploaded_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT staff_meals_uploaded_by_fkey
        FOREIGN KEY (uploaded_by)
        REFERENCES users (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT staff_meals_image_storage_key_not_blank CHECK (btrim(image_storage_key) <> ''),
    CONSTRAINT staff_meals_valid_range_check
        CHECK (valid_from IS NULL OR valid_to IS NULL OR valid_to >= valid_from)
);

