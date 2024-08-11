package main

import (
	"bufio"
	"fmt"
	sacdb "internal/sacdb"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Printf("sacmoney\n")

	db, err := sacdb.GetDatabase()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failure initializing database: %s\n", err))
	}

	defer db.Close()

	var accountName string
	if !sacdb.HasAccount(db) {
		log.Printf("You have no accounts configured.\n")

		accountName = getStringFromUser("Enter name for account: ")
		accountName = strings.TrimSpace(accountName)
		log.Printf("Creating (%s) account.\n", accountName)

		err = sacdb.CreateNewAccount(db, accountName)
		if err != nil {
			log.Printf("Failed to create account: %s\n", err)
		} else {
			log.Printf("Account created.\n")
		}
	} else {
		accountName = sacdb.GetDefaultAccount(db)
	}

	if err == nil {
		clear()
		displayMenu(accountName)
	}
}

func clear() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func displayMenu(accountName string) {
	fmt.Printf("%s\n", headerRow(accountName))
}

func headerRow(accountName string) string {
	currentTime := time.Now()
	var header string

	title := fmt.Sprintf("sacmoney - %s", accountName)
	date := currentTime.Format("Mon  02 Jan 2006")

	headLen := 80
	space := headLen - len(title) - len(date)
	spacer := strings.Repeat(" ", space)

	header = fmt.Sprintf("%s%s%s", title, spacer, date)

	return header
}

func getStringFromUser(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read input string.")
	}

	return input
}
