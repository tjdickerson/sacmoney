package main

import (
	"fmt"
	sacdb "internal/sacdb"
	"log"
)

func main() {
	fmt.Printf("sacmoney\n")

	db, err := sacdb.GetDatabase()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failure initializing database: %s\n", err))
	}

	defer db.Close()
}
