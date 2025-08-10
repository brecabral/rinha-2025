package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/webserver/handlers"
)

func main() {
	connStr := "host=db user=rinha password=rinha dbname=rinha sslmode=disable"
	time.Sleep(5 * time.Second)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ERROR] Não foi possivel se conectar ao banco de dados: %v", err)
	}
	dbClient, err := database.NewDatabase(db)
	if err != nil {
		log.Fatalf("[ERROR] Não foi possivel criar uma conexão com o banco de dados: %v", err)
	}
	defer dbClient.Close()
	decider := decision.NewDecider()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			decider.CheckProcessors()
			decider.Chose()
		}
	}()
	h := handlers.NewPaymentsHandler(
		decider,
		dbClient,
	)
	http.HandleFunc("/payments", h.ProcessorPayment)
	http.HandleFunc("/payments-summary", h.RequestSummary)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
