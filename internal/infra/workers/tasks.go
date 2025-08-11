package workers

import (
	"context"

	"time"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/dto"
	"github.com/brecabral/rinha-2025/internal/infra/clients"
	"github.com/brecabral/rinha-2025/internal/infra/database"
)

type Task interface {
	Process() error
}

type TaskInfo struct {
	ID          string
	Amount      float64
	RequestedAt time.Time
}

type PaymentTask struct {
	Data             TaskInfo
	ProcessorDecider *decision.Decider
	PaymentsDB       *database.Transactions
	DefaultProcessor bool
}

func NewPaymentTask(data TaskInfo, pd *decision.Decider, db *database.Transactions) *PaymentTask {
	return &PaymentTask{
		Data:             data,
		ProcessorDecider: pd,
		PaymentsDB:       db,
		DefaultProcessor: true,
	}
}

func (t *PaymentTask) Process() error {
	req := dto.ProcessorPaymentRequest{
		CorrelationID: t.Data.ID,
		Amount:        t.Data.Amount,
		RequestedAt:   t.Data.RequestedAt,
	}
	var (
		err       error
		processor *clients.ProcessorClient
	)

	processor = t.ProcessorDecider.ChooseProcessor()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err = processor.PostPayment(ctx, req)
	if err != nil {
		t.ProcessorDecider.UpdateProcessorFailing(processor.DefaultProcessor, true)
		<-time.Tick(100 * time.Millisecond)
		processor = t.ProcessorDecider.ChooseProcessor()
		err = processor.PostPayment(ctx, req)
	}

	t.DefaultProcessor = processor.DefaultProcessor
	return err
}

type DatabaseTask struct {
	Data             TaskInfo
	PaymentsDB       *database.Transactions
	DefaultProcessor bool
}

func NewDatabaseTask(paymentTask *PaymentTask) *DatabaseTask {
	return &DatabaseTask{
		Data:             paymentTask.Data,
		PaymentsDB:       paymentTask.PaymentsDB,
		DefaultProcessor: paymentTask.DefaultProcessor,
	}
}

func (t *DatabaseTask) Process() error {
	transaction := dto.DatabaseSaveTransaction{
		ID:               t.Data.ID,
		Amount:           t.Data.Amount,
		RequestedAt:      t.Data.RequestedAt,
		ProcessorDefault: t.DefaultProcessor,
	}
	return t.PaymentsDB.SaveTransaction(transaction)
}
