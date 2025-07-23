package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/payment"
	"github.com/google/uuid"
)

const urlDefaultPayment = "http://payment-processor-default:8080"

var	defaultPaymentClient = payment.PaymentClient{
		Client: &http.Client{}, 
		BaseUrl: urlDefaultPayment,
	}

type PaymentHealth struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

func main(){
	http.HandleFunc("/get", testeGet)
	http.HandleFunc("/post", testePost)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func testeGet(w http.ResponseWriter, r *http.Request) {
	log.Print("[INFO] Iniciando GET")
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	paymentHealth, err := defaultPaymentClient.GetPaymentHealth(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "requisição inválida"})
		log.Print("[ERROR] GET falhou")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paymentHealth)
	log.Print("[INFO] GET efetuado")
}

func testePost(w http.ResponseWriter, r *http.Request) {
	log.Print("[INFO] Iniciando POST")
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	body := payment.PaymentRequest{
		CorrelationID: uuid.NewString(),
		Amount: 10.0,
		RequestedAt: time.Now(),
	}


	err := defaultPaymentClient.PostPayment(ctx, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "requisição inválida"})
		log.Print("[ERROR] POST falhou")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"info": "requisição enviada"})
	log.Print("[INFO] POST efetuado")
}