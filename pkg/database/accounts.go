package database

import (
	"database/sql"
	"fmt"
	"log"
)

type Account struct {
	Id             int
	Name           string
	TotalAvailable int64
}

func (a *Account) insert() error {
	stmt, err := dbc.db.Prepare("insert into accounts(name) values(@name);")
	if err != nil {
		return fmt.Errorf("Error preparing account for insert: %s", err)
	}

	test, err := stmt.Exec(sql.Named("name", a.Name))
	if err != nil {
		return fmt.Errorf("Error inserting account: %s", err)
	}

	log.Printf("Result of insert: %v\n", test)
	return nil
}

func (a *Account) update() error {
	return fmt.Errorf("Not yet supported")
}

func (a *Account) delete() error {
	return fmt.Errorf("Not yet supported")
}

func getAccount(id int) (Account, error) {
	stmt, err := dbc.db.Prepare(Q_GET_ACCOUNT)
	if err != nil {
		return Account{}, fmt.Errorf("Error preparing fetching account: %s", err)
	}

	row := stmt.QueryRow(sql.Named("id", id))

	var aid int
	var name string
	var total int64
	err = row.Scan(&aid, &name, &total)
	if err != nil {
		return Account{}, fmt.Errorf("Error reading account: %s", err)
	}

	return Account{
		Id:             aid,
		Name:           name,
		TotalAvailable: total,
	}, nil

}

const Q_GET_ACCOUNT = `
	select a.id
	     , a.name
	     , coalesce(sum(t.amount), 0) as total_available
	from accounts a
	left join transactions t on a.id = t.account_id
	where a.id = @id
	group by a.id, a.name
`

const CT_ACCOUNT = `
	create table if not exists accounts (
		id integer primary key,
	    name varchar(100)
	);
`

const Q_TOTAL_AVAIL = `
	select sum(t.amount)
	from transactions t
	where account_id = @account_id
`
