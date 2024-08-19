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
	setup "sacdev/sacmoney/pkg/setup"
	trn "sacdev/sacmoney/pkg/transactions"
	utils "sacdev/sacmoney/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type PageData struct {
	Title       string
	AccountName string
	Available   string
}

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

type TransactionData struct {
	Id              string
	TransactionDate string
	Name            string
	AmountClass     string
	Amount          string
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
	name := strings.TrimSpace(ntj.Name)
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

func fnHomeHandler(h *postHandler, r *http.Request) string {
	available, _ := trn.GetAvailable(h.info.db, h.info.accountId)
	tmpl, _ := template.ParseFiles("templates/home.html")

	p := &PageData{
		Title:       "sacmoney",
		AccountName: html.EscapeString(h.info.accountName),
		Available:   fmt.Sprintf("%.2f", available),
	}

	var htmlTmpl bytes.Buffer
	tmpl.Execute(&htmlTmpl, p)
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

	addTransactionHandler := &postHandler{
		info:    info,
		handler: fnAddTransactionHandler,
	}

	deleteTransactionHandler := &postHandler{
		info:    info,
		handler: fnDeleteTransactionHandler,
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", homeHandler)
	http.Handle("/getTransactions", listTransactionsHandler)
	http.Handle("/addTransaction", addTransactionHandler)
	http.Handle("/deleteTransaction", deleteTransactionHandler)

	fmt.Printf("Running Server..\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
