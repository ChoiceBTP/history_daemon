package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type HistoryRow struct {
	id            int
	url           string
	title         string
	visitCount    int
	typedCount    int
	lastVisitTime int
	hidden        int
	m             map[string]string // url to classification
}

func HistoryRoutine() {
	const file string = "/Users/tanmaygairola/Library/Application Support/Google/Chrome/Default/History"
	db, err := sql.Open("sqlite3", file)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)
	if err != nil {
		log.Println(err)
	}
	for {
		query := "select * from urls order by last_visit_time desc limit 10"
		rows, _ := db.Query(query)

		history := make([]HistoryRow, 0)

		for rows.Next() {
			historyRow := HistoryRow{}
			err := rows.Scan(&historyRow.id,
				&historyRow.url,
				&historyRow.title,
				&historyRow.visitCount,
				&historyRow.typedCount,
				&historyRow.lastVisitTime,
				&historyRow.hidden)
			if err != nil {
				log.Fatal(err)
			}
			history = append(history, historyRow)

		}
		for _, historyRow := range history {
			fmt.Printf("title: %v, url: %v\n", historyRow.title, historyRow.url)
		}
		time.Sleep(30 * time.Minute)
	}
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/getHistory", getHistory)
	mux.HandleFunc("/ping", pong)

	http.ListenAndServe(":3333", mux)
}

func pong(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "pong\n")
}

func getHistory(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "sent default history for "+request.URL.Query().Get("hour"))
}

func main() {
	go HistoryRoutine()
	go Server()
	time.Sleep(1000 * time.Minute)
}
