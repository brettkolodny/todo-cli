-- Create the main todos table first since it will be referenced by entries
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL,
    CONSTRAINT unique_todo_title UNIQUE (title)  -- Named constraint for better error handling
);

-- Create the entries table with a foreign key relationship to todos
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    todo_id INTEGER,
    FOREIGN KEY (todo_id) REFERENCES todos(id)
        ON DELETE CASCADE  -- Automatically delete entries when their todo is deleted
        ON UPDATE CASCADE  -- Automatically update entries when their todo's ID changes
);

-- Create an index to improve query performance when looking up entries by todo
CREATE INDEX idx_entries_todo_id ON entries(todo_id);
