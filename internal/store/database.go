package store

import (
	"database/sql"
	"time"

	"github.com/brecabral/rinha-2025/internal/domain"
)

const insertTransactionDefault = `INSERT INTO transactionsDefault(id, amount)
								 VALUES ($1, $2);`

const selectTotalDefault = `SELECT count(*), COALESCE(sum(amount), 0) 
							FROM transactionsDefault;`

const selectPeriodDefault = `SELECT count(*), COALESCE(sum(amount), 0) 
							FROM transactionsDefault 
							WHERE created_at BETWEEN $1 AND $2;`

const insertTransactionFallback = `INSERT INTO transactionsFallback(id, amount)
								 VALUES ($1, $2);`

const selectTotalFallback = `SELECT count(*), COALESCE(sum(amount), 0) 
							FROM transactionsFallback;`

const selectPeriodFallback = `SELECT count(*), COALESCE(sum(amount), 0) 
							FROM transactionsFallback 
							WHERE created_at BETWEEN $1 AND $2;`

type TableStmt struct {
	InsertStmt       *sql.Stmt
	SelectTotalStmt  *sql.Stmt
	SelectPeriodStmt *sql.Stmt
}

type Database struct {
	Client        *sql.DB
	DefaultTable  TableStmt
	FallbackTable TableStmt
}

func NewDatabase(client *sql.DB) (*Database, error) {
	stmtInsertDefault, err := client.Prepare(insertTransactionDefault)
	if err != nil {
		return nil, err
	}

	stmtTotalDefault, err := client.Prepare(selectTotalDefault)
	if err != nil {
		return nil, err
	}

	stmtPeriodDefault, err := client.Prepare(selectPeriodDefault)
	if err != nil {
		return nil, err
	}

	tableDefault := TableStmt{
		InsertStmt:       stmtInsertDefault,
		SelectTotalStmt:  stmtTotalDefault,
		SelectPeriodStmt: stmtPeriodDefault,
	}

	stmtInsertFallback, err := client.Prepare(insertTransactionFallback)
	if err != nil {
		return nil, err
	}

	stmtTotalFallback, err := client.Prepare(selectTotalFallback)
	if err != nil {
		return nil, err
	}

	stmtPeriodFallback, err := client.Prepare(selectPeriodFallback)
	if err != nil {
		return nil, err
	}

	tableFallback := TableStmt{
		InsertStmt:       stmtInsertFallback,
		SelectTotalStmt:  stmtTotalFallback,
		SelectPeriodStmt: stmtPeriodFallback,
	}

	return &Database{
		Client:        client,
		DefaultTable:  tableDefault,
		FallbackTable: tableFallback,
	}, nil
}

func (d *Database) SaveTransaction(data domain.PaymentRequest, defaultProcessor bool) error {
	var err error
	if defaultProcessor {
		_, err = d.DefaultTable.InsertStmt.Exec(data.CorrelationID, data.Amount)
	} else {
		_, err = d.FallbackTable.InsertStmt.Exec(data.CorrelationID, data.Amount)
	}
	return err
}

func (d *Database) ReadAllTransactions() (*domain.PaymentsSummaryResponse, error) {
	row := d.DefaultTable.SelectTotalStmt.QueryRow()
	var summaryDefault domain.PaymentSummary

	err := row.Scan(&summaryDefault.TotalRequests, &summaryDefault.TotalAmount)
	if err != nil {
		return nil, err
	}

	row = d.FallbackTable.SelectTotalStmt.QueryRow()
	var summaryFallback domain.PaymentSummary

	err = row.Scan(&summaryFallback.TotalRequests, &summaryFallback.TotalAmount)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentsSummaryResponse{
		DefaultProcessor:  summaryDefault,
		FallbackProcessor: summaryFallback,
	}, nil
}

func (d *Database) ReadTransactionsOnPeriod(from time.Time, to time.Time) (*domain.PaymentsSummaryResponse, error) {
	row := d.DefaultTable.SelectPeriodStmt.QueryRow(from, to)
	var summaryDefault domain.PaymentSummary

	err := row.Scan(&summaryDefault.TotalRequests, &summaryDefault.TotalAmount)
	if err != nil {
		return nil, err
	}

	row = d.FallbackTable.SelectPeriodStmt.QueryRow(from, to)
	var summaryFallback domain.PaymentSummary

	err = row.Scan(&summaryFallback.TotalRequests, &summaryFallback.TotalAmount)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentsSummaryResponse{
		DefaultProcessor:  summaryDefault,
		FallbackProcessor: summaryFallback,
	}, nil
}

func (d *Database) Close() error {
	d.DefaultTable.InsertStmt.Close()
	d.DefaultTable.SelectTotalStmt.Close()
	d.DefaultTable.SelectPeriodStmt.Close()

	d.FallbackTable.InsertStmt.Close()
	d.FallbackTable.SelectTotalStmt.Close()
	d.FallbackTable.SelectPeriodStmt.Close()

	return d.Client.Close()
}
