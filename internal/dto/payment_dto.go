package dto

type PaymentsRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type PaymentsSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	DefaultProcessor  PaymentsSummary `json:"default"`
	FallbackProcessor PaymentsSummary `json:"fallback"`
}
