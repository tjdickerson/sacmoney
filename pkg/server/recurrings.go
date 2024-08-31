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
	Net                   string
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

	var outErr string = ""
	id, err := strconv.Atoi(r.Id)
	if err != nil {
		outErr = outErr + "Error reading id. "
	}

	day, err := strconv.Atoi(r.Day)
	if err != nil {
		outErr = outErr + "Day of occurrence required. "
	} else {
		if day < 1 || day > 28 {
			outErr = outErr + "Day needs to be between 1-28 inclusive. "
		}
	}

	if len(name) == 0 {
		outErr = outErr + "Name required. "
	}

	if len(outErr) > 0 {
		return db.Recurring{}, fmt.Errorf("%s", outErr)
	}

	return db.Recurring{
		Id:     id,
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

	if servctx.currentAccount == nil {
		handleNoAccount(w, t)
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

	net, err := db.GetRecurringNetBalance()
	if err != nil {
		outError = fmt.Sprintf("%s<br />%s", outError, err)
		net = 0
	}

	data := RecurringMain{
		AccountName:           accountName,
		RecurringTransactions: recurringData,
		Net:                   fmt.Sprintf("%.2f", float32(net)*float32(0.01)),
		Error:                 outError,
	}

	var outHtml bytes.Buffer
	t.Execute(&outHtml, data)
	io.WriteString(w, outHtml.String())
}

func SaveRecurringHandler(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	if recurring.Id == 0 {
		err = db.Insert(&recurring)
	} else {
		err = db.Update(&recurring)
	}

	if err != nil {
		outErr := fmt.Sprintf("Failed to add recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	io.WriteString(w, "SUCCESS")
}

func DeleteRecurringHandler(w http.ResponseWriter, r *http.Request) {
	var data RecurringData
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		outErr := fmt.Sprintf("Failed to decode recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	id, err := strconv.Atoi(data.Id)
	if err != nil {
		outErr := fmt.Sprintf("Failed to convert recurring transaction id: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	temp := db.Recurring{Id: id}
	err = db.Delete(&temp)
	if err != nil {
		outErr := fmt.Sprintf("Error deleting recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	RefreshAccount()
	io.WriteString(w, "SUCCESS")
}
