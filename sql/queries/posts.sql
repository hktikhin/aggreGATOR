-- name: CreatePost :one
INSERT INTO posts (id,created_at,updated_at,title,url,description,published_at,feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT
  p.*
FROM feed_follows ff
JOIN users u
  ON ff.user_id = u.id
JOIN posts p
  ON ff.feed_id = p.feed_id
WHERE u.id = $1
ORDER BY published_at DESC
LIMIT $2
;
