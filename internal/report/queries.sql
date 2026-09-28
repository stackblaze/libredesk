-- name: get-overview-counts
WITH convs AS (
    SELECT
        COUNT(*) AS open,
        COUNT(*) FILTER (
            WHERE
                c.last_message_sender = 'contact'
        ) AS awaiting_response,
        COUNT(*) FILTER (
            WHERE
                c.assigned_user_id IS NULL
        ) AS unassigned,
        COUNT(*) FILTER (
            WHERE
                c.first_reply_at IS NULL
        ) AS pending
    FROM
        conversations c
        INNER JOIN conversation_statuses s ON c.status_id = s.id
    WHERE
        s.category != 'resolved'
),
agents AS (
    SELECT
        COUNT(*) FILTER (
            WHERE
                availability_status = 'online'
        ) AS agents_online,
        COUNT(*) FILTER (
            WHERE
                availability_status = 'away_manual'
        ) AS agents_away,
        COUNT(*) FILTER (
            WHERE
                availability_status = 'away_and_reassigning'
        ) AS agents_reassigning,
        COUNT(*) FILTER (
            WHERE
                availability_status IN ('offline', 'away')
        ) AS agents_offline
    FROM
        users
    WHERE
        type = 'agent'
        AND deleted_at IS NULL
)
SELECT
    json_build_object(
        'open', open,
        'awaiting_response', awaiting_response,
        'unassigned', unassigned,
        'pending', pending,
        'agents_online', agents_online,
        'agents_away', agents_away,
        'agents_reassigning', agents_reassigning,
        'agents_offline', agents_offline
    )
FROM
    convs,
    agents;

-- name: get-overview-sla-counts
-- Count only each conversation's latest applied SLA; superseded rows are kept as history and would double-count.
WITH latest_applied AS (
    SELECT DISTINCT ON (conversation_id)
        created_at, first_response_met_at, first_response_breached_at,
        resolution_met_at, resolution_breached_at
    FROM applied_slas
    WHERE created_at >= CASE
        WHEN %d = 0 THEN CURRENT_DATE
        ELSE NOW() - INTERVAL '%d days'
    END
    ORDER BY conversation_id, created_at DESC, id DESC
),
first_and_resolution AS (
    SELECT
        COUNT(*) FILTER (
            WHERE
                first_response_met_at IS NOT NULL
        ) AS first_response_met_count,
        COUNT(*) FILTER (
            WHERE
                first_response_breached_at IS NOT NULL
        ) AS first_response_breached_count,
        COUNT(*) FILTER (
            WHERE
                resolution_met_at IS NOT NULL
        ) AS resolution_met_count,
        COUNT(*) FILTER (
            WHERE
                resolution_breached_at IS NOT NULL
        ) AS resolution_breached_count,
        COALESCE(
            AVG(
                EXTRACT(
                    EPOCH
                    FROM
                        (first_response_met_at - created_at)
                )
            ) FILTER (
                WHERE
                    first_response_met_at IS NOT NULL
            ),
            0
        ) AS avg_first_response_time_sec,
        COALESCE(
            AVG(
                EXTRACT(
                    EPOCH
                    FROM
                        (resolution_met_at - created_at)
                )
            ) FILTER (
                WHERE
                    resolution_met_at IS NOT NULL
            ),
            0
        ) AS avg_resolution_time_sec
    FROM
        latest_applied
),
next_response AS (
    -- A reply after the deadline carries both met_at and breached_at, so counting the
    -- timestamps puts one event in both buckets. status holds a single terminal verdict.
    SELECT
        COUNT(*) FILTER (
            WHERE
                status = 'met'
        ) AS next_response_met_count,
        COUNT(*) FILTER (
            WHERE
                status = 'breached'
        ) AS next_response_breached_count,
        COALESCE(
            AVG(
                EXTRACT(
                    EPOCH
                    FROM
                        (met_at - created_at)
                )
            ) FILTER (
                WHERE
                    status = 'met'
            ),
            0
        ) AS avg_next_response_time_sec
    FROM
        sla_events
    WHERE
        created_at >= CASE
            WHEN %d = 0 THEN CURRENT_DATE
            ELSE NOW() - INTERVAL '%d days'
        END
        AND type = 'next_response'
)
SELECT
    fas.first_response_met_count,
    fas.first_response_breached_count,
    fas.avg_first_response_time_sec,
    nr.next_response_met_count,
    nr.next_response_breached_count,
    nr.avg_next_response_time_sec,
    fas.resolution_met_count,
    fas.resolution_breached_count,
    fas.avg_resolution_time_sec,
    CASE
        WHEN (fas.first_response_met_count + fas.first_response_breached_count) > 0
        THEN ROUND((fas.first_response_met_count::numeric / (fas.first_response_met_count + fas.first_response_breached_count)::numeric) * 100, 1)
        ELSE 0
    END AS first_response_compliance_percent,
    CASE
        WHEN (nr.next_response_met_count + nr.next_response_breached_count) > 0
        THEN ROUND((nr.next_response_met_count::numeric / (nr.next_response_met_count + nr.next_response_breached_count)::numeric) * 100, 1)
        ELSE 0
    END AS next_response_compliance_percent,
    CASE
        WHEN (fas.resolution_met_count + fas.resolution_breached_count) > 0
        THEN ROUND((fas.resolution_met_count::numeric / (fas.resolution_met_count + fas.resolution_breached_count)::numeric) * 100, 1)
        ELSE 0
    END AS resolution_compliance_percent
