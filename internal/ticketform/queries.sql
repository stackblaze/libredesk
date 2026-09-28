-- name: get-all-ticket-forms
SELECT
    f.id,
    f.created_at,
    f.updated_at,
    f.name,
    f.description,
    f.inbox_id,
    f.enabled,
    f.fields,
    COALESCE(i.name, '') AS inbox_name
FROM ticket_forms f
LEFT JOIN inboxes i ON i.id = f.inbox_id
ORDER BY f.name;

-- name: get-ticket-form
SELECT
    f.id,
    f.created_at,
    f.updated_at,
    f.name,
    f.description,
    f.inbox_id,
    f.enabled,
    f.fields,
    COALESCE(i.name, '') AS inbox_name
FROM ticket_forms f
LEFT JOIN inboxes i ON i.id = f.inbox_id
WHERE f.id = $1;

-- name: insert-ticket-form
WITH inserted AS (
    INSERT INTO ticket_forms (name, description, inbox_id, enabled, fields)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    i.id,
    i.created_at,
    i.updated_at,
    i.name,
    i.description,
    i.inbox_id,
    i.enabled,
    i.fields,
    COALESCE(ib.name, '') AS inbox_name
FROM inserted i
LEFT JOIN inboxes ib ON ib.id = i.inbox_id;

-- name: update-ticket-form
WITH updated AS (
    UPDATE ticket_forms
    SET
        name = $2,
        description = $3,
        inbox_id = $4,
        enabled = $5,
        fields = $6,
        updated_at = NOW()
    WHERE id = $1
    RETURNING *
)
SELECT
    u.id,
    u.created_at,
    u.updated_at,
    u.name,
    u.description,
    u.inbox_id,
    u.enabled,
    u.fields,
    COALESCE(i.name, '') AS inbox_name
FROM updated u
LEFT JOIN inboxes i ON i.id = u.inbox_id;

-- name: delete-ticket-form
DELETE FROM ticket_forms WHERE id = $1;
