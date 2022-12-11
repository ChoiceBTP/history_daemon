package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type ClassificationResponse struct {
	url   string
	class string
}

const post_url string = "http://127.0.0.1:5000/process_urls"
const get_classes_url string = "http://127.0.0.1:5000/get_url_classes"

const file string = "C:/Users/jhasa/Desktop/History.sqlite"

func HistoryRoutine() {
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
		query := "select * from urls order by last_visit_time desc limit 50"
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
		// fmt.Println(url_list)
		req, err := http.NewRequest(http.MethodPost, post_url, dataReader)
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
		time.Sleep(100000 * time.Second)
		//20
	}
}

func Server() {
	mux := http.NewServeMux()
	// localhost:3333/getHistory?time=20323
	mux.HandleFunc("/getHistory", getHistory)
	mux.HandleFunc("/ping", pong)
	http.ListenAndServe(":3333", mux)
}

func pong(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "pong\n")
}

func getHistory(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("client side server req")
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
	if err != nil {
		io.WriteString(writer, err.Error())
	}
	query := "select * from urls where last_visit_time >" + request.URL.Query().Get("time")
	fmt.Println(query)
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
			&historyRow.hidden,
		)
		if err != nil {
			log.Fatal(err)
		}
		history = append(history, historyRow)
	}
	//fmt.Println(history)
	url_list := make([]string, 0)
	for _, historyRow := range history {
		url_list = append(url_list, historyRow.url)
	}
	data, _ := json.Marshal(UrlListReq{UrlList: url_list})
	dataReader := bytes.NewReader(data)
	fmt.Println(url_list)
	req, err := http.NewRequest(http.MethodPost, get_classes_url, dataReader)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 3000 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(res.Body)
	//classificationResponse := ClassificationResponse{}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

    writer.Header().Set("Content-Type", "application/json")
    writer.Write(body)
}

func main() {
	go HistoryRoutine()
	go Server()
	time.Sleep(1000 * time.Minute)
}
