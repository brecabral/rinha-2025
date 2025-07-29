package handle

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/domain"
	"github.com/brecabral/rinha-2025/internal/payment"
	"github.com/brecabral/rinha-2025/internal/store"
)

type Handler struct {
	Processor      *payment.PaymentClient
	DatabaseClient *store.Database
}

func (h *Handler) PaymentsHandler(w http.ResponseWriter, r *http.Request) {
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

	var data domain.PaymentRequest
	if err := json.Unmarshal(requestBody, &data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	body := domain.ExternalPaymentRequest{
		CorrelationID: data.CorrelationID,
		Amount:        data.Amount,
		RequestedAt:   time.Now(),
	}

	if err := h.Processor.PostPayment(ctx, body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("[ERROR] POST falhou")
		return
	}

	h.DatabaseClient.SaveTransaction(data, h.Processor.DefaultProcessor)

	w.WriteHeader(http.StatusOK)
	log.Printf("CorrelationID: %s, Amount: %f", data.CorrelationID, data.Amount)
}

func (h *Handler) PaymentsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	log.Print("[INFO] Requisição recebida")

	query := r.URL.Query()
	from := query.Get("from")
	to := query.Get("to")

	var summary *domain.PaymentsSummaryResponse
	var err error

	if (from != "") && (to != "") {
		layout := time.RFC3339
		timeFrom, err := time.Parse(layout, from)
		if err != nil {
			log.Printf("[ERROR] (from - %v) não pode ser parseado: %v", from, err)
			return
		}
		timeTo, err := time.Parse(layout, to)
		if err != nil {
			log.Printf("[ERROR] (to - %v) não pode ser parseado: %v", to, err)
			return
		}
		summary, err = h.DatabaseClient.ReadTransactionsOnPeriod(timeFrom, timeTo)
		if err != nil {
			log.Printf("[ERROR] Não foi possivel ler do banco de dados: %v", err)
			return
		}
	} else {
		summary, err = h.DatabaseClient.ReadAllTransactions()
		if err != nil {
			log.Printf("[ERROR] Não foi possivel ler do banco de dados: %v", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
