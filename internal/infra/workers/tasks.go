package workers

import (
	"errors"

	"github.com/brecabral/rinha-2025/internal/infra/clients"
	"github.com/brecabral/rinha-2025/internal/infra/database"
	"github.com/brecabral/rinha-2025/internal/infra/queue"
)

type Task interface {
	Process() bool
}

type paymentTask struct {
	task              queue.TaskInfo
	defaultProcessor  *clients.ProcessorClient
	fallbackProcessor *clients.ProcessorClient
	transactionsDB    *database.Transactions
}

func NewPaymentTask(taskInfo queue.TaskInfo, dp, fp *clients.ProcessorClient, db *database.Transactions) *paymentTask {
	return &paymentTask{
		task:              taskInfo,
		defaultProcessor:  dp,
		fallbackProcessor: fp,
		transactionsDB:    db,
	}
}

func (t *paymentTask) Process() bool {
	firstErr := t.defaultProcessor.PostPayment(
		t.task.ID,
		t.task.Amount,
		t.task.RequestedAt)
	if firstErr == nil {
		t.transactionsDB.SaveTransaction(
			t.task.ID,
			t.task.Amount,
			t.task.RequestedAt,
			true)
	} else if errors.Is(firstErr, clients.ErrRetry) {
		secondErr := t.fallbackProcessor.PostPayment(
			t.task.ID,
			t.task.Amount,
			t.task.RequestedAt)
		if secondErr == nil {
			t.transactionsDB.SaveTransaction(
				t.task.ID,
				t.task.Amount,
				t.task.RequestedAt,
				false)
		} else if errors.Is(secondErr, clients.ErrRetry) {
			return false
		}
	}
	return true
}
