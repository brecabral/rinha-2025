package workers

import (
	"sync"
	"time"

	"github.com/brecabral/rinha-2025/internal/infra/clients"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/queue"
)

type WorkerPool struct {
	tasksQueue        *queue.TasksQueue
	transactionsDB    *database.Transactions
	defaultProcessor  *clients.ProcessorClient
	fallbackProcessor *clients.ProcessorClient
	workers           int
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

func NewWorkerPool(tq *queue.TasksQueue, db *database.Transactions, dp, fp *clients.ProcessorClient, w int) *WorkerPool {
	wp := &WorkerPool{
		tasksQueue:        tq,
		transactionsDB:    db,
		defaultProcessor:  dp,
		fallbackProcessor: fp,
		workers:           w,
		stopChan:          make(chan struct{}),
	}
	return wp
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.stopChan:
			return
		default:
			taskInfo, err := wp.tasksQueue.Dequeue()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			task := NewPaymentTask(
				taskInfo,
				wp.defaultProcessor,
				wp.fallbackProcessor,
				wp.transactionsDB)
			workDone := task.Process()
			if !workDone {
				wp.tasksQueue.Prepend(taskInfo)
			}
		}
	}
}

func (wp *WorkerPool) Stop() {
	for i := 0; i < wp.workers; i++ {
		wp.stopChan <- struct{}{}
	}
	wp.wg.Wait()
	close(wp.stopChan)
}