FROM
    first_and_resolution fas,
    next_response nr;

-- name: get-overview-charts
WITH new_conversations AS (
    SELECT
        json_agg(row_to_json(agg)) AS data
    FROM
        (
            SELECT
                TO_CHAR(created_at :: date, 'YYYY-MM-DD') AS date,
                COUNT(*) AS count
            FROM
                conversations c
            WHERE
                c.created_at >= CASE
                    WHEN %d = 0 THEN CURRENT_DATE
                    ELSE NOW() - INTERVAL '%d days'
                END
            GROUP BY
                date
            ORDER BY
                date
        ) agg
),
resolved_conversations AS (
    SELECT
        json_agg(row_to_json(agg)) AS data
    FROM
        (
            SELECT
                TO_CHAR(resolved_at :: date, 'YYYY-MM-DD') AS date,
                COUNT(*) AS count
            FROM
                conversations c
            WHERE
                c.resolved_at >= CASE
                    WHEN %d = 0 THEN CURRENT_DATE
                    ELSE NOW() - INTERVAL '%d days'
                END
            GROUP BY
                date
            ORDER BY
                date
        ) agg
)
SELECT
    json_build_object(
        'new_conversations',
        (
            SELECT
                data
            FROM
                new_conversations
        ),
        'resolved_conversations',
        (
            SELECT
                data
            FROM
                resolved_conversations
        )
    ) AS result;

-- name: get-overview-csat
SELECT
    json_build_object(
        'average_rating',
        COALESCE(AVG(rating) FILTER (WHERE rating > 0), 0),
        'total_responses',
        COUNT(*) FILTER (WHERE rating > 0),
        'total_sent',
        COUNT(*),
        'response_rate',
        CASE
            WHEN COUNT(*) > 0
            THEN ROUND((COUNT(*) FILTER (WHERE rating > 0)::numeric / COUNT(*)::numeric) * 100, 1)
            ELSE 0
        END
    ) AS result
FROM
    csat_responses
WHERE
    created_at >= CASE
        WHEN %d = 0 THEN CURRENT_DATE
        ELSE NOW() - INTERVAL '%d days'
    END;

