package domain

import "time"

type ExternalPaymentRequest struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

type ExternalPaymentHealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

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
