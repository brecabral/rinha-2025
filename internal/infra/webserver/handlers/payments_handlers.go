package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/dto"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/workers"
)

type PaymentsHandler struct {
	TasksWP          *workers.WorkerPool
	TransactionsDB   *database.Transactions
	ProcessorDecider *decision.Decider
}

func NewPaymentsHandler(wp *workers.WorkerPool, db *database.Transactions, decider *decision.Decider) *PaymentsHandler {
	return &PaymentsHandler{
		TasksWP:          wp,
		TransactionsDB:   db,
		ProcessorDecider: decider,
	}
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

	var req dto.PaymentsRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := workers.TaskInfo{
		ID:          req.CorrelationID,
		Amount:      req.Amount,
		RequestedAt: time.Now(),
	}
	task := workers.NewPaymentTask(data, h.ProcessorDecider, h.TransactionsDB)
	h.TasksWP.Submit(task)

	w.WriteHeader(http.StatusOK)
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
		summary dto.PaymentsSummaryResponse
		result  dto.DatabaseReadTransactions
	)

	if (from != "") && (to != "") {
		layout := time.RFC3339
		timeFrom, errFrom := time.Parse(layout, from)
		timeTo, errTo := time.Parse(layout, to)
		if errFrom != nil || errTo != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err = h.TransactionsDB.ReadTransactionsOnPeriod(timeFrom, timeTo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		result, err = h.TransactionsDB.ReadAllTransactions()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	summary.DefaultProcessor = dto.PaymentsSummary(result.DefaultProcessor)
	summary.FallbackProcessor = dto.PaymentsSummary(result.FallbackProcessor)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
