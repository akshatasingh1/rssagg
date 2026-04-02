-- name: CreatePost :one
INSERT INTO posts (
    id,
    created_at, 
    updated_at, 
    title, 
    description, 
    published_at, 
    url,     
    feed_id,
    summary
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (url) DO NOTHING
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.* from posts
join feed_follows on posts.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2 OFFSET $3;

-- name: GetNextPostToSummarize :one
SELECT * FROM posts 
WHERE summary IS NULL 
ORDER BY published_at DESC 
LIMIT 1;

-- name: UpdatePostSummary :exec
UPDATE posts 
SET summary = $2, updated_at = NOW() 
WHERE id = $1;


-- name: SearchPostsForUser :many
SELECT posts.* FROM posts
JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
AND to_tsvector('english', posts.title) @@ plainto_tsquery('english', $2)
ORDER BY posts.published_at DESC 
LIMIT $3 OFFSET $4;