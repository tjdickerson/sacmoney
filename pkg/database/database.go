package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
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

const DbInitError = "Database not initialized. Call InitDatabase() before calling any other database functions. (Also defer CloseDatabase())"

func InitDatabase(dbPath string, isRollover bool) error {
	var currentAccount Account
	var recurrings []Recurring
	if isRollover {
		a, err := getAccount(dbc.currentAccountId)
		if err != nil {
			return fmt.Errorf("Error getting current account information for rollover: %s", err)
		}
		currentAccount = a

		r, err := fetchAllRecurrings()
		if err != nil {
			return fmt.Errorf("Error getting recurring transactions for rollover: %s", err)
		}
		recurrings = r

		if dbc.db != nil {
			dbc.db.Close()
			dbc.db = nil
		}
	}

	_, existErr := os.Stat(dbPath)
	if errors.Is(existErr, os.ErrNotExist) {
		tdb, cErr := createSchema(dbPath)
		if cErr != nil {
			panic(fmt.Sprintf("Couldn't create the database: %s\n", cErr))
		}
		tdb.Close()
	}

	db, err := sql.Open("sqlite3", dbPath+"?cache=shared")
	db.SetMaxOpenConns(1)
	dbc.db = db

	if isRollover {
		err := rolloverDatabase(currentAccount, recurrings)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error occurred during rollover: %s\n", err))
		}
	}

	return err
}

func CloseDatabase() error {
	if dbc.db != nil {
		if err := dbc.db.Close(); err != nil {
			return err
		} else {
			dbc.db = nil
		}
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
		return Account{}, fmt.Errorf(DbInitError)
	}
	account, err := getAccount(1)
	dbc.currentAccountId = account.Id
	return account, err
}

func CreateTransactionFromRecurring(id int) error {
	recurring, err := getRecurringById(id)
	if err != nil {
		return err
	}

	newTrans := &Transaction{
		Name:   recurring.Name,
		Amount: recurring.Amount,
		Date:   time.Now(),
	}

	return newTrans.insert()
}

func GetRecurringNetBalance() (int64, error) {
	return getNetRecurringBalance()
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

func rolloverDatabase(account Account, recurrings []Recurring) error {
	if err := account.insert(); err != nil {
		return fmt.Errorf("Error rolling over account information: %s", err)
	}

	for _, r := range recurrings {
		if err := r.insert(); err != nil {
			return fmt.Errorf("Error rolling over recurring transactions: %s", err)
		}
	}

	initialTransaction := &Transaction{
		Name:   "Starting Balance",
		Amount: account.TotalAvailable,
		Date:   time.Now(),
	}
	if err := initialTransaction.insert(); err != nil {
		return fmt.Errorf("Error creating initial transaction for starting balance.")
	}

	return nil
}
