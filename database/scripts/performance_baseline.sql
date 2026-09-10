-- Bounded local/staging performance checks for the SPEC-01 ticket read paths.
-- Review the plan before running this against a production database.

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, title, status, created_at
FROM tickets
WHERE status = 'pending'::ticket_status
ORDER BY created_at DESC, id DESC
LIMIT 50;

EXPLAIN (ANALYZE, BUFFERS)
SELECT t.id, t.title, t.status, t.created_at
FROM tickets AS t
JOIN users AS u ON u.id = t.requester_id
WHERE u.username = 'hothienty'
ORDER BY t.created_at DESC, t.id DESC
LIMIT 50;
