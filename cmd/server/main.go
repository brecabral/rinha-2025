package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/handle"
	"github.com/brecabral/rinha-2025/internal/store"
)

func main() {
	processor := decision.Decider()
	connStr := "host=db user=rinha password=rinha dbname=rinha sslmode=disable"
	time.Sleep(5 * time.Second)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ERROR] Não foi possivel se conectar ao banco de dados: %v", err)
	}
	dbClient, err := store.NewDatabase(db)
	if err != nil {
		log.Fatalf("[ERROR] Não foi possivel criar uma conexão com o banco de dados: %v", err)
	}
	defer dbClient.Close()
	h := handle.Handler{
		Processor:      &processor,
		DatabaseClient: dbClient,
	}
	http.HandleFunc("/payments", h.PaymentsHandler)
	http.HandleFunc("/payments-summary", h.PaymentsSummaryHandler)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
