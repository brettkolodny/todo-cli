SELECT
	entries.title,
	entries.completed
FROM
	todos
	INNER JOIN entries ON entries.todo_id = todos.id
where
	todos.title = $1
