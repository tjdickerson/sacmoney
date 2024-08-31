package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	db "tjdickerson/sacmoney/pkg/database"
	utils "tjdickerson/sacmoney/pkg/utils"
)

func handleNoAccount(w http.ResponseWriter, t *template.Template) {
	data := TransMain{
		AccountName:    "No account, click on accounts at top.",
		TotalAvailable: "0",
		Transactions:   nil,
		Error:          "",
	}

	var outHtml bytes.Buffer
	t.Execute(&outHtml, data)
	io.WriteString(w, outHtml.String())
}

type TransactionData struct {
	Id     string
	Date   string
	Name   string
	Amount string
	IsNeg  bool
}

type RecurringDisplay struct {
	Id       string
	Name     string
	Day      string
	Amount   string
	CssClass string
	IsNeg    bool
}

type TransMain struct {
	AccountName    string
	AvailClass     string
	Month          string
	Year           string
	NextMonth      string
	NextYear       string
	TotalAvailable string
	Transactions   []TransactionData
	Recurrings     []RecurringDisplay
	Error          string
}

func convertTransaction(t *db.Transaction) TransactionData {
	return TransactionData{
		Id:     strconv.Itoa(t.Id),
		Name:   html.EscapeString(strings.TrimSpace(t.Name)),
		Date:   t.Date.Format("Mon 02 Jan"),
		Amount: fmt.Sprintf("%.2f", float32(t.Amount)*float32(0.01)),
		IsNeg:  t.Amount < 0,
	}
}

func (t *TransactionData) toDbTransaction() (db.Transaction, error) {
	name := html.EscapeString(strings.TrimSpace(t.Name))
	amount := utils.GetCentsFromString(t.Amount)

	var outErr string = ""
	id, err := strconv.Atoi(t.Id)
	if err != nil {
		outErr = outErr + "Error reading id. "
	}

	date, err := time.Parse("2006-01-02", t.Date)
	if err != nil {
		date = time.Now()
	}

	if len(name) == 0 {
		outErr = outErr + "Name required. "
	}

	if len(outErr) > 0 {
		return db.Transaction{}, fmt.Errorf("%s", outErr)
	}

	return db.Transaction{
		Id:     id,
		Name:   name,
		Amount: amount,
		Date:   date,
	}, nil
}

func TransMainHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"templates/transactions/trans_main_tmpl.html",
		"templates/core/title_tmpl.html")

	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing template: %s", err))
	}

	if servctx.currentAccount == nil {
		handleNoAccount(w, t)
		return
	}

	outError := ""
	accountName := servctx.currentAccount.Name
	totalAvailable := fmt.Sprintf("%.2f", float32(servctx.currentAccount.TotalAvailable)*float32(0.01))
	transactions, err := db.FetchAllTransactions()
	if err != nil {
		outError = fmt.Sprintf("%s", err)
		log.Println(outError)
	}

	transactionData := []TransactionData{}
	for _, dbTrans := range transactions {
		transactionData = append(transactionData, convertTransaction(&dbTrans))
	}

	recurrings, err := db.FetchAllRecurrings()
	if err != nil {
		outError = fmt.Sprintf("%s", err)
		log.Println(outError)
	}

	recurringData := []RecurringDisplay{}
	for _, r := range recurrings {
		exists := transactionsContainsRecurring(&transactionData, r.Name)
		var cssClass string
		if exists {
			cssClass = "accounted-for"
		} else {
			cssClass = ""
		}

		recurringData = append(recurringData, RecurringDisplay{
			Id:       fmt.Sprintf("%d", r.Id),
			Name:     r.Name,
			Amount:   fmt.Sprintf("%.2f", float32(r.Amount)*float32(0.01)),
			IsNeg:    r.Amount < 0,
			Day:      fmt.Sprintf("%d", r.Day),
			CssClass: cssClass,
		})
	}

	currentYear := servctx.currentYear
	currentMonth := servctx.currentMonth
	nextYear, nextMonth, err := GetNextYearMonth(currentYear, currentMonth)
	if err != nil {
		outError = fmt.Sprintf("%s", err)
		log.Printf("%s\n", outError)
	}

	availClass := "pos"
	if servctx.currentAccount.TotalAvailable < 0 {
		availClass = "neg"
	}

	data := TransMain{
		AccountName:    accountName,
		Month:          currentMonth,
		Year:           currentYear,
		NextMonth:      nextMonth,
		NextYear:       nextYear,
		TotalAvailable: totalAvailable,
		Transactions:   transactionData,
		Recurrings:     recurringData,
		AvailClass:     availClass,
		Error:          outError,
	}

	var outHtml bytes.Buffer
	t.Execute(&outHtml, data)
	io.WriteString(w, outHtml.String())
}

func SaveTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var data TransactionData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		outErr := fmt.Sprintf("Failed to decode transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	transaction, err := data.toDbTransaction()
	if err != nil {
		outErr := fmt.Sprintf("%s", err)
		io.WriteString(w, outErr)
		return
	}

	if transaction.Id == 0 {
		err = db.Insert(&transaction)
	} else {
		err = db.Update(&transaction)
	}

	if err != nil {
		outErr := fmt.Sprintf("Failed to add transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	RefreshAccount()
	io.WriteString(w, "SUCCESS")
}

func DeleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var data TransactionData
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		outErr := fmt.Sprintf("Failed to decode transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	id, err := strconv.Atoi(data.Id)
	if err != nil {
		outErr := fmt.Sprintf("Failed to convert transaction id: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	temp := db.Transaction{Id: id}
	err = db.Delete(&temp)
	if err != nil {
		outErr := fmt.Sprintf("Error deleting transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	RefreshAccount()
	io.WriteString(w, "SUCCESS")
}

func ApplyRecurringHandler(w http.ResponseWriter, r *http.Request) {
	type jsonData struct {
		Id string
	}
	jd := jsonData{}
	if err := json.NewDecoder(r.Body).Decode(&jd); err != nil {
		outErr := fmt.Sprintf("Error getting recurring transaction id: %s", err)
		log.Printf("%s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	id, err := strconv.Atoi(jd.Id)
	if err != nil {
		outErr := fmt.Sprintf("Error converting recurring transaction id: %s", err)
		log.Printf("%s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	db.CreateTransactionFromRecurring(id)

	RefreshAccount()
	io.WriteString(w, "SUCCESS")
}

func transactionsContainsRecurring(transactions *[]TransactionData, recurringName string) bool {
	for _, t := range *transactions {
		if strings.ToLower(strings.TrimSpace(t.Name)) == strings.ToLower(strings.TrimSpace(recurringName)) {
			return true
		}
	}

	return false
}
