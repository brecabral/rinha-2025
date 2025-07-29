package decision

import (
	"net/http"

	"github.com/brecabral/rinha-2025/internal/payment"
)

const urlDefaultPayment = "http://payment-processor-default:8080"

var defaultPaymentClient = payment.PaymentClient{
	Client:           &http.Client{},
	BaseUrl:          urlDefaultPayment,
	DefaultProcessor: true,
}

const urlFallbackPayment = "http://payment-processor-fallback:8080"

var fallbackPaymentClient = payment.PaymentClient{
	Client:           &http.Client{},
	BaseUrl:          urlFallbackPayment,
	DefaultProcessor: false,
}

func Decider() payment.PaymentClient {
	return defaultPaymentClient
}
