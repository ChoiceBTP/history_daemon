package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UrlListReq struct {
	UrlList []string `json:"url_list"`
}

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

const url string = "http://127.0.0.1:5000/process_urls"

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
		url_list := make([]string, 0)
		for _, historyRow := range history {
			url_list = append(url_list, historyRow.url)
		}
		data, _ := json.Marshal(UrlListReq{UrlList: url_list})
		dataReader := bytes.NewReader(data)
		fmt.Println(url_list)
		req, err := http.NewRequest(http.MethodPost, url, dataReader)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{
			Timeout: 30 * time.Second,
		}

		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("client: error making http request: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(res)
		time.Sleep(2 * time.Second)
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
