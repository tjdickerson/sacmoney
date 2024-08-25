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
	db "tjdickerson/sacmoney/pkg/database"
	utils "tjdickerson/sacmoney/pkg/utils"
)

type RecurringData struct {
	Id     string
	Day    string
	Name   string
	Amount string
	IsNeg  bool
}

type RecurringMain struct {
	AccountName           string
	RecurringTransactions []RecurringData
	Error                 string
}

func convertRecurring(r *db.Recurring) RecurringData {
	return RecurringData{
		Id:     strconv.Itoa(r.Id),
		Name:   r.Name,
		Day:    strconv.Itoa(int(r.Day)),
		Amount: fmt.Sprintf("%.2f", float32(r.Amount)*float32(0.01)),
		IsNeg:  r.Amount < 0,
	}
}

func (r *RecurringData) toDbRecurring() (db.Recurring, error) {
	name := html.EscapeString(strings.TrimSpace(r.Name))
	amount := utils.GetCentsFromString(r.Amount)
	day, err := strconv.Atoi(r.Day)

	if err != nil {
		return db.Recurring{}, fmt.Errorf("Failed to parse Day: %s", err)
	}

	return db.Recurring{
		Name:   name,
		Amount: amount,
		Day:    uint8(day),
	}, nil
}

func RecurringMainHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"templates/recurrings/recurr_main_tmpl.html",
		"templates/core/title_tmpl.html")

	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing template: %s", err))
	}

	if servctx == nil {
		handleNoAccount(w)
		return
	}

	outError := ""
	accountName := servctx.currentAccount.Name
	recurrings, err := db.FetchAllRecurrings()
	if err != nil {
		outError = fmt.Sprintf("%s", err)
		log.Println(outError)
	}

	recurringData := []RecurringData{}
	for _, dbRecurr := range recurrings {
		recurringData = append(recurringData, convertRecurring(&dbRecurr))
	}

	data := RecurringMain{
		AccountName:           accountName,
		RecurringTransactions: recurringData,
		Error:                 outError,
	}

	var outHtml bytes.Buffer
	t.Execute(&outHtml, data)
	io.WriteString(w, outHtml.String())
}

func AddRecurringHandler(w http.ResponseWriter, r *http.Request) {
	var data RecurringData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		outErr := fmt.Sprintf("Failed to decode recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	recurring, err := data.toDbRecurring()
	if err != nil {
		outErr := fmt.Sprintf("Failed to add recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
	}

	err = db.Insert(&recurring)
	if err != nil {
		outErr := fmt.Sprintf("Failed to add recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
	}

	io.WriteString(w, "SUCCESS")
}
