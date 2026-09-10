CREATE TYPE user_role AS ENUM ('admin', 'staff');

CREATE TYPE ticket_status AS ENUM ('pending', 'accepted', 'closed');

CREATE TYPE notification_type AS ENUM (
    'ticket_created',
    'ticket_accepted',
    'ticket_assigned',
    'ticket_closed',
    'new_message',
    'announcement',
    'system'
);

CREATE TYPE audit_action AS ENUM (
    'create',
    'update',
    'delete',
    'login',
    'logout',
    'accept',
    'assign',
    'close',
    'publish',
    'unpublish'
);
