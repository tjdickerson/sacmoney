package main

import (
	"database/sql"
)

type Transaction struct {
	name     string
	category string
	amount   int64
}

func AddTransaction(db *sql.DB, transaction *Transaction) {

}