-- name: get-overview-message-volume
WITH per_conversation AS (
    SELECT
        conversation_id,
        COUNT(*) AS total,
        COUNT(*) FILTER (WHERE type = 'incoming') AS incoming,
        COUNT(*) FILTER (WHERE type = 'outgoing') AS outgoing
    FROM
        conversation_messages
    WHERE
        type IN ('incoming', 'outgoing')
        AND private = false
        AND (type = 'incoming' OR status = 'sent')
        AND created_at >= CASE
            WHEN %d = 0 THEN CURRENT_DATE
            ELSE NOW() - INTERVAL '%d days'
        END
    GROUP BY
        conversation_id
),
stats AS (
    SELECT
        COALESCE(SUM(total), 0) AS total,
        COALESCE(SUM(incoming), 0) AS incoming,
        COALESCE(SUM(outgoing), 0) AS outgoing,
        COUNT(*) AS convos
    FROM
        per_conversation
)
SELECT
    json_build_object(
        'total_messages', total,
        'incoming_messages', incoming,
        'outgoing_messages', outgoing,
        'messages_per_conversation',
        CASE
            WHEN convos > 0 THEN ROUND(total::numeric / convos::numeric, 1)
            ELSE 0
        END
    ) AS result
FROM
    stats;

-- name: get-overview-tag-distribution
WITH tag_counts AS (
    SELECT
        t.id AS tag_id,
        t.name AS tag_name,
        COUNT(c.id) AS count
    FROM
        tags t
        LEFT JOIN conversation_tags ct ON t.id = ct.tag_id
        LEFT JOIN conversations c ON ct.conversation_id = c.id
            AND c.created_at >= CASE
                WHEN %d = 0 THEN CURRENT_DATE
                ELSE NOW() - INTERVAL '%d days'
            END
    GROUP BY
        t.id, t.name
    ORDER BY
        count DESC, t.id
    LIMIT 10
),
tagging AS (
    SELECT
        COUNT(*) AS total,
        COUNT(*) FILTER (
            WHERE EXISTS (
                SELECT 1 FROM conversation_tags ct
                WHERE ct.conversation_id = c.id
            )
        ) AS tagged
    FROM
        conversations c
    WHERE
        c.created_at >= CASE
            WHEN %d = 0 THEN CURRENT_DATE
            ELSE NOW() - INTERVAL '%d days'
        END
)
SELECT
    json_build_object(
        'top_tags',
        COALESCE((SELECT json_agg(row_to_json(tc)) FROM tag_counts tc), '[]'::json),
        'tagged_conversations', tagged,
        'untagged_conversations', total - tagged,
        'tagged_percentage',
        CASE
            WHEN total > 0
            THEN ROUND((tagged::numeric / total::numeric) * 100, 1)
            ELSE 0
        END
    ) AS result
FROM
    tagging;

-- name: get-agent-reports
SELECT
    u.id,
    u.first_name,
    u.last_name,
    COUNT(DISTINCT c.id) AS tickets_assigned,
    COUNT(DISTINCT c.id) FILTER (
        WHERE s.category = 'resolved' AND c.resolved_at >= CASE
            WHEN %d = 0 THEN CURRENT_DATE
            ELSE NOW() - INTERVAL '%d days'
        END
    ) AS tickets_resolved,
    COALESCE((
        SELECT COUNT(*)
        FROM conversation_messages m
        WHERE m.sender_id = u.id AND m.type = 'outgoing' AND m.private = false
          AND m.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
    ), 0) AS replies,
    COALESCE(AVG(EXTRACT(EPOCH FROM (c.first_reply_at - c.created_at))) FILTER (WHERE c.first_reply_at IS NOT NULL), 0) AS avg_first_reply_seconds,
    COALESCE(AVG(cs.rating), 0) AS csat_avg
FROM conversations c
JOIN users u ON u.id = c.assigned_user_id
LEFT JOIN conversation_statuses s ON s.id = c.status_id
LEFT JOIN LATERAL (
    SELECT rating FROM csat_responses
    WHERE conversation_id = c.id
    ORDER BY response_timestamp DESC NULLS LAST, created_at DESC
    LIMIT 1
) cs ON true
WHERE c.assigned_user_id IS NOT NULL
  AND c.created_at >= CASE
      WHEN %d = 0 THEN CURRENT_DATE
      ELSE NOW() - INTERVAL '%d days'
  END
