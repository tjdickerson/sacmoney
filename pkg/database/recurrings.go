package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Recurring struct {
	Id     int
	Name   string
	Amount int64
	Day    uint8
}

func getRecurringById(id int) (Recurring, error) {
	stmt, err := dbc.db.Prepare("select id, name, amount from recurrings where id = @id")
	if err != nil {
		return Recurring{}, fmt.Errorf("Error preparing recurring by id: %s", err)
	}

	row := stmt.QueryRow(sql.Named("id", id))

	recurring := Recurring{}
	if err := row.Scan(&recurring.Id, &recurring.Name, &recurring.Amount); err != nil {
		return recurring, fmt.Errorf("Error retrieving recurring: %s", err)
	}

	return recurring, nil
}

func getNetRecurringBalance() (int64, error) {
	stmt, err := dbc.db.Prepare("select sum(amount) from recurrings where account_id = @account_id")
	if err != nil {
		return 0, fmt.Errorf("Error preparing statement for recurring net balance: %s", err)
	}

	row := stmt.QueryRow(sql.Named("account_id", dbc.currentAccountId))
	var balance int64
	err = row.Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("Error getting recurring net balance: %s", err)
	}

	return balance, nil
}

func fetchAllRecurrings() ([]Recurring, error) {
	stmt, err := dbc.db.Prepare(Q_RECURRING_TRANSACTIONS)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(sql.Named("account_id", dbc.currentAccountId))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []Recurring
	var id int
	var name string
	var amount int64
	var day uint8

	for rows.Next() {
		err = rows.Scan(&id, &name, &amount, &day)
		if err != nil {
			return nil, fmt.Errorf("Error reading recurring transactions : %s", err)
		}

		results = append(results, Recurring{
			Id:     id,
			Name:   name,
			Amount: amount,
			Day:    day,
		})
	}

	return results, nil
}

func (r *Recurring) insert() error {
	stmt, err := dbc.db.Prepare(INS_RECURRING_TRANSACTION)
	if err != nil {
		return fmt.Errorf("Error preparing recurring for insert: %s", err)
	}

	// values (@account_id, @name, @amount, @occurrence_day, @timestamp_added)
	_, err = stmt.Exec(
		sql.Named("account_id", dbc.currentAccountId),
		sql.Named("name", r.Name),
		sql.Named("amount", r.Amount),
		sql.Named("occurrence_day", r.Day),
		sql.Named("timestamp_added", time.Now().UnixMilli()),
	)
	if err != nil {
		return fmt.Errorf("Error inserting recurring: %s", err)
	}

	return nil
}

func (r *Recurring) update() error {
	return fmt.Errorf("Not yet implemented")
}

func (r *Recurring) delete() error {
	stmt, err := dbc.db.Prepare(DEL_RECURRING_TRANSACTION)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(sql.Named("id", r.Id))
	return err
}

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

const Q_RECURRING_TRANSACTIONS = `
	select rt.id
	     , rt.name
		 , rt.amount
	     , rt.occurrence_day
	from recurrings rt
	where account_id = @account_id
	order by rt.occurrence_day 
			,rt.timestamp_added desc
`

const INS_RECURRING_TRANSACTION = `
	insert into recurrings (
		  account_id
		, name
	    , amount
	    , occurrence_day
	    , timestamp_added)
	values (@account_id, @name, @amount, @occurrence_day, @timestamp_added)
`

const UPD_RECURRING_TRANSACTION = `
	update recurrings r 
	set r.name = @name,
		r.amount = @amount,
		r.occurrence_day = @day
    where r.id = @id;
`

const DEL_RECURRING_TRANSACTION = `
	delete from recurrings where id = @id;
`
