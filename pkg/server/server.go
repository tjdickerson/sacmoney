package server

import (
	"fmt"
	_ "html"
	_ "html/template"
	_ "io"
	"log"
	"net/http"
	db "tjdickerson/sacmoney/pkg/database"
)

type serverContext struct {
	currentAccount *db.Account
}

var (
	servctx *serverContext
)

func RefreshAccount() {
	account, err := db.GetDefaultAccount()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting account: %s\n", err))
	}
	servctx.currentAccount = &account
}

func Run() {
	db.InitDatabase()
	defer db.CloseDatabase()

	servctx = &serverContext{}
	if !db.HasAccount() {
		servctx = nil
	} else {
		RefreshAccount()
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", TransMainHandler)
	http.HandleFunc("/addTransaction", AddTransactionHandler)

	http.HandleFunc("/recurrings", RecurringMainHandler)
	http.HandleFunc("/addRecurring", AddRecurringHandler)

	http.HandleFunc("/accounts", AccountMainHandler)
	http.HandleFunc("/addAccount", AddAccountHandler)

	//
	// http.Handle("/getTransactions", listTransactionsHandler)
	// http.Handle("/deleteTransaction", deleteTransactionHandler)
	//
	// http.Handle("/recurrings", recurringsPageHandler)
	// http.Handle("/getRecurrings", listRecurringsHandler)
	// http.Handle("/addRecurring", addRecurringHandler)
	//
	// fmt.Printf("Running Server..\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
