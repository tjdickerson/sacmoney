package database

import (
	"database/sql"
	"fmt"
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

	_, err = stmt.Exec(sql.Named("name", a.Name))
	if err != nil {
		return fmt.Errorf("Error inserting account: %s", err)
	}

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

func fetchAllAccounts() ([]Account, error) {
	stmt, err := dbc.db.Prepare("select a.id, a.name from accounts a order by a.Name")
	if err != nil {
		return nil, fmt.Errorf("Error preparing to fetch accounts: %s", err)
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("Error fetching accounts: %s", err)
	}

	var id int
	var name string
	var accounts []Account
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("Error reading accounts: %s", err)
		}

		accounts = append(accounts, Account{
			Id:   id,
			Name: name,
		})
	}

	return accounts, nil
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
