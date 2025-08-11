package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/infra/cache"
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
		&http.Client{
			Timeout: 300 * time.Millisecond,
		},
		urlDefaultProcessor,
		true,
	)

	const urlFallbackProcessor = "http://payment-processor-fallback:8080"
	var fallbackProcessorClient = clients.NewProcessorClient(
		&http.Client{
			Timeout: 200 * time.Millisecond,
		},
		urlFallbackProcessor,
		false,
	)

	processorCache := cache.NewProcessorsCache()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := processorCache.RedisClient.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("[ERROR] redis não responde")
	}

	processorDecider := decision.NewDecider(defaultProcessorClient, fallbackProcessorClient, processorCache)
	go processorDecider.StartHealthCheck()
	return processorDecider
}

func CreateDBConection() *database.Transactions {
	const connStr = "host=db user=rinha password=rinha dbname=rinha sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ERROR] drive do banco de dados incompatível: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("[ERROR] banco de dados não responde")
	}

	dbClient, err := database.NewTransactionsDB(db)
	if err != nil {
		log.Fatalf("[ERROR] não foi possivel criar o db: %v", err)
	}

	return dbClient
}

func CreateWorkerPool() *workers.WorkerPool {
	return workers.NewWorkerPool(2)
}
