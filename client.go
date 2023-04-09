package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n, err")
	}
	defer file.Close()
	cotacao, error := BuscaCotacao()
	TreatError(error)

	fmt.Println(cotacao.Bid)
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", cotacao.Bid))

}

func BuscaCotacao() (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	TreatError(err)

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	TreatError(err)
	defer res.Body.Close()
	res2, err := io.ReadAll(res.Body)
	TreatError(err)

	var cotacao Cotacao
	err = json.Unmarshal(res2, &cotacao)
	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}

func TreatError(err error) {
	if err != nil {
		panic(err)
	}
}
