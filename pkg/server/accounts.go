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
)

type AccountData struct {
	Id   string
	Name string
}

type AccountMain struct {
	CurrentAccount string
	Accounts       []AccountData
	Error          string
}

func convertAccount(a *db.Account) AccountData {
	return AccountData{
		Id:   strconv.Itoa(a.Id),
		Name: a.Name,
	}
}

func (r *AccountData) toDbAccount() db.Account {
	name := html.EscapeString(strings.TrimSpace(r.Name))
	return db.Account{
		Name: name,
	}
}

func AccountMainHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"templates/accounts/accounts_main_tmpl.html",
		"templates/core/title_tmpl.html")

	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing template: %s", err))
	}

	outError := ""
	accounts, err := db.FetchAllAccounts()
	if err != nil {
		outError = fmt.Sprintf("%s", err)
		log.Println(outError)
	}

	accountData := []AccountData{}
	for _, dbAccount := range accounts {
		accountData = append(accountData, convertAccount(&dbAccount))
	}

	data := AccountMain{
		CurrentAccount: servctx.currentAccount.Name,
		Accounts:       accountData,
		Error:          outError,
	}

	var outHtml bytes.Buffer
	t.Execute(&outHtml, data)
	io.WriteString(w, outHtml.String())
}

func AddAccountHandler(w http.ResponseWriter, r *http.Request) {
	var data AccountData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		outErr := fmt.Sprintf("Failed to decode recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
		return
	}

	recurring := data.toDbAccount()

	err = db.Insert(&recurring)
	if err != nil {
		outErr := fmt.Sprintf("Failed to add recurring transaction: %s", err)
		log.Printf("Error: %s\n", outErr)
		io.WriteString(w, outErr)
	}

	if servctx == nil {
		RefreshAccount()
	}

	io.WriteString(w, "SUCCESS")
}
