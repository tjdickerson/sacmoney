package sacmoney

import (
	"database/sql"
	setup "github.com/tjdickerson/sacmoney/internal/setup"
)

func GetDatabase() *sql.DB {
	return setup.GetDatabase()
}
