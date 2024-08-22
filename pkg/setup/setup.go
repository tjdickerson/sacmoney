package setup

import (
	"database/sql"
	"errors"
	"fmt"
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
		tdb, cErr := createSchema(DB_PATH)
		if cErr != nil {
			log.Fatal(fmt.Sprintf("Couldn't create the database: %s\n", cErr))
		}
		tdb.Close()
	}

	log.Printf("Loading database...\n")
	db, err := sql.Open("sqlite3", DB_PATH+"?cache=shared")
	db.SetMaxOpenConns(1)
	return db, err
}

func HasAccount(db *sql.DB) bool {
	result, err := db.Query("select count(1) from accounts;")
	if err != nil {
		log.Printf("Failed to find accounts table.\n")
		log.Printf("  %s\n", err)
		return false
	}

	var count int32 = 0
	if result.Next() {
		result.Scan(&count)
	}

	result.Close()

	return count > 0
}

func GetDefaultAccount(db *sql.DB) (int32, string) {
	result, err := db.Query("select id, name from accounts where id = 1")
	if err != nil {
		log.Printf("Failed to find account...\n")
		log.Fatal(err)
		unk := "unknown"
		return 0, unk
	}

	defer result.Close()
	var accountId int32
	var accountName string
	if result.Next() {
		result.Scan(&accountId, &accountName)
	}

	log.Printf("Using default Account: %s\n", accountName)
	return accountId, accountName
}

func CreateNewAccount(db *sql.DB, name string) error {
	stmt, err := db.Prepare("insert into accounts(name) values(?);")
	if err != nil {
		log.Printf("Couldn't prepare new account.\n")
		return err
	}

	log.Printf("About to insert the thing: %s...\n", name)

	_, err = stmt.Exec(name)
	if err != nil {
		log.Printf(fmt.Sprintf("Couldn't create new account: %s\n", err))
		return err
	}

	return nil
}

func createSchema(db_path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", db_path)

	if err != nil {
		return nil, err
	}

	createTable(db, CT_ACCOUNT)
	createTable(db, CT_CATEGORIES)
	createTable(db, CT_LEDGER)
	createTable(db, CT_RECURRINGS)

	return db, nil
}

func createTable(db *sql.DB, statement string) error {

	stmt, err := db.Prepare(statement)
	if err != nil {
		log.Printf("Failure preparing statement: %s\n%s\n", err, statement)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Printf("Failure executing statement: %s\n%s\n", err, statement)
		return err
	}

	return nil
}

const CT_ACCOUNT = `
	create table if not exists accounts (
		id integer primary key,
	    name varchar(100)
	);
	`

const CT_LEDGER = `
	create table if not exists ledger (
		id integer primary key,
		transaction_date integer,
		amount integer,
	    name varchar(1000),
	    account_id integer,
	    category_id integer,
		timestamp_added integer,
	    foreign key(account_id) references accounts(id),
	    foreign key(category_id) references categories(id)
	);
`

const CT_CATEGORIES = `
	create table if not exists categories (
		id integer primary key,
		account_id integer
		name varchar(100),
		foreign key(account_id) references accounts(id)
	);
`

// TODO: move this out of setup
const CT_RECURRINGS = `
	create table if not exists recurrings (
		id integer primary key,
		account_id integer,
		category_id integer,
		name varchar(100),
		occurrence_day integer,
		amount integer,
	    timestamp_added integer,
		foreign key(account_id) references accounts(id),
		foreign key(category_id) references categories(id)
	);
`
