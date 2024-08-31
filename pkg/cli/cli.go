package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	db "tjdickerson/sacmoney/pkg/database"
	utils "tjdickerson/sacmoney/pkg/utils"
)

const WIDTH = 100

func Run() {

	err := db.InitDatabase()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failure initializing database: %s\n", err))
	}

	defer db.CloseDatabase()

	var accountName string
	if !db.HasAccount() {
		log.Printf("You have no accounts configured.\n")

		accountName = getStringFromUser("Enter name for account: ")
		accountName = strings.TrimSpace(accountName)
		log.Printf("Creating (%s) account.\n", accountName)

		a := &db.Account{Name: accountName}
		db.Insert(a)
	}

	account, err := db.GetDefaultAccount()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting account: %s", err))
	}

	var msg string
	running := true
	for running {
		clear()
		displayMenu(&account, msg)

		option := getStringFromUser("> ")
		option = strings.TrimSpace(option)
		switch option {
		case "q":
			running = false
			break
		case "1":
			createWithdrawal()
			break
		case "2":
			createDeposit()
			break
		case "d":
			msg = deleteEntry()
			break
		}
	}
}

func createDeposit() {
	name := getStringFromUser("Deposit Name > ")
	amount := getStringFromUser("Deposit Amount > ")

	iAmount := utils.GetCentsFromString(amount)
	transaction := &db.Transaction{
		Name:   name,
		Amount: iAmount,
		Date:   time.Now(),
	}

	err := db.Insert(transaction)
	if err != nil {
		log.Printf("Error adding transaction: %s\n", err)
	}
}

func createWithdrawal() {
	name := getStringFromUser("Debit Name > ")
	amount := getStringFromUser("Debit Amount > ")

	iAmount := utils.GetCentsFromString(amount)
	iAmount = iAmount * -1

	transaction := &db.Transaction{
		Name:   name,
		Amount: iAmount,
		Date:   time.Now(),
	}

	err := db.Insert(transaction)
	if err != nil {
		log.Printf("Error adding transaction: %s\n", err)
	}
}

func deleteEntry() string {
	entry := getStringFromUser("Entry ID > ")
	entry = strings.TrimSpace(entry)
	iEntry, err := strconv.Atoi(entry)
	if err != nil {
		log.Printf("Invalid entry: %s\n", err)
		return "Invalid Entry"
	}

	trn := &db.Transaction{Id: iEntry}
	err = db.Delete(trn)
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

func displayMenu(account *db.Account, msg string) {
	fmt.Printf("%s\n", headerRow(account.Name))
	fmt.Printf("%s\n", msg)

	amount := fmt.Sprintf("%.2f", float64(account.TotalAvailable)*0.10)
	fmt.Printf("%s\n\n", amount)

	top10, err := db.FetchAllTransactions()
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
