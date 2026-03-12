-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id,created_at,updated_at,user_id,feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds on feeds.id = inserted_feed_follow.feed_id
INNER JOIN users on users.id = inserted_feed_follow.user_id
;

-- name: GetFeedFollowsForUser :many
SELECT
    ff.*,
    f.name AS feed_name,
    u.name AS user_name
FROM feed_follows ff
JOIN users u
  ON ff.user_id = u.id
JOIN feeds f
  ON ff.feed_id = f.id
WHERE u.name = $1
;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2
;
