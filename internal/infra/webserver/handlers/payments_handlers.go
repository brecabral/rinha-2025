package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/dto"
	"github.com/brecabral/rinha-2025/internal/infra/database"
)

type PaymentsHandler struct {
	ProcessorDecider *decision.Decider
	PaymentsDB       *database.Database
}

func NewPaymentsHandler(processorDecider *decision.Decider, db *database.Database) *PaymentsHandler {
	return &PaymentsHandler{
		ProcessorDecider: processorDecider,
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

	var data dto.PaymentRequest
	if err := json.Unmarshal(requestBody, &data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	body := dto.ProcessorPaymentRequest{
		CorrelationID: data.CorrelationID,
		Amount:        data.Amount,
		RequestedAt:   time.Now(),
	}

	if err := h.ProcessorDecider.Processor.PostPayment(ctx, body); err != nil {
		h.ProcessorDecider.Processor.Failing = true
		h.ProcessorDecider.Chose()
		return
	}

	h.PaymentsDB.SaveTransaction(data, h.ProcessorDecider.Processor.DefaultProcessor)

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
