package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	Id     int
	Name   string
	Amount int64
	Date   time.Time
}

func (t *Transaction) insert() error {
	stmt, err := dbc.db.Prepare(INS_TRANSACTION)
	if err != nil {
		return fmt.Errorf("Error preparing transaction for inserting: %s", err)
	}

	// values (@account_id, @name, @amount, @transaction_date, @timestamp_added)
	_, err = stmt.Exec(
		sql.Named("account_id", dbc.currentAccountId),
		sql.Named("name", t.Name),
		sql.Named("amount", t.Amount),
		sql.Named("transaction_date", t.Date.UnixMilli()),
		sql.Named("timestamp_added", time.Now().UnixMilli()),
	)

	if err != nil {
		return fmt.Errorf("Error inserting transaction: %s", err)
	}

	return nil
}

func (t *Transaction) update() error {
	stmt, err := dbc.db.Prepare(UPD_TRANSACTION)
	if err != nil {
		return fmt.Errorf("Error preparing update for transaction: %s", err)
	}

	_, err = stmt.Exec(
		sql.Named("id", t.Id),
		sql.Named("name", t.Name),
		sql.Named("amount", t.Amount),
	)

	if err != nil {
		return fmt.Errorf("Error updating transaction: %s", err)
	}

	return nil
}

func fetchAllTransactions() ([]Transaction, error) {
	stmt, err := dbc.db.Prepare(Q_TRANSACTIONS)
	if err != nil {
		return nil, fmt.Errorf("Error preparing for fetching transactions: %s", err)
	}

	rows, err := stmt.Query(sql.Named("account_id", dbc.currentAccountId))
	if err != nil {
		return nil, fmt.Errorf("Error fetching transactions: %s", err)
	}

	defer rows.Close()

	var results []Transaction
	var id int
	var name string
	var amount int64
	var date int64
	var utcDate time.Time
	utc, _ := time.LoadLocation("UTC")

	for rows.Next() {
		err = rows.Scan(&id, &name, &amount, &date)
		if err != nil {
			return nil, fmt.Errorf("Error reading transactions: %s", err)
		}

		utcDate = time.UnixMilli(date).In(utc)

		results = append(results, Transaction{
			Id:     id,
			Name:   name,
			Amount: amount,
			Date:   utcDate,
		})
	}

	return results, nil
}

func (t *Transaction) delete() error {
	stmt, err := dbc.db.Prepare(DEL_TRANSACTION)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(sql.Named("id", t.Id))
	return err
}

func (t *Transaction) ToCliString(width int) string {
	id := strconv.Itoa(int(t.Id))
	name := t.Name
	amount := fmt.Sprintf("$%.2f", float64(t.Amount)*float64(0.01))
	date := t.Date.Format("Mon 02 Jan")

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

const INS_TRANSACTION = `
	insert into transactions (
		  account_id
		, name
	    , amount
	    , transaction_date
	    , timestamp_added)
	values (@account_id, @name, @amount, @transaction_date, @timestamp_added)
`

const UPD_TRANSACTION = `
	update transactions 
	set name = @name,
	    amount = @amount
	where id = @id;
`

const Q_TRANSACTIONS = `
	select t.id
	     , t.name
		 , t.amount
	     , t.transaction_date
	from transactions t
	where account_id = @account_id
	order by t.transaction_date desc
			,t.timestamp_added desc
`

const DEL_TRANSACTION = `
	delete from transactions where id = @id
`

const CT_TRANSACTIONS = `
	create table if not exists transactions (
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
