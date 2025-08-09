package decision

import (
	"net/http"

	"github.com/brecabral/rinha-2025/internal/entity"
)

const urlDefaultPayment = "http://payment-processor-default:8080"

var defaultPaymentClient = entity.NewProcessorClient(
	&http.Client{},
	urlDefaultPayment,
	true,
)

const urlFallbackPayment = "http://payment-processor-fallback:8080"

var fallbackPaymentClient = entity.NewProcessorClient(
	&http.Client{},
	urlFallbackPayment,
	false,
)

func Decider() *entity.ProcessorClient {
	return defaultPaymentClient
}
