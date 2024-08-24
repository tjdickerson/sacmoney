package database

import (
	"fmt"
)

type Category struct {
}

func (c *Category) insert() error {
	return fmt.Errorf("Not yet implemented.")
}

func (c *Category) update() error {
	return fmt.Errorf("Not yet implemented.")
}

func (c *Category) delete() error {
	return fmt.Errorf("Not yet implemented.")
}

const CT_CATEGORIES = `
	create table if not exists categories (
		id integer primary key,
		account_id integer
		name varchar(100),
		foreign key(account_id) references accounts(id)
	);
`
