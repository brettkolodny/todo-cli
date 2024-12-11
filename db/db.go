package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type TodoTableRow struct {
	Title     string
	CreatedAt time.Time
}

type TodoTable []TodoTableRow

func (table TodoTable) Print() {
	var builder strings.Builder

	builder.WriteString("  | Title \t Created at\n")

	for i, row := range table {
		builder.WriteString(fmt.Sprintf("%d | %s \t %s\n", i, row.Title, row.CreatedAt.String()))
	}

	fmt.Println(builder.String())
}

// Open the DB at ~/.config/todo.db or the path specified in TODO_DB_PATH
// If the DB did not exist before this call then create the needed tables
func Open() *sql.DB {
	path := getPath()

	// We're checking if the DB file exists before opening the DB connection because `sql.Open` creates
	// the file if it doesn't and we want to know if we should run the initial setup.
	//
	// There is probably a better way to tell whether we need to run the setup, or mmaybe it's fine
	// to run the setup every time, either through some migration or some more SQLite native means that
	// would be worth looking into in the future.
	alreadyExisted := fileExists(path)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}

	if !alreadyExisted {
		createTablesSql := `
-- Create the main todos table first since it will be referenced by entries
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL
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
CREATE INDEX idx_entries_todo_id ON entries(todo_id);`

		_, err = db.Exec(createTablesSql)
		if err != nil {
			log.Fatalf("Could not create todos table: %v", err)
		}

	}

	return db
}

// Create a todo list with the given title
func CreateTodo(db *sql.DB, title string) error {
	stmt, err := db.Prepare("INSERT INTO todos (title) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(title); err != nil {
		return err
	}

	return nil
}

// List all todo lists within the database
func ListTodos(db *sql.DB) (TodoTable, error) {
	query := "SELECT title, created_at FROM todos"

	rows, err := db.Query(query)
	if err != nil {
		return TodoTable{}, err
	}
	defer rows.Close()

	var todoRows []TodoTableRow
	for rows.Next() {
		var title string
		var createdAt time.Time

		if err := rows.Scan(&title, &createdAt); err != nil {
			return TodoTable{}, err
		}

		row := TodoTableRow{
			Title:     title,
			CreatedAt: createdAt,
		}

		todoRows = append(todoRows, row)
	}

	return todoRows, nil
}

// Check if a file exists at the given path.
// This feels like something that may alread exist in the stdlib so it may be worth investigating
// if it can be removed for something built in.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatalf("Could not check if DB file at path %v exists", path)
		}
	}

	return true
}

// Get the path to the DB by first checking if the env variable TODO_DB_PATH is set, otherwise
// defaulting to ~/.config/todo.db
func getPath() string {
	db_path := os.Getenv("TODO_DB_PATH")

	if db_path == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Could not load the HOME directory: ", err)
		}

		pathToDb := filepath.Join(dirname, ".config", "todo")
		mkdirErr := os.MkdirAll(pathToDb, os.ModePerm)
		if mkdirErr != nil {
			log.Fatalf("Unable to create folder at path %v", pathToDb)
		}
		return filepath.Join(dirname, ".config", "todo", "todo.db")
	} else {
		return db_path
	}
}
