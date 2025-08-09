package entity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brecabral/rinha-2025/internal/dto"
)

type ProcessorClient struct {
	WebClient        *http.Client
	BaseUrl          string
	DefaultProcessor bool
}

func NewProcessorClient(webClient *http.Client, baseUrl string, defaultProcessor bool) *ProcessorClient {
	return &ProcessorClient{
		WebClient:        webClient,
		BaseUrl:          baseUrl,
		DefaultProcessor: defaultProcessor,
	}
}

func (p *ProcessorClient) PostPayment(ctx context.Context, reqBody dto.ProcessorPaymentRequest) error {
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
	res, err := p.WebClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
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
