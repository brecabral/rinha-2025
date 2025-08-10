package workers

import (
	"log"

	"github.com/brecabral/rinha-2025/internal/decision"
	"github.com/brecabral/rinha-2025/internal/dto"
	"github.com/brecabral/rinha-2025/internal/infra/database"
)

type Task interface {
	Process() bool
}

type PaymentTask struct {
	Data             dto.ProcessorPaymentRequest
	ProcessorDecider *decision.Decider
	PaymentsDB       *database.Database
}

func NewPaymentTask(data dto.ProcessorPaymentRequest, decider *decision.Decider, db *database.Database) *PaymentTask {
	return &PaymentTask{
		Data:             data,
		ProcessorDecider: decider,
		PaymentsDB:       db,
	}
}

func (t *PaymentTask) Process() bool {
	err := t.ProcessorDecider.Processor.PostPayment(t.Data)
	log.Printf("%v", err)
	return err == nil
}

type DatabaseTask struct {
	Data             dto.ProcessorPaymentRequest
	ProcessorDecider *decision.Decider
	PaymentsDB       *database.Database
}

func NewDatabaseTask(data dto.ProcessorPaymentRequest, decider *decision.Decider, db *database.Database) *DatabaseTask {
	return &DatabaseTask{
		Data:             data,
		ProcessorDecider: decider,
		PaymentsDB:       db,
	}
}

func (t *DatabaseTask) Process() bool {
	err := t.PaymentsDB.SaveTransaction(t.Data, t.ProcessorDecider.Processor.DefaultProcessor)
	return err == nil
}
