package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	db "tjdickerson/sacmoney/pkg/database"
)

const DbDirectory = "data/"

type serverContext struct {
	currentAccount *db.Account
	currentMonth   string
	currentYear    string
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

func checkEnvironment() error {
	if _, err := os.Stat(DbDirectory); os.IsNotExist(err) {
		err = os.Mkdir(DbDirectory, 0700)
		if err != nil {
			return fmt.Errorf("Error getting data directory status: %s", err)
		}
	}

	return nil
}

func getTargetDbName() string {
	now := time.Now()

	entries, err := os.ReadDir(DbDirectory)
	if err != nil {
		log.Printf("Error while reading directory contents: %s\n", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("%s.db", now.Format("2006January"))
	}

	var databases = make([]string, 0, 100)
	for _, e := range entries {
		if !e.IsDir() {
			databases = append(databases, e.Name())
		}
	}

	sort.Slice(databases, func(i, j int) bool {
		return i < j
	})

	return databases[len(databases)-1]
}

func GetNextYearMonth(year string, month string) (string, string, error) {
	const defaultYear = "1970"
	const defaultMonth = "January"
	t, err := time.Parse("2006January", fmt.Sprintf("%s%s", year, month))
	if err != nil {
		return defaultYear, defaultMonth, fmt.Errorf("Failed to convert %s to an actual month.", month)
	}

	var nextMonth time.Month
	var nextYear int

	currentMonth := t.Month()
	currentYear, err := strconv.Atoi(year)
	if err != nil {
		return defaultYear, defaultMonth, fmt.Errorf("Failed to convert year %s to an actual year.", year)
	}

	if currentMonth == time.December {
		nextMonth = time.January
		nextYear = currentYear + 1
	} else {
		nextMonth = currentMonth + 1
		nextYear = currentYear
	}

	return strconv.Itoa(nextYear), nextMonth.String(), nil
}

func NextMonthRollover(w http.ResponseWriter, r *http.Request) {
	year, month, err := GetNextYearMonth(servctx.currentYear, servctx.currentMonth)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("Error getting rollover date: %s", err))
	}

	servctx.currentMonth = month
	servctx.currentYear = year

	newDbPath := fmt.Sprintf("%s/%s%s.db", DbDirectory, year, month)
	if err = db.InitDatabase(newDbPath, true); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to new database instance: %s\n", err))
	}

	RefreshAccount()
	io.WriteString(w, "SUCCESS")
}

func Run() {
	checkEnvironment()
	dbName := getTargetDbName()
	dbPath := fmt.Sprintf("%s/%s", DbDirectory, dbName)
	db.InitDatabase(dbPath, false)
	defer db.CloseDatabase()

	temp := strings.TrimRight(dbName, ".db")
	month := temp[4:]
	year := dbName[:4]

	servctx = &serverContext{}
	servctx.currentMonth = month
	servctx.currentYear = year

	if !db.HasAccount() {
		servctx.currentAccount = nil
	} else {
		RefreshAccount()
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", TransMainHandler)
	http.HandleFunc("/saveTransaction", SaveTransactionHandler)
	http.HandleFunc("/deleteTransaction", DeleteTransactionHandler)

	http.HandleFunc("/recurrings", RecurringMainHandler)
	http.HandleFunc("/saveRecurring", SaveRecurringHandler)
	http.HandleFunc("/deleteRecurring", DeleteRecurringHandler)

	http.HandleFunc("/accounts", AccountMainHandler)
	http.HandleFunc("/addAccount", AddAccountHandler)

	http.HandleFunc("/rollover", NextMonthRollover)
	http.HandleFunc("/applyRecurring", ApplyRecurringHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
