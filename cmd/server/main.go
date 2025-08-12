package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/brecabral/rinha-2025/internal/infra/clients"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/queue"
	"github.com/brecabral/rinha-2025/internal/infra/webserver/handlers"
	"github.com/brecabral/rinha-2025/internal/infra/workers"
)

func main() {
	tasksQueue := CreateTasksQueue()
	transactionsDB := CreateDBConection()
	defer transactionsDB.Close()

	workersCountStr := os.Getenv("WORKERS_COUNT")
	workersCount, err := strconv.Atoi(workersCountStr)
	if err != nil {
		workersCount = 10
	}

	workerPool := CreateWorkerPool(tasksQueue, transactionsDB, workersCount)
	workerPool.Start()
	defer workerPool.Stop()

	paymentsHandler := handlers.NewPaymentsHandler(
		tasksQueue,
		transactionsDB,
	)

	http.HandleFunc("/payments", paymentsHandler.ProcessorPayment)
	http.HandleFunc("/payments-summary", paymentsHandler.RequestSummary)

	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func CreateTasksQueue() *queue.TasksQueue {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("[ERROR] redis não responde")
	}
	return queue.NewTasksQueue(redisClient)
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

func CreateWorkerPool(tq *queue.TasksQueue, db *database.Transactions, w int) *workers.WorkerPool {
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
	return workers.NewWorkerPool(tq, db, defaultProcessorClient, fallbackProcessorClient, w)
}
