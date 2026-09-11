-- Read-only checks for the SPEC-01 database foundation.

SELECT version, description, success
FROM _sqlx_migrations
ORDER BY version;

SELECT version, description, success
FROM _sqlx_migrations
WHERE version >= 19
ORDER BY version;

SELECT table_name
FROM information_schema.tables
WHERE table_schema = current_schema()
  AND table_name IN (
    'departments',
    'users',
    'user_preferences',
    'auth_sessions',
    'locations',
    'tickets',
    'ticket_assigned_departments',
    'ticket_assigned_users',
    'ticket_attachments',
    'ticket_activity',
    'ticket_messages',
    'message_attachments',
    'checklist_items',
    'announcements',
    'staff_meals',
    'notifications',
    'audit_logs'
  )
ORDER BY table_name;

SELECT indexname, tablename
FROM pg_indexes
WHERE schemaname = current_schema()
  AND indexname IN (
    'idx_tickets_status_created_at',
    'idx_tickets_requester_created_at',
    'idx_tickets_department_status_created',
    'idx_tickets_location_created_at',
    'idx_ticket_assigned_departments_ticket_department',
    'idx_ticket_assigned_departments_department_ticket',
    'idx_ticket_assigned_users_ticket_user',
    'idx_ticket_assigned_users_user_ticket',
    'idx_ticket_attachments_ticket_created',
    'idx_ticket_messages_ticket_created',
    'idx_notifications_user_unread_created',
    'idx_announcements_published_at',
    'idx_announcements_author_created_at',
    'idx_auth_sessions_active_user_expires_at',
    'idx_staff_meals_valid_range',
    'idx_audit_logs_entity_created'
  )
ORDER BY tablename, indexname;

SELECT indexname
FROM pg_indexes
WHERE schemaname = current_schema()
  AND indexname = 'idx_tickets_assigned_status';

SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'tickets'
  AND column_name IN ('priority', 'due_at')
ORDER BY column_name;

SELECT column_name
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'tickets'
  AND column_name IN ('assigned_to', 'assigned_at')
ORDER BY column_name;

SELECT username, role, email
FROM users
WHERE username = 'hothienty';
