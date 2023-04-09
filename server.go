package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type USDBRL struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {

	http.HandleFunc("/cotacao", Handler)
	http.ListenAndServe(":8080", nil)

}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("Request")
	defer log.Println("Request finalizada")
	select {
	case <-time.After(5 * time.Second):
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		TreatError(err)

		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		TreatError(err)
		defer res.Body.Close()
		res2, err := io.ReadAll(res.Body)
		TreatError(err)

		var data USDBRL
		err = json.Unmarshal(res2, &data)
		TreatError(err)

		db, err := sql.Open("sqlite3", "banco.db")
		TreatError(err)
		defer db.Close()

		CreateTable(db)
		err = InsertCotacao(db, &data, r.Context())
		TreatError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		var cotacao Cotacao
		cotacao.Bid = data.Usdbrl.Bid
		json.NewEncoder(w).Encode(cotacao)

	case <-ctx.Done():
		log.Println("Request cancelada")
	}

}

func CreateTable(db *sql.DB) {
	createCotacaoTableSQL := `CREATE TABLE IF NOT EXISTS cotacao (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"cotacao" TEXT		
	  );` // SQL Statement for Create Table

	log.Println("Create cotacao table...")
	statement, err := db.Prepare(createCotacaoTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("cotacao table created if not exixts")
}

func InsertCotacao(db *sql.DB, cotacao *USDBRL, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("insert into cotacao(cotacao) values ($1)")
	TreatError(err)
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao.Usdbrl.Bid)
	TreatError(err)
	return nil
}

func TreatError(err error) {
	if err != nil {
		panic(err)
	}
}

/*
{"USDBRL":{"code":"USD","codein":"BRL","name":"Dólar Americano/Real Brasileiro","high":"5.0599","low":"5.0559","varBid":"0","pctChange":"0","bid":"5.0565","ask":"5.0575","timestamp":"1680878735","create_date":"2023-04-07 11:45:35"}}(0x102ddbf30,0x140002053e0)
*/
