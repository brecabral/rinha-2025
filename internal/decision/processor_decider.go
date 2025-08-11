package decision

import (
	"sync"
	"time"

	"github.com/brecabral/rinha-2025/internal/infra/cache"
	"github.com/brecabral/rinha-2025/internal/infra/clients"
)

const defaultProcessor = true
const fallbackProcessor = false

type Decider struct {
	defaultProcessorClient  *clients.ProcessorClient
	fallbackProcessorClient *clients.ProcessorClient
	processorStatus         *cache.ProcessorsCache
	mu                      sync.RWMutex
}

func NewDecider(dp *clients.ProcessorClient, fp *clients.ProcessorClient, ps *cache.ProcessorsCache) *Decider {
	return &Decider{
		defaultProcessorClient:  dp,
		fallbackProcessorClient: fp,
		processorStatus:         ps,
	}
}

func (d *Decider) ChooseProcessor() *clients.ProcessorClient {
	d.mu.RLock()
	defer d.mu.RUnlock()

	defaultProcessorFailing, err := d.processorStatus.GetProcessorFailing(defaultProcessor)
	if err != nil {
		d.processorStatus.SetProcessorFailingTrue(defaultProcessor)
	}
	fallbackProcessorFailing, err := d.processorStatus.GetProcessorFailing(fallbackProcessor)
	if err != nil {
		d.processorStatus.SetProcessorFailingTrue(fallbackProcessor)
	}

	for defaultProcessorFailing && fallbackProcessorFailing {
		ticker := time.NewTicker(500 * time.Millisecond)
		<-ticker.C
		defaultProcessorFailing, _ = d.processorStatus.GetProcessorFailing(defaultProcessor)
		fallbackProcessorFailing, _ = d.processorStatus.GetProcessorFailing(fallbackProcessor)
	}

	if !defaultProcessorFailing {

		return d.defaultProcessorClient
	}
	return d.fallbackProcessorClient
}

func (d *Decider) UpdateProcessorFailing(defaultProcessor bool, failing bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.processorStatus.SetProcessorFailing(defaultProcessor, failing)
}

func (d *Decider) CheckProcessorsHealth() {
	d.mu.Lock()
	defer d.mu.Unlock()

	defaultHealth, err := d.defaultProcessorClient.GetProcessorHealth()
	if err != nil {
		d.processorStatus.SetProcessorFailingTrue(defaultProcessor)
	} else {
		d.processorStatus.SetProcessorFailing(defaultProcessor, defaultHealth.Failing)
	}

	fallbackHealth, err := d.fallbackProcessorClient.GetProcessorHealth()
	if err != nil {
		d.processorStatus.SetProcessorFailingTrue(fallbackProcessor)
	} else {
		d.processorStatus.SetProcessorFailing(fallbackProcessor, fallbackHealth.Failing)
	}
}

func (d *Decider) StartHealthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		d.CheckProcessorsHealth()
	}
}
