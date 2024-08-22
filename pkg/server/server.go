package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	rec "sacdev/sacmoney/pkg/recurrings"
	setup "sacdev/sacmoney/pkg/setup"
	trn "sacdev/sacmoney/pkg/transactions"
	utils "sacdev/sacmoney/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type sacmoneyInfo struct {
	db          *sql.DB
	accountId   int32
	accountName string
}

type postHandler struct {
	info    *sacmoneyInfo
	handler func(h *postHandler, r *http.Request) string
}

func (h *postHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html := h.handler(h, r)
	io.WriteString(w, html)
}

type AddAccountJson struct {
	Name string
}

func fnAddAccountHandler(h *postHandler, r *http.Request) string {
	var aaj AddAccountJson
	err := json.NewDecoder(r.Body).Decode(&aaj)
	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	name := html.EscapeString(strings.TrimSpace(aaj.Name))
	err = setup.CreateNewAccount(h.info.db, name)
	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	accountId, accountName := setup.GetDefaultAccount(h.info.db)
	h.info.accountId = accountId
	h.info.accountName = accountName

	return "Account created."
}

type NewTransactionJson struct {
	Date   string
	Name   string
	Amount string
}

func fnAddTransactionHandler(h *postHandler, r *http.Request) string {
	var ntj NewTransactionJson
	err := json.NewDecoder(r.Body).Decode(&ntj)

	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		fmt.Printf("%s\n", serr)
		return serr
	}

	date, _ := time.Parse("2006-01-02", ntj.Date)
	name := html.EscapeString(strings.TrimSpace(ntj.Name))
	iAmount := utils.GetCentsFromString(ntj.Amount)

	transaction := &trn.Transaction{
		Name:   name,
		Amount: iAmount,
		Date:   date,
	}

	err = trn.AddTransaction(h.info.db, h.info.accountId, *transaction)
	if err != nil {
		serr := fmt.Sprintf("Error adding transaction: %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	newAmount, _ := trn.GetAvailable(h.info.db, h.info.accountId)
	sNewAmount := fmt.Sprintf("%.2f", newAmount)

	return fmt.Sprintf("Added transaction.::%s", sNewAmount)
}

type AddRecurringJson struct {
	Name   string
	Date   string
	Amount string
}

func fnAddRecurringHandler(h *postHandler, r *http.Request) string {
	var arj AddRecurringJson
	err := json.NewDecoder(r.Body).Decode(&arj)
	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	name := html.EscapeString(strings.TrimSpace(arj.Name))
	iAmount := utils.GetCentsFromString(arj.Amount)
	date, err := time.Parse("2006-01-02", arj.Date)
	if err != nil {
		serr := fmt.Sprintf("Error parsing date: %s", err)
		log.Printf("%s\n", serr)
	}

	recurring := &rec.RecurringTransaction{
		Name:   name,
		Amount: iAmount,
		Day:    uint8(date.Day()),
	}

	err = rec.AddRecurringTransaction(h.info.db, h.info.accountId, recurring)
	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	return "Recurring transaction created."
}

type DeleteInfoJson struct {
	Id string
}

func fnDeleteTransactionHandler(h *postHandler, r *http.Request) string {
	var dij DeleteInfoJson
	err := json.NewDecoder(r.Body).Decode(&dij)

	if err != nil {
		serr := fmt.Sprintf("Error deleting (JSON): %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	trnId, err := strconv.Atoi(dij.Id)
	if err != nil {
		serr := fmt.Sprintf("Error deleting (ID error): %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	err = trn.DeleteTransaction(h.info.db, int32(trnId))
	if err != nil {
		serr := fmt.Sprintf("Error deleting (DB): %s", err)
		log.Printf("%s\n", serr)
		return serr
	}

	newAmount, _ := trn.GetAvailable(h.info.db, h.info.accountId)
	sNewAmount := fmt.Sprintf("%.2f", newAmount)

	return fmt.Sprintf("Transaction deleted.::%s", sNewAmount)
}

type TransactionData struct {
	Id              string
	TransactionDate string
	Name            string
	AmountClass     string
	Amount          string
}

func fnTransactionsHandler(h *postHandler, r *http.Request) string {
	transactions, err := trn.GetLastTransactions(h.info.db, h.info.accountId)

	if err != nil {
		return fmt.Sprintf("<div class=\"error\">%s</div>", err)
	}

	var htmlTmpl bytes.Buffer
	tmpl, _ := template.ParseFiles("templates/transaction_list.html")
	for _, t := range transactions {

		strDate := t.Date.Format("Mon 02 Jan")
		amount := fmt.Sprintf("%.2f", float64(t.Amount)*float64(0.01))

		amountClass := "amount"
		if t.Amount > 0 {
			amountClass = "amount pos"
		} else if t.Amount < 0 {
			amountClass = "amount neg"
		}

		name := html.EscapeString(t.Name)

		td := &TransactionData{
			Id:              fmt.Sprintf("%d", t.Id),
			TransactionDate: strDate,
			Amount:          amount,
			AmountClass:     amountClass,
			Name:            name,
		}

		tmpl.Execute(&htmlTmpl, td)
	}

	return htmlTmpl.String()
}

type PageData struct {
	Title       string
	AccountName string
	Available   string
}

func fnHomeHandler(h *postHandler, r *http.Request) string {
	available, _ := trn.GetAvailable(h.info.db, h.info.accountId)
	tmpl, _ := template.ParseFiles("templates/home.html", "templates/tmpl_title.html")

	accountName := html.EscapeString(h.info.accountName)
	if h.info.accountId == 0 {
		accountName = "Click Accounts Link to Create an Account"
	}

	p := &PageData{
		Title:       "sacmoney",
		AccountName: accountName,
		Available:   fmt.Sprintf("%.2f", available),
	}

	var html bytes.Buffer
	tmpl.Execute(&html, p)
	return html.String()
}

type AccountsData struct {
	Title    string
	Accounts []string
}

func fnAccountsHandler(h *postHandler, r *http.Request) string {
	tmpl, err := template.ParseFiles("templates/accounts.html", "templates/tmpl_title.html")
	if err != nil {
		log.Printf("Failed to parse template: %s\n", err)
		return "ERROR"
	}

	p := &AccountsData{
		Title:    "sacmoney",
		Accounts: []string{h.info.accountName},
	}

	var htmlTmpl bytes.Buffer
	tmpl.Execute(&htmlTmpl, p)
	return htmlTmpl.String()
}

type RecurringPageData struct {
	Title string
}

func fnRecurringsPageHandler(h *postHandler, r *http.Request) string {
	tmpl, _ := template.ParseFiles("templates/recurrings.html", "templates/tmpl_title.html")

	rpd := &RecurringPageData{
		Title: "sacmoney",
	}

	var htmlTmpl bytes.Buffer
	tmpl.Execute(&htmlTmpl, rpd)
	return htmlTmpl.String()
}

type Recurrings struct {
	Id          string
	Day         string
	Name        string
	AmountClass string
	Amount      string
}

func fnRecurringsHandler(h *postHandler, r *http.Request) string {
	recurrings, err := rec.GetRecurringTransactions(h.info.db, h.info.accountId)

	if err != nil {
		serr := fmt.Sprintf("Error: %s", err)
		log.Printf("fnRecurringsHandler %s\n", serr)
		return serr
	}

	var htmlTmpl bytes.Buffer
	tmpl, _ := template.ParseFiles("templates/recurring_list.html")
	for _, r := range recurrings {
		amount := fmt.Sprintf("%.2f", float64(r.Amount)*float64(0.01))
		amountClass := "amount"
		if r.Amount > 0 {
			amountClass = "amount pos"
		} else if r.Amount < 0 {
			amountClass = "amount neg"
		}

		name := html.EscapeString(r.Name)

		td := &Recurrings{
			Id:          fmt.Sprintf("%d", r.Id),
			Day:         fmt.Sprintf("%d", r.Day),
			Amount:      amount,
			AmountClass: amountClass,
			Name:        name,
		}

		tmpl.Execute(&htmlTmpl, td)
	}

	log.Printf("Returning:\n%s\n", htmlTmpl.String())
	return htmlTmpl.String()
}

func Run() {
	var err error
	db, err := setup.GetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	acct, accountName := setup.GetDefaultAccount(db)

	info := &sacmoneyInfo{
		db:          db,
		accountId:   acct,
		accountName: accountName,
	}

	listTransactionsHandler := &postHandler{
		info:    info,
		handler: fnTransactionsHandler,
	}

	homeHandler := &postHandler{
		info:    info,
		handler: fnHomeHandler,
	}

	deleteTransactionHandler := &postHandler{
		info:    info,
		handler: fnDeleteTransactionHandler,
	}

	accountsHandler := &postHandler{
		info:    info,
		handler: fnAccountsHandler,
	}

	addTransactionHandler := &postHandler{
		info:    info,
		handler: fnAddTransactionHandler,
	}

	addAccountHandler := &postHandler{
		info:    info,
		handler: fnAddAccountHandler,
	}

	recurringsPageHandler := &postHandler{
		info:    info,
		handler: fnRecurringsPageHandler,
	}

	listRecurringsHandler := &postHandler{
		info:    info,
		handler: fnRecurringsHandler,
	}

	addRecurringHandler := &postHandler{
		info:    info,
		handler: fnAddRecurringHandler,
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", homeHandler)

	http.Handle("/accounts", accountsHandler)
	http.Handle("/addAccount", addAccountHandler)

	http.Handle("/getTransactions", listTransactionsHandler)
	http.Handle("/addTransaction", addTransactionHandler)
	http.Handle("/deleteTransaction", deleteTransactionHandler)

	http.Handle("/recurrings", recurringsPageHandler)
	http.Handle("/getRecurrings", listRecurringsHandler)
	http.Handle("/addRecurring", addRecurringHandler)

	fmt.Printf("Running Server..\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
