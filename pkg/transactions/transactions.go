package transactions

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Account struct {
	id           int32
	name         string
	transactions []Transaction
}

type Transaction struct {
	Id     int32
	Name   string
	Amount int64
	Date   int64
}

func (t *Transaction) ToCliString(width int) string {
	id := strconv.Itoa(int(t.Id))
	name := t.Name
	amount := fmt.Sprintf("$%.2f", float64(t.Amount)*float64(0.01))
	date := time.UnixMilli(t.Date).Format("02 Mon")

	padding := 9 // account for spacers between data elements
	width = width - padding
	amountWidth := 10
	idWidth := 10
	nameWidth := width - idWidth - amountWidth - len(date)

	preIdSpace := idWidth - len(id)
	preAmountSpace := amountWidth - len(amount)

	if len(name) > nameWidth {
		name = name[0:nameWidth-3] + "..."
	}

	id = strings.Repeat(" ", preIdSpace) + id
	amount = strings.Repeat(" ", preAmountSpace) + amount

	result := fmt.Sprintf("%s | %s | %s | %s", id, date, amount, name)
	return result
}

func GetLastTransactions(db *sql.DB, accountId int32) ([]Transaction, error) {
	stmt, err := db.Prepare(Q_LAST_TRANSACTIONS)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(accountId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []Transaction
	var id int32
	var name string
	var amount int64
	var date int64
	for rows.Next() {
		rows.Scan(&id, &name, &amount, &date)

		results = append(results, Transaction{
			Id:     id,
			Name:   name,
			Amount: amount,
			Date:   date,
		})
	}

	return results, nil
}

func AddTransaction(db *sql.DB, accountId int32, transaction Transaction) error {
	stmt, err := db.Prepare(INS_TRANSACTION)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(accountId, transaction.Name, transaction.Amount, transaction.Date)
	return err
}

func DeleteTransaction(db *sql.DB, transactionId int32) error {
	stmt, err := db.Prepare(DEL_TRANSACTION)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(transactionId)
	return err
}

func GetAvailable(db *sql.DB, accountId int32) (float64, error) {
	stmt, err := db.Prepare(Q_TOTAL_AVAIL)
	if err != nil {
		return 0, err
	}

	row, err := stmt.Query(accountId)
	if err != nil {
		return 0, err
	}

	defer row.Close()

	var amount int64
	if row.Next() {
		row.Scan(&amount)
	}

	result := float64(amount) * 0.01
	return result, nil
}

const INS_TRANSACTION = `
	insert into ledger (
		  account_id
		, name
	    , amount
	    , transaction_date)
	values ( ?, ?, ?, ?)
`

const Q_LAST_TRANSACTIONS = `
	select l.id
	     , l.name
		 , l.amount
	     , l.transaction_date
	from ledger l
	where account_id = ?
	order by l.transaction_date desc
	limit 10
`

const Q_TOTAL_AVAIL = `
	select sum(l.amount)
	from ledger l
	where account_id = ?
`

const DEL_TRANSACTION = `
	delete from ledger where id = ?
`
