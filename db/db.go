package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

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
		createTableSql := `CREATE TABLE IF NOT EXISTS todos (
	id integer PRIMARY KEY autoincrement,
	created_at datetime DEFAULT CURRENT_TIMESTAMP,
	title text NOT NULL)`

		_, err = db.Exec(createTableSql)
		if err != nil {
			log.Fatalf("Could not create todos table: %v", err)
		}

	}

	return db
}

// Check if a file exists at the given path.
// This feels like something that may alread exist in the stdlib so it may be worth investigating
// if it can be removed for something built in.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
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
