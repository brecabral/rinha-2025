package dto

type PaymentRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type PaymentSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	DefaultProcessor  PaymentSummary `json:"default"`
	FallbackProcessor PaymentSummary `json:"fallback"`
}
