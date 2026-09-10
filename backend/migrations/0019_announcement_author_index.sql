CREATE INDEX IF NOT EXISTS idx_announcements_author_created_at
    ON announcements (author_id, created_at DESC, id DESC);
