package main

import (
	"log"
	"net/http"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/handle"
)

func main() {
	processor := decision.Decider()
	h := handle.Handler{Processor: &processor}
	http.HandleFunc("/payments", h.PaymentsHandler)
	log.Print("[INFO] Servidor ouvindo...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
