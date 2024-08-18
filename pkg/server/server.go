package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	setup "sacdev/sacmoney/pkg/setup"
	trn "sacdev/sacmoney/pkg/transactions"
	"time"
)

type PageData struct {
	Title string
}

type postHandler struct {
	db        *sql.DB
	accountId int32
	handler   func(h *postHandler) string
}

func (h *postHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html := h.handler(h)
	io.WriteString(w, html)
}

func getTransactionsHandler(h *postHandler) string {
	transactions, err := trn.GetLastTransactions(h.db, h.accountId)

	if err != nil {
		return fmt.Sprintf("<div class=\"error\">%s</div>", err)
	}

	html := "<div class=\"transactions\">"
	for _, t := range transactions {
		strDate := time.UnixMilli(t.Date).Format("02 Mon")
		amount := fmt.Sprintf("$%.2f", float64(t.Amount)*float64(0.01))
		amountClass := "amount"
		if t.Amount > 0 {
			amountClass = "amount pos"
		} else if t.Amount < 0 {
			amountClass = "amount neg"
		}

		html += "<div class=\"transaction\">"
		html += "<div class=\"date\">" + strDate + "</div>"
		html += "<div class=\"name\">" + t.Name + "</div>"
		html += fmt.Sprintf("<div class=\"%s\">%s</div>", amountClass, amount)
		html += "</div>"
	}
	html += "</div>"

	return html
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed: %s", err))
	}

	p := &PageData{
		Title: "sacmoney",
	}

	t.Execute(w, p)
}

func Run() {
	var err error
	db, err := setup.GetDatabase()
	if err != nil {
		log.Fatal(err)
	}
	acct, _ := setup.GetDefaultAccount(db)
	getTransHandler := &postHandler{
		db:        db,
		accountId: acct,
		handler:   getTransactionsHandler,
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", handler)
	http.Handle("/getTransactions", getTransHandler)

	fmt.Printf("Running Server..\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
