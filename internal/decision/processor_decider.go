package decision

import (
	"context"
	"net/http"
	"time"

	"github.com/brecabral/rinha-2025/internal/entity"
)

const urlDefaultProcessor = "http://payment-processor-default:8080"

var defaultProcessorClient = entity.NewProcessorClient(
	&http.Client{},
	urlDefaultProcessor,
	true,
)

const urlFallbackProcessor = "http://payment-processor-fallback:8080"

var fallbackProcessorClient = entity.NewProcessorClient(
	&http.Client{},
	urlFallbackProcessor,
	false,
)

type Decider struct {
	Processor *entity.ProcessorClient
}

func NewDecider() *Decider {
	return &Decider{
		Processor: defaultProcessorClient,
	}
}

func (d *Decider) Chose() {
	if defaultProcessorClient.Failing && fallbackProcessorClient.Failing {
		<-time.After(100 * time.Millisecond)
		d.Chose()
		return
	} else if defaultProcessorClient.Failing {
		d.Processor = fallbackProcessorClient
	} else if fallbackProcessorClient.Failing {
		d.Processor = defaultProcessorClient
	}
}

func (d *Decider) CheckProcessors() {
	checkProcessorHealth(defaultProcessorClient)
	checkProcessorHealth(fallbackProcessorClient)
}

func checkProcessorHealth(processorClient *entity.ProcessorClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	health, err := processorClient.GetPaymentHealth(ctx)
	if err != nil {
		processorClient.Failing = true
		processorClient.MinResponseTime = 1000
	} else {
		processorClient.Failing = health.Failing
		processorClient.MinResponseTime = health.MinResponseTime
	}
}
