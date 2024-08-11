package main

import (
	"bufio"
	"fmt"
	sacm "github.com/tjdickerson/sacmoney"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Printf("sacmoney\n")

	db, err := sacm.GetDatabase()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failure initializing database: %s\n", err))
	}

	defer db.Close()

	var accountName string
	if !sacm.HasAccount(db) {
		log.Printf("You have no accounts configured.\n")

		accountName = getStringFromUser("Enter name for account: ")
		accountName = strings.TrimSpace(accountName)
		log.Printf("Creating (%s) account.\n", accountName)

		err = sacm.CreateNewAccount(db, accountName)
		if err != nil {
			log.Printf("Failed to create account: %s\n", err)
		} else {
			log.Printf("Account created.\n")
		}
	} else {
		accountName = sacm.GetDefaultAccount(db)
	}

	if err == nil {
		running := true
		for running {
			clear()
			displayMenu(accountName)

			option := getStringFromUser("> ")
			option = strings.TrimSpace(option)
			switch option {
			case "q":
				running = false
				break
			case "1":
				fmt.Printf("soon...")
				break
			case "2":
				fmt.Printf("soon...")
				break
			}
		}
	}
}

func clear() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func displayMenu(accountName string) {
	fmt.Printf("%s\n\n", headerRow(accountName))
	fmt.Printf("%s\n\n", availableRow(0))
	fmt.Printf("%s\n", commandRow())
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

func availableRow(totalAvail int64) string {
	var div float32 = 0.10
	var available = float32(totalAvail) * div
	current := fmt.Sprintf("Available: $%.2f", available)
	return current
}

func commandRow() string {
	commands := "1) Debit  2) Deposit    q) Quit"
	return commands
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
