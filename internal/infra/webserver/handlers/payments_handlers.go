package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/queue"
)

type PaymentsHandler struct {
	tasksQueue     *queue.TasksQueue
	transactionsDB *database.Transactions
}

func NewPaymentsHandler(tq *queue.TasksQueue, db *database.Transactions) *PaymentsHandler {
	return &PaymentsHandler{
		tasksQueue:     tq,
		transactionsDB: db,
	}
}

type paymentsRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func (h *PaymentsHandler) ProcessorPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req paymentsRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task := queue.TaskInfo{
		ID:          req.CorrelationID,
		Amount:      req.Amount,
		RequestedAt: time.Now().UTC(),
	}

	err = h.tasksQueue.Enqueue(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type paymentsSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type paymentsSummaryResponse struct {
	DefaultProcessor  paymentsSummary `json:"default"`
	FallbackProcessor paymentsSummary `json:"fallback"`
}

func (h *PaymentsHandler) RequestSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	from := query.Get("from")
	to := query.Get("to")

	var (
		err     error
		summary paymentsSummaryResponse
		result  database.TransactionsSummary
	)

	if (from != "") && (to != "") {
		layout := time.RFC3339
		timeFrom, errFrom := time.Parse(layout, from)
		timeTo, errTo := time.Parse(layout, to)
		if errFrom != nil || errTo != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err = h.transactionsDB.ReadTransactionsOnPeriod(timeFrom, timeTo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		result, err = h.transactionsDB.ReadAllTransactions()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	summary.DefaultProcessor = paymentsSummary(result.DefaultProcessor)
	summary.FallbackProcessor = paymentsSummary(result.FallbackProcessor)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
