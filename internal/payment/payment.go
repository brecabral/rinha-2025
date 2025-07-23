package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PaymentClient struct {
	Client *http.Client
	BaseUrl string
}

type PaymentRequest struct{
    CorrelationID string    `json:"correlationId"`
    Amount        float64   `json:"amount"`
    RequestedAt   time.Time `json:"requestedAt"`
}

func (p *PaymentClient) PostPayment(ctx context.Context, reqBody PaymentRequest) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := p.BaseUrl + "/payments"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := p.Client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.StatusCode >= 200 && res.StatusCode < 300 {
        return nil
    }

    return fmt.Errorf("erro no pagamento: status %d", res.StatusCode)
}

type PaymentHealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

func (p *PaymentClient) GetPaymentHealth(ctx context.Context) (*PaymentHealthResponse, error){
	url := p.BaseUrl + "/payments/service-health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var health PaymentHealthResponse
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return nil, err
	}
	return &health, nil
}