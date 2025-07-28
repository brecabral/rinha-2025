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
)

type Handler struct {
	Processor *payment.PaymentClient
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

	w.WriteHeader(http.StatusOK)
	log.Printf("CorrelationID: %s, Amount: %f", data.CorrelationID, data.Amount)
}
