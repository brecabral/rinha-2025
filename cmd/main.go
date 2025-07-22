package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type PaymentHealth struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

func main(){
	http.HandleFunc("/", teste)
	log.Print("ouvindo")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func teste(w http.ResponseWriter, r *http.Request) {
	paymentHealth := getPaymentHealth()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paymentHealth)
}

func getPaymentHealth() *PaymentHealth{
	req, err := http.NewRequest(http.MethodGet, "http://payment-processor-default:8080/payments/service-health", nil)
	if err != nil {
		panic(err)
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var health *PaymentHealth
	err = json.Unmarshal(body, &health)
	if err != nil {
		panic(err)
	}
	return health
}