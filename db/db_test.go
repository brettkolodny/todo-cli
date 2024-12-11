package db

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/glebarez/go-sqlite"
)

func OpenTestDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func TestCreateTodo(t *testing.T) {
	db := OpenTestDB()
	defer db.Close()

	// TODO
}
