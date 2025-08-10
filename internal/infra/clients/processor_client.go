package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/dto"
)

type ProcessorClient struct {
	WebClient        *http.Client
	BaseUrl          string
	DefaultProcessor bool
	Failing          bool
	MinResponseTime  time.Duration
}

func NewProcessorClient(webClient *http.Client, baseUrl string, defaultProcessor bool) *ProcessorClient {
	return &ProcessorClient{
		WebClient:        webClient,
		BaseUrl:          baseUrl,
		DefaultProcessor: defaultProcessor,
		Failing:          false,
		MinResponseTime:  0,
	}
}

func (p *ProcessorClient) PostPayment(reqBody dto.ProcessorPaymentRequest) error {
	var ErrRetryable = errors.New("retry")

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	url := p.BaseUrl + "/payments"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := p.WebClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	if res.StatusCode >= 500 && res.StatusCode < 600 {
		return ErrRetryable
	}

	return fmt.Errorf("erro no pagamento: status %d", res.StatusCode)
}

func (p *ProcessorClient) GetPaymentHealth(ctx context.Context) (*dto.ProcessorHealthResponse, error) {
	url := p.BaseUrl + "/payments/service-health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := p.WebClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var health dto.ProcessorHealthResponse
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return nil, err
	}
	return &health, nil
}

func (p *ProcessorClient) UpdateProcessorHealth() error {
	url := p.BaseUrl + "/payments/service-health"

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		p.Failing = true
		p.MinResponseTime = 500
		return err
	}

	res, err := p.WebClient.Do(req)
	if err != nil {
		p.Failing = true
		p.MinResponseTime = 500
		return err
	}
	defer res.Body.Close()

	var health dto.ProcessorHealthResponse
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		p.Failing = true
		p.MinResponseTime = 500
		return err
	}

	p.Failing = health.Failing
	p.MinResponseTime = health.MinResponseTime

	return nil
}
