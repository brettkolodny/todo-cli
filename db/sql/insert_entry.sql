
WITH target_todo AS (
    SELECT id 
    FROM todos 
    WHERE title = $1
    LIMIT 1
)
INSERT INTO entries (title, todo_id)
SELECT 
    $2,
    id 
FROM target_todo
