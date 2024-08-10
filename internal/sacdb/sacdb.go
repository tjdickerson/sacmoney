package database

import (
	"database/sql"
	"errors"
	sqlite3 "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

const DB_PATH = "./sacmoney.db"

func GetDatabase() (*sql.DB, error) {
	ver, _, _ := sqlite3.Version()
	log.Printf("sqlite %s\n", ver)

	_, existErr := os.Stat(DB_PATH)
	if errors.Is(existErr, os.ErrNotExist) {
		log.Printf("Database not found. Initializing database...\n")
		return createSchema(DB_PATH)
	}

	log.Printf("Loading database...\n")
	return sql.Open("sqlite3", DB_PATH)
}

func createSchema(db_path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", db_path)

	if err != nil {
		return nil, err
	}

	var testTable string = "create table if not exists test (id integer primary key, name varchar(50))"

	statement, err := db.Prepare(testTable)
	statement.Exec()

	return db, nil
}
