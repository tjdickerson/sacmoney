package cli

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	setup "tjdickerson/sacmoney/pkg/setup"
	trn "tjdickerson/sacmoney/pkg/transactions"
	utils "tjdickerson/sacmoney/pkg/utils"
)

const WIDTH = 100

func Run() {
	db, err := setup.GetDatabase()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failure initializing database: %s\n", err))
	}

	defer db.Close()

	var accountId int32
	var accountName string
	if !setup.HasAccount(db) {
		log.Printf("You have no accounts configured.\n")

		accountName = getStringFromUser("Enter name for account: ")
		accountName = strings.TrimSpace(accountName)
		log.Printf("Creating (%s) account.\n", accountName)

		err = setup.CreateNewAccount(db, accountName)
		if err != nil {
			log.Printf("Failed to create account: %s\n", err)
		} else {
			log.Printf("Account created.\n")
		}
	}

	accountId, accountName = setup.GetDefaultAccount(db)

	var msg string
	if err == nil {
		running := true
		for running {
			clear()
			displayMenu(db, accountId, accountName, msg)

			option := getStringFromUser("> ")
			option = strings.TrimSpace(option)
			switch option {
			case "q":
				running = false
				break
			case "1":
				createWithdrawal(db, accountId)
				break
			case "2":
				createDeposit(db, accountId)
				break
			case "d":
				msg = deleteEntry(db)
				break
			}
		}
	}
}

func createDeposit(db *sql.DB, accountId int32) {
	name := getStringFromUser("Deposit Name > ")
	amount := getStringFromUser("Deposit Amount > ")

	iAmount := utils.GetCentsFromString(amount)
	transaction := &trn.Transaction{
		Name:   name,
		Amount: iAmount,
		Date:   time.Now(),
	}

	err := trn.AddTransaction(db, accountId, *transaction)
	if err != nil {
		log.Printf("Error adding transaction: %s\n", err)
	}
}

func createWithdrawal(db *sql.DB, accountId int32) {
	name := getStringFromUser("Debit Name > ")
	amount := getStringFromUser("Debit Amount > ")

	iAmount := utils.GetCentsFromString(amount)
	iAmount = iAmount * -1
	transaction := &trn.Transaction{
		Name:   name,
		Amount: iAmount,
		Date:   time.Now(),
	}

	err := trn.AddTransaction(db, accountId, *transaction)
	if err != nil {
		log.Printf("Error adding transaction: %s\n", err)
	}
}

func deleteEntry(db *sql.DB) string {
	entry := getStringFromUser("Entry ID > ")
	entry = strings.TrimSpace(entry)
	iEntry, err := strconv.Atoi(entry)
	if err != nil {
		log.Printf("Invalid entry: %s\n", err)
		return "Invalid Entry"
	}

	err = trn.DeleteTransaction(db, int32(iEntry))
	if err != nil {
		log.Printf("Error deleting transaction: %s\n", err)
		return "Couldn't Delete Transaction"
	}

	return "Transaction Deleted"
}

func clear() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func displayMenu(db *sql.DB, accountId int32, accountName string, msg string) {
	fmt.Printf("%s\n", headerRow(accountName))
	fmt.Printf("%s\n", msg)
	fmt.Printf("%s\n\n", availableRow(db, accountId))

	top10, err := trn.GetLastTransactions(db, accountId)
	if err != nil {
		log.Printf("Error getting last transactions: %s\n", err)
	} else {
		for _, transaction := range top10 {
			fmt.Printf("%s\n", transaction.ToCliString(WIDTH))
		}
	}

	fmt.Printf("\n%s\n", commandRow())
}

func headerRow(accountName string) string {
	currentTime := time.Now()
	var header string

	title := fmt.Sprintf("sacmoney - %s", accountName)
	date := currentTime.Format("Mon  02 Jan 2006")

	space := WIDTH - len(title) - len(date)
	spacer := strings.Repeat(" ", space)

	header = fmt.Sprintf("%s%s%s", title, spacer, date)

	return header
}

func availableRow(db *sql.DB, accountId int32) string {
	available, err := trn.GetAvailable(db, accountId)
	var current string
	if err != nil {
		current = fmt.Sprintf("ERR %s", err)
	} else {
		current = fmt.Sprintf("Available: $%.2f", available)
	}

	return current
}

func commandRow() string {
	commands := "1) Debit  2) Deposit  d) Delete Entry    q) Quit"
	return commands
}

func getStringFromUser(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read input string.")
	}

	return strings.TrimSpace(input)
}
