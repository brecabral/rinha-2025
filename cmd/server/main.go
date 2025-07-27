package main

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

const urlDefaultPayment = "http://payment-processor-default:8080"

var defaultPaymentClient = payment.PaymentClient{
	Client:  &http.Client{},
	BaseUrl: urlDefaultPayment,
}

func main() {
	http.HandleFunc("/payments", paymentsHandler)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func paymentsHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := defaultPaymentClient.PostPayment(ctx, body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("[ERROR] POST falhou")
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("CorrelationID: %s, Amount: %f", data.CorrelationID, data.Amount)
}
