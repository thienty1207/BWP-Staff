CREATE UNIQUE INDEX users_email_unique_idx
    ON users (lower(email))
    WHERE email IS NOT NULL;

CREATE INDEX idx_users_department_id
    ON users (department_id);

CREATE INDEX idx_users_is_active
    ON users (id)
    WHERE is_active = TRUE;

CREATE INDEX idx_auth_sessions_user_expires_at
    ON auth_sessions (user_id, expires_at DESC, id DESC);

CREATE INDEX idx_auth_sessions_expires_at
    ON auth_sessions (expires_at);

CREATE INDEX idx_auth_sessions_active_user_expires_at
    ON auth_sessions (user_id, expires_at DESC, id DESC)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_locations_name
    ON locations (name);

CREATE INDEX idx_tickets_status_created_at
    ON tickets (status, created_at DESC, id DESC);

CREATE INDEX idx_tickets_requester_created_at
    ON tickets (requester_id, created_at DESC, id DESC);

CREATE INDEX idx_tickets_assigned_status
    ON tickets (assigned_to, status);

CREATE INDEX idx_tickets_department_status_created
    ON tickets (department_id, status, created_at DESC, id DESC);

CREATE INDEX idx_tickets_location_created_at
    ON tickets (location_id, created_at DESC, id DESC);

CREATE INDEX idx_ticket_activity_ticket_created
    ON ticket_activity (ticket_id, created_at DESC, id DESC);

CREATE INDEX idx_ticket_activity_actor_created
    ON ticket_activity (actor_user_id, created_at DESC, id DESC);

CREATE INDEX idx_ticket_messages_ticket_created
    ON ticket_messages (ticket_id, created_at ASC, id ASC);

CREATE INDEX idx_ticket_messages_sender_created
    ON ticket_messages (sender_id, created_at DESC, id DESC);

CREATE INDEX idx_message_attachments_message_id
    ON message_attachments (message_id, id ASC);

CREATE INDEX idx_checklist_items_ticket_order
    ON checklist_items (ticket_id, sort_order ASC, id ASC);

CREATE INDEX idx_checklist_items_ticket_completed
    ON checklist_items (ticket_id, is_completed, id ASC);

CREATE INDEX idx_checklist_items_assigned_completed
    ON checklist_items (assigned_to, is_completed, id ASC);

CREATE INDEX idx_announcements_created_at
    ON announcements (created_at DESC, id DESC);

CREATE INDEX idx_announcements_published_at
    ON announcements (published_at DESC, id DESC)
    WHERE is_published = TRUE;

CREATE INDEX idx_staff_meals_active_created
    ON staff_meals (is_active, created_at DESC, id DESC);

CREATE INDEX idx_staff_meals_valid_range
    ON staff_meals USING GIST (
        daterange(
            COALESCE(valid_from, '-infinity'::date),
            COALESCE(valid_to, 'infinity'::date),
            '[]'
        )
    );

CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC, id DESC);

CREATE INDEX idx_notifications_user_unread_created
    ON notifications (user_id, created_at DESC, id DESC)
    WHERE is_read = FALSE;

CREATE INDEX idx_audit_logs_actor_created
    ON audit_logs (actor_user_id, created_at DESC, id DESC);

CREATE INDEX idx_audit_logs_entity_created
    ON audit_logs (entity_type, entity_id, created_at DESC, id DESC);

CREATE INDEX idx_audit_logs_action_created
    ON audit_logs (action, created_at DESC, id DESC);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$function$;

CREATE TRIGGER departments_set_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER user_preferences_set_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER locations_set_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER tickets_set_updated_at
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER checklist_items_set_updated_at
    BEFORE UPDATE ON checklist_items
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER announcements_set_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER staff_meals_set_updated_at
    BEFORE UPDATE ON staff_meals
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
