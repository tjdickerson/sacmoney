package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Crudder interface {
	insert() error
	delete() error
	update() error
}

type dbContext struct {
	db               *sql.DB
	currentAccountId int
}

var (
	dbc dbContext = dbContext{}
)

const DB_PATH = "sacmoney.db"
const DB_INIT_ERR = "Database not initialized. Call InitDatabase() before calling any other database functions. (Also defer CloseDatabase())"

func InitDatabase() error {
	_, existErr := os.Stat(DB_PATH)
	if errors.Is(existErr, os.ErrNotExist) {
		tdb, cErr := createSchema(DB_PATH)
		if cErr != nil {
			panic(fmt.Sprintf("Couldn't create the database: %s\n", cErr))
		}
		tdb.Close()
	}

	db, err := sql.Open("sqlite3", DB_PATH+"?cache=shared")
	db.SetMaxOpenConns(1)

	dbc.db = db
	return err
}

func CloseDatabase() error {
	if dbc.db != nil {
		return dbc.db.Close()
	}
	return nil
}

func Insert(c Crudder) error {
	return c.insert()
}

func Delete(c Crudder) error {
	return c.delete()
}

func Update(c Crudder) error {
	return c.update()
}

func FetchAllTransactions() ([]Transaction, error) {
	return fetchAllTransactions()
}

func FetchAllRecurrings() ([]Recurring, error) {
	return fetchAllRecurrings()
}

func FetchAllAccounts() ([]Account, error) {
	return fetchAllAccounts()
}

func GetDefaultAccount() (Account, error) {
	if dbc.db == nil {
		return Account{}, fmt.Errorf(DB_INIT_ERR)
	}
	account, err := getAccount(1)
	dbc.currentAccountId = account.Id
	return account, err
}

func HasAccount() bool {
	result, err := dbc.db.Query("select count(1) from accounts;")
	if err != nil {
		return false
	}

	var count int32 = 0
	if result.Next() {
		err = result.Scan(&count)
		if err != nil {
			return false
		}
	}

	result.Close()

	return count > 0
}

func createSchema(db_path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", db_path)

	if err != nil {
		return nil, err
	}

	createTable(db, CT_ACCOUNT)
	createTable(db, CT_CATEGORIES)
	createTable(db, CT_TRANSACTIONS)
	createTable(db, CT_RECURRINGS)

	return db, nil
}

func createTable(db *sql.DB, statement string) error {
	stmt, err := db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("Failure preparing statement: %s", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("Failure executing statement: %s", err)
	}

	return nil
}
