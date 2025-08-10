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
	ProcessorDecider *decision.Decider
	PaymentsWP       *workers.WorkerPool
	PaymentsDB       *database.Database
}

func NewPaymentsHandler(decider *decision.Decider, wp *workers.WorkerPool, db *database.Database) *PaymentsHandler {
	return &PaymentsHandler{
		ProcessorDecider: decider,
		PaymentsWP:       wp,
		PaymentsDB:       db,
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

	var req dto.PaymentRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := dto.ProcessorPaymentRequest{
		CorrelationID: req.CorrelationID,
		Amount:        req.Amount,
		RequestedAt:   time.Now(),
	}

	task := workers.NewPaymentTask(data, h.ProcessorDecider, h.PaymentsDB)
	h.PaymentsWP.Submit(task)

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

	var summary *dto.PaymentsSummaryResponse
	var err error

	if (from != "") && (to != "") {
		layout := time.RFC3339
		timeFrom, err := time.Parse(layout, from)
		if err != nil {
			return
		}
		timeTo, err := time.Parse(layout, to)
		if err != nil {
			return
		}
		summary, err = h.PaymentsDB.ReadTransactionsOnPeriod(timeFrom, timeTo)
		if err != nil {
			return
		}
	} else {
		summary, err = h.PaymentsDB.ReadAllTransactions()
		if err != nil {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
