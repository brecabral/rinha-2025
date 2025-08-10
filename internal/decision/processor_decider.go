package decision

import (
	"sync"
	"time"

	"github.com/brecabral/rinha-2025/internal/infra/clients"
)

type Decider struct {
	defaultProcessor  *clients.ProcessorClient
	fallbackProcessor *clients.ProcessorClient
	mu                sync.RWMutex
}

func NewDecider(dp *clients.ProcessorClient, fp *clients.ProcessorClient) *Decider {
	return &Decider{
		defaultProcessor:  dp,
		fallbackProcessor: fp,
	}
}

func (d *Decider) ChooseProcessor() *clients.ProcessorClient {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.fallbackProcessor.Failing {
		return d.defaultProcessor
	}
	if d.defaultProcessor.Failing {
		return d.fallbackProcessor
	}
	if d.defaultProcessor.MinResponseTime > 100 && d.defaultProcessor.MinResponseTime > d.fallbackProcessor.MinResponseTime {
		return d.fallbackProcessor
	}
	return d.defaultProcessor
}

func (d *Decider) UpdateProcessorHealth(defaultProcessor bool, failing bool, minResponseTime time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var processor *clients.ProcessorClient
	if defaultProcessor {
		processor = d.defaultProcessor
	} else {
		processor = d.fallbackProcessor
	}
	processor.Failing = failing
	processor.MinResponseTime = minResponseTime
}

func (d *Decider) CheckProcessorsHealth() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.defaultProcessor.UpdateProcessorHealth()
	d.fallbackProcessor.UpdateProcessorHealth()
}

func (d *Decider) StartHealthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		d.CheckProcessorsHealth()
	}
}