GROUP BY u.id, u.first_name, u.last_name
ORDER BY tickets_assigned DESC;

-- name: get-team-reports
SELECT
    t.id,
    t.name,
    COUNT(DISTINCT c.id) AS tickets_assigned,
    COUNT(DISTINCT c.id) FILTER (
        WHERE s.category = 'resolved' AND c.resolved_at >= CASE
            WHEN %d = 0 THEN CURRENT_DATE
            ELSE NOW() - INTERVAL '%d days'
        END
    ) AS tickets_resolved,
    COALESCE((
        SELECT COUNT(*)
        FROM conversation_messages m
        JOIN conversations c2 ON c2.id = m.conversation_id
        WHERE c2.assigned_team_id = t.id AND m.type = 'outgoing' AND m.private = false
          AND m.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
    ), 0) AS replies,
    COALESCE(AVG(EXTRACT(EPOCH FROM (c.first_reply_at - c.created_at))) FILTER (WHERE c.first_reply_at IS NOT NULL), 0) AS avg_first_reply_seconds,
    COALESCE(AVG(cs.rating), 0) AS csat_avg
FROM conversations c
JOIN teams t ON t.id = c.assigned_team_id
LEFT JOIN conversation_statuses s ON s.id = c.status_id
LEFT JOIN LATERAL (
    SELECT rating FROM csat_responses
    WHERE conversation_id = c.id
    ORDER BY response_timestamp DESC NULLS LAST, created_at DESC
    LIMIT 1
) cs ON true
WHERE c.assigned_team_id IS NOT NULL
  AND c.created_at >= CASE
      WHEN %d = 0 THEN CURRENT_DATE
      ELSE NOW() - INTERVAL '%d days'
  END
GROUP BY t.id, t.name
ORDER BY tickets_assigned DESC;

-- name: get-ticket-reports
WITH created_rows AS (
    SELECT
        i.channel::text AS channel,
        c.created_at::date AS day,
        COALESCE(s.name, '') AS status,
        COALESCE(p.name, '') AS priority
    FROM conversations c
    JOIN inboxes i ON i.id = c.inbox_id
    LEFT JOIN conversation_statuses s ON s.id = c.status_id
    LEFT JOIN conversation_priorities p ON p.id = c.priority_id
    WHERE c.created_at >= CASE
        WHEN %d = 0 THEN CURRENT_DATE
        ELSE NOW() - INTERVAL '%d days'
    END
    %s
),
resolved_rows AS (
    SELECT i.channel::text AS channel, c.resolved_at::date AS day
    FROM conversations c
    JOIN inboxes i ON i.id = c.inbox_id
    WHERE c.resolved_at IS NOT NULL
      AND c.resolved_at >= CASE
        WHEN %d = 0 THEN CURRENT_DATE
        ELSE NOW() - INTERVAL '%d days'
    END
    %s
),
channels AS (
    SELECT channel FROM created_rows
    UNION
    SELECT channel FROM resolved_rows
)
SELECT json_build_object(
    'created', (SELECT COUNT(*) FROM created_rows),
    'resolved', (SELECT COUNT(*) FROM resolved_rows),
    'by_channel', COALESCE((
        SELECT json_agg(row_to_json(x) ORDER BY x.channel)
        FROM (
            SELECT
                ch.channel,
                (SELECT COUNT(*) FROM created_rows cr WHERE cr.channel = ch.channel) AS created,
                (SELECT COUNT(*) FROM resolved_rows rr WHERE rr.channel = ch.channel) AS resolved
            FROM channels ch
        ) x
    ), '[]'::json),
    'new_conversations', COALESCE((
        SELECT json_agg(row_to_json(agg) ORDER BY agg.date)
        FROM (
            SELECT TO_CHAR(day, 'YYYY-MM-DD') AS date, COUNT(*) AS count
            FROM created_rows
            GROUP BY day
        ) agg
    ), '[]'::json),
    'resolved_conversations', COALESCE((
        SELECT json_agg(row_to_json(agg) ORDER BY agg.date)
            FROM (
            SELECT TO_CHAR(day, 'YYYY-MM-DD') AS date, COUNT(*) AS count
            FROM resolved_rows
            GROUP BY day
        ) agg
    ), '[]'::json),
    'by_status', COALESCE((
        SELECT json_agg(row_to_json(x) ORDER BY x.status)
        FROM (
            SELECT status, COUNT(*) AS created
            FROM created_rows
            WHERE status <> ''
            GROUP BY status
        ) x
    ), '[]'::json),
    'by_priority', COALESCE((
        SELECT json_agg(row_to_json(x) ORDER BY x.priority)
        FROM (
            SELECT priority, COUNT(*) AS created
            FROM created_rows
            WHERE priority <> ''
            GROUP BY priority
        ) x
    ), '[]'::json)
) AS result;

