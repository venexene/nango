-- name: CreateLink :one
INSERT INTO links (short_code, original_url)
VALUES ($1, $2)
RETURNING id, short_code, original_url, created_at;

-- name: GetLinkByShortCode :one
SELECT id, short_code, original_url, created_at
FROM links
WHERE short_code = $1;

-- name: GetLinkByOriginalURL :one
SELECT id, short_code, original_url, created_at
FROM links
WHERE original_url = $1;

-- name: RecordClick :exec
INSERT INTO clicks (link_id, user_agent, ip)
VALUES ($1, $2, $3);

-- name: GetClicksByDay :many
SELECT DATE(clicked_at)::TEXT AS day, COUNT(*)::INT AS count
FROM clicks
WHERE link_id = $1
GROUP BY DATE(clicked_at)
ORDER BY day;

-- name: GetClicksByMonth :many
SELECT DATE_TRUNC('month', clicked_at)::TEXT AS month, COUNT(*)::INT AS count
FROM clicks
WHERE link_id = $1
GROUP BY DATE_TRUNC('month', clicked_at)
ORDER BY month;

-- name: GetTotalClicks :one
SELECT COUNT(*)::INT AS total
FROM clicks
WHERE link_id = $1;

-- name: GetClicksByUserAgent :many
SELECT user_agent, COUNT(*)::INT AS count
FROM clicks
WHERE link_id = $1
GROUP BY user_agent
ORDER BY count DESC;