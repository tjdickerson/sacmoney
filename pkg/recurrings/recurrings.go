package recurrings

import (
	"database/sql"
	"log"
	"time"
)

type RecurringTransaction struct {
	Id     int32
	Name   string
	Amount int64
	Day    uint8
}

func GetRecurringTransactions(db *sql.DB, accountId int32) ([]RecurringTransaction, error) {
	stmt, err := db.Prepare(Q_LAST_RECURRING_TRANSACTIONS)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(accountId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []RecurringTransaction
	var id int32
	var name string
	var amount int64
	var day uint8

	for rows.Next() {
		err = rows.Scan(&id, &name, &amount, &day)
		if err != nil {
			log.Printf("Error : %s\n", err)
		}

		results = append(results, RecurringTransaction{
			Id:     id,
			Name:   name,
			Amount: amount,
			Day:    day,
		})
	}

	return results, nil
}

func AddRecurringTransaction(db *sql.DB, accountId int32, recurring *RecurringTransaction) error {
	stmt, err := db.Prepare(INS_RECURRING_TRANSACTION)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = stmt.Exec(
		accountId,
		recurring.Name,
		recurring.Amount,
		recurring.Day,
		now)
	return err
}

func DeleteTransaction(db *sql.DB, recurringId int32) error {
	stmt, err := db.Prepare(DEL_RECURRING_TRANSACTION)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(recurringId)
	return err
}

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

const Q_LAST_RECURRING_TRANSACTIONS = `
	select rt.id
	     , rt.name
		 , rt.amount
	     , rt.occurrence_day
	from recurrings rt
	where account_id = ?
	order by rt.occurrence_day desc
			,rt.timestamp_added desc
	limit 50
`

const INS_RECURRING_TRANSACTION = `
	insert into recurrings (
		  account_id
		, name
	    , amount
	    , occurrence_day
	    , timestamp_added)
	values ( ?, ?, ?, ?, ?)
`
const DEL_RECURRING_TRANSACTION = `
	delete from recurrings where id = ?
`