-- name: get-efficiency-reports
SELECT json_build_object(
    'median_first_reply_seconds', (
        SELECT percentile_cont(0.5) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (c.first_reply_at - c.created_at))
        )
        FROM conversations c
        WHERE c.first_reply_at IS NOT NULL
          AND c.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    ),
    'first_reply_count', (
        SELECT COUNT(*)
        FROM conversations c
        WHERE c.first_reply_at IS NOT NULL
          AND c.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    ),
    'median_resolution_seconds', (
        SELECT percentile_cont(0.5) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (c.resolved_at - c.created_at))
        )
        FROM conversations c
        WHERE c.resolved_at IS NOT NULL
          AND c.resolved_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    ),
    'resolution_count', (
        SELECT COUNT(*)
        FROM conversations c
        WHERE c.resolved_at IS NOT NULL
          AND c.resolved_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    ),
    'sla_met', (
        SELECT COUNT(*)
        FROM applied_slas a
        JOIN conversations c ON c.id = a.conversation_id
        WHERE a.first_response_met_at IS NOT NULL
          AND a.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    ),
    'sla_breached', (
        SELECT COUNT(*)
        FROM applied_slas a
        JOIN conversations c ON c.id = a.conversation_id
        WHERE a.first_response_breached_at IS NOT NULL
          AND a.created_at >= CASE
              WHEN %d = 0 THEN CURRENT_DATE
              ELSE NOW() - INTERVAL '%d days'
          END
          %s
    )
) AS result;

-- name: get-backlog-reports
WITH days AS (
    SELECT generate_series(
        (CURRENT_DATE - INTERVAL '%d days')::date,
        CURRENT_DATE,
        INTERVAL '1 day'
    )::date AS day
),
open_now AS (
    SELECT i.channel::text AS channel, COUNT(*) AS open
    FROM conversations c
    JOIN conversation_statuses s ON s.id = c.status_id
    JOIN inboxes i ON i.id = c.inbox_id
    WHERE s.category <> 'resolved'
    %s
    GROUP BY i.channel
),
series AS (
    SELECT
        TO_CHAR(d.day, 'YYYY-MM-DD') AS date,
        COUNT(c.id) AS count
    FROM days d
    LEFT JOIN conversations c
        ON c.created_at < (d.day + INTERVAL '1 day')
       AND (c.resolved_at IS NULL OR c.resolved_at >= (d.day + INTERVAL '1 day'))
       %s
    GROUP BY d.day
)
SELECT json_build_object(
    'open', COALESCE((SELECT SUM(open) FROM open_now), 0),
    'by_channel', COALESCE((
        SELECT json_agg(row_to_json(o) ORDER BY o.channel) FROM open_now o
    ), '[]'::json),
    'series', COALESCE((
        SELECT json_agg(row_to_json(s) ORDER BY s.date) FROM series s
    ), '[]'::json)
) AS result;