package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

var (
	ErrRetry       = errors.New("retry")
	ErrRetryOnSame = errors.New("retry on same")
	ErrNoRetry     = errors.New("no retry")
)

type ProcessorClient struct {
	WebClient *http.Client
	BaseUrl   string
	IsDefault bool
}

func NewProcessorClient(webClient *http.Client, baseUrl string, isDefault bool) *ProcessorClient {
	return &ProcessorClient{
		WebClient: webClient,
		BaseUrl:   baseUrl,
		IsDefault: isDefault,
	}
}

type processorPaymentRequest struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

func (p *ProcessorClient) PostPayment(id string, amount float64, requestAt time.Time) error {
	reqBody := processorPaymentRequest{
		CorrelationID: id,
		Amount:        amount,
		RequestedAt:   requestAt,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return ErrNoRetry
	}

	url := p.BaseUrl + "/payments"
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ErrNoRetry
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := p.WebClient.Do(req)
	if err != nil {
		processorType := "fallback"
		if p.IsDefault {
			processorType = "default"
		}
		log.Printf("[ERROR] processor: %s, não aceitou o pagamento: %v", processorType, err)
		return ErrRetryOnSame
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		processorType := "fallback"
		if p.IsDefault {
			processorType = "default"
		}
		log.Printf("[INFO] processor: %s, aceitou o pagamento", processorType)
		return nil
	}

	if res.StatusCode >= 500 && res.StatusCode < 600 {
		processorType := "fallback"
		if p.IsDefault {
			processorType = "default"
		}
		log.Printf("[ERROR] processor: %s, não aceitou o pagamento status code: %d", processorType, res.StatusCode)
		return ErrRetry
	}

	return ErrNoRetry
}
