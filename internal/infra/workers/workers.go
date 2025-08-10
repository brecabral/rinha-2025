package workers

import "sync"

type WorkerPool struct {
	tasksChan chan Task
	wg        sync.WaitGroup
}

func NewWorkerPool(concurrency int) *WorkerPool {
	wp := &WorkerPool{
		tasksChan: make(chan Task, 100),
	}
	for i := 0; i < concurrency; i++ {
		go wp.worker()
	}
	return wp
}

func (wp *WorkerPool) worker() {
	for task := range wp.tasksChan {
		payment, ok := task.(*PaymentTask)
		work := task.Process()
		if work && ok {
			wp.Submit(NewDatabaseTask(payment.Data, payment.ProcessorDecider, payment.PaymentsDB))
		}
		if !work {
			wp.Submit(task)
		}
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Submit(task Task) {
	wp.wg.Add(1)
	wp.tasksChan <- task
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
