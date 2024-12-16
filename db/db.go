package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	singleBullet    = "─"
	startingBullet  = "┌╴"
	connectorBullet = "├╴"
	endingBullet    = "└╴"
)

var (
	//go:embed sql/setup.sql
	setupSql string

	//go:embed sql/get_todos.sql
	getTodosSql string

	//go:embed sql/get_entries.sql
	getEntriesSql string

	//go:embed sql/insert_todo.sql
	insertTodoSql string

	//go:embed sql/insert_entry.sql
	insertEntrySql string
)

type TodosTableRow struct {
	Title     string
	CreatedAt time.Time
}

type TodoTable []TodosTableRow

type TodoEntry struct {
	Title     string
	Completed bool
}

type TodoList struct {
	Title   string
	Entries []TodoEntry
}

func (list TodoList) Print() {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s\n", list.Title))

	if len(list.Entries) == 1 {
		row := list.Entries[0]
		var checkBox string
		if row.Completed {
			checkBox = "[x]"
		} else {
			checkBox = "[ ]"
		}
		builder.WriteString(fmt.Sprintf("%s %s %s\n", singleBullet, checkBox, row.Title))
	} else {
		for i, row := range list.Entries {
			var bullet string
			switch i {
			case 0:
				bullet = startingBullet
			case len(list.Entries) - 1:
				bullet = endingBullet
			default:
				bullet = connectorBullet
			}

			var checkBox string
			if row.Completed {
				checkBox = "[x]"
			} else {
				checkBox = "[ ]"
			}

			builder.WriteString(fmt.Sprintf("%s %s %s\n", bullet, checkBox, row.Title))
		}
	}

	fmt.Println(builder.String())
}

func (table TodoTable) Print() {
	var builder strings.Builder

	if len(table) == 1 {
		row := table[0]
		builder.WriteString(fmt.Sprintf("%s %s\n", singleBullet, row.Title))
	} else {
		for i, row := range table {
			var bullet string
			switch i {
			case 0:
				bullet = startingBullet
			case len(table) - 1:
				bullet = endingBullet
			default:
				bullet = connectorBullet
			}

			builder.WriteString(fmt.Sprintf("%s %s\n", bullet, row.Title))
		}
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
		setupDb(db)
	}

	return db
}

// Create a todo list with the given title
func CreateTodo(db *sql.DB, title string) error {
	stmt, err := db.Prepare(insertTodoSql)
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
	rows, err := db.Query(getTodosSql)
	if err != nil {
		return TodoTable{}, err
	}
	defer rows.Close()

	var todoRows []TodosTableRow
	for rows.Next() {
		var title string
		var createdAt time.Time

		if err := rows.Scan(&title, &createdAt); err != nil {
			return TodoTable{}, err
		}

		row := TodosTableRow{
			Title:     title,
			CreatedAt: createdAt,
		}

		todoRows = append(todoRows, row)
	}
	if err := rows.Err(); err != nil {
		return TodoTable{}, err
	}

	return todoRows, nil
}

// List
func ListEntries(db *sql.DB, todoName string) (TodoList, error) {
	rows, err := db.Query(getEntriesSql, todoName)
	if err != nil {
		return TodoList{}, err
	}

	var entries []TodoEntry
	for rows.Next() {
		var title string
		var completed bool

		if err := rows.Scan(&title, &completed); err != nil {
			return TodoList{}, err
		}

		entries = append(entries, TodoEntry{Title: title, Completed: completed})
	}

	return TodoList{Title: todoName, Entries: entries}, nil
}

// List
func InsertEntry(db *sql.DB, todoName string, title string) error {
	stmt, err := db.Prepare(insertEntrySql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(todoName, title); err != nil {
		return err
	}

	return nil
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

func setupDb(db *sql.DB) {
	_, err := db.Exec(setupSql)
	if err != nil {
		log.Fatalf("Could not create todos table: %v", err)
	}
}
