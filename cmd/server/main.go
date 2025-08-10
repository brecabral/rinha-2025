package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/infra/clients"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/webserver/handlers"
	"github.com/brecabral/rinha-2025/internal/infra/workers"
)

func main() {
	processorDecider := CreateProcessorDecider()
	wokerPool := CreateWorkerPool()
	transactionsDB := CreateDBConection()
	defer transactionsDB.Close()

	paymentsHandler := handlers.NewPaymentsHandler(
		wokerPool,
		transactionsDB,
		processorDecider,
	)

	http.HandleFunc("/payments", paymentsHandler.ProcessorPayment)
	http.HandleFunc("/payments-summary", paymentsHandler.RequestSummary)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CreateProcessorDecider() *decision.Decider {
	const urlDefaultProcessor = "http://payment-processor-default:8080"
	var defaultProcessorClient = clients.NewProcessorClient(
		&http.Client{},
		urlDefaultProcessor,
		true,
	)

	const urlFallbackProcessor = "http://payment-processor-fallback:8080"
	var fallbackProcessorClient = clients.NewProcessorClient(
		&http.Client{},
		urlFallbackProcessor,
		false,
	)

	processorDecider := decision.NewDecider(defaultProcessorClient, fallbackProcessorClient)
	go processorDecider.StartHealthCheck()
	return processorDecider
}

func CreateDBConection() *database.Transactions {
	const connStr = "host=db user=rinha password=rinha dbname=rinha sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ERROR] drive do banco de dados incompatível: %v", err)
	}

	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			log.Fatalf("[ERROR] banco de dados não responde")
		}
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	dbClient, err := database.NewTransactionsDB(db)
	if err != nil {
		log.Fatalf("[ERROR] a conexão com o banco de dados não foi estabelecida: %v", err)
	}

	return dbClient
}

func CreateWorkerPool() *workers.WorkerPool {
	return workers.NewWorkerPool(100)
}
