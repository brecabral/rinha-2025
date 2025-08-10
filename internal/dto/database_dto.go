package dto

import "time"

type DatabaseSaveTransaction struct {
	ID               string
	Amount           float64
	RequestedAt      time.Time
	ProcessorDefault bool
}

type TransactionsSummaryByProcessor struct {
	TotalRequests int
	TotalAmount   float64
}

type DatabaseReadTransactions struct {
	DefaultProcessor  TransactionsSummaryByProcessor
	FallbackProcessor TransactionsSummaryByProcessor
}
