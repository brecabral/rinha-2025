package database

import (
	"database/sql"
	"errors"
	"time"
)

const insertTransaction = `INSERT INTO 
								transactions(id, amount, requested_at, processor_type) 
							VALUES 
								($1, $2, $3, $4);`

const selectAllTransactions = `SELECT 
									processor_type, count(*), COALESCE(sum(amount), 0) 
								FROM 
									transactions
								GROUP BY 
									processor_type;`

const selectTransactionOnPeriod = `SELECT 
										processor_type, count(*), COALESCE(sum(amount), 0) 
									FROM 
										transactions 
									WHERE 
										requested_at BETWEEN $1 AND $2
									GROUP BY 
										processor_type;`

type stmtsTransactions struct {
	insertTransactionStmt         *sql.Stmt
	selectAllTransactionsStmt     *sql.Stmt
	selectTransactionOnPeriodStmt *sql.Stmt
}

func prepareStmts(db *sql.DB) (*stmtsTransactions, error) {
	var stmts stmtsTransactions

	preparedInsertTransaction, err := db.Prepare(insertTransaction)
	if err != nil {
		return nil, err
	}
	stmts.insertTransactionStmt = preparedInsertTransaction

	preparedSelectAllTransactions, err := db.Prepare(selectAllTransactions)
	if err != nil {
		return nil, err
	}
	stmts.selectAllTransactionsStmt = preparedSelectAllTransactions

	preparedSelectTransactionOnPeriod, err := db.Prepare(selectTransactionOnPeriod)
	if err != nil {
		return nil, err
	}
	stmts.selectTransactionOnPeriodStmt = preparedSelectTransactionOnPeriod

	return &stmts, nil
}

type Transactions struct {
	database *sql.DB
	stmts    *stmtsTransactions
}

func NewTransactionsDB(db *sql.DB) (*Transactions, error) {
	preparedStmts, err := prepareStmts(db)
	if err != nil {
		return nil, err
	}

	return &Transactions{
		database: db,
		stmts:    preparedStmts,
	}, nil
}

func (t *Transactions) SaveTransaction(id string, amount float64, requestedAt time.Time, isDefault bool) error {
	processorType := "fallback"
	if isDefault {
		processorType = "default"
	}
	_, err := t.stmts.insertTransactionStmt.Exec(id, amount, requestedAt, processorType)
	return err
}

type TransactionsSummaryByProcessor struct {
	TotalRequests int
	TotalAmount   float64
}

type TransactionsSummary struct {
	DefaultProcessor  TransactionsSummaryByProcessor
	FallbackProcessor TransactionsSummaryByProcessor
}

func readRows(rows *sql.Rows) TransactionsSummary {
	var summary TransactionsSummary
	for rows.Next() {
		var (
			processorType string
			totalRequests int
			totalAmount   float64
		)
		rows.Scan(&processorType, &totalRequests, &totalAmount)
		switch processorType {
		case "default":
			summary.DefaultProcessor.TotalRequests = totalRequests
			summary.DefaultProcessor.TotalAmount = totalAmount
		case "fallback":
			summary.FallbackProcessor.TotalRequests = totalRequests
			summary.FallbackProcessor.TotalAmount = totalAmount
		}
	}
	return summary
}

func (t *Transactions) ReadAllTransactions() (TransactionsSummary, error) {
	var summary TransactionsSummary
	rows, err := t.stmts.selectAllTransactionsStmt.Query()
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	summary = readRows(rows)
	return summary, nil
}

func (t *Transactions) ReadTransactionsOnPeriod(from time.Time, to time.Time) (TransactionsSummary, error) {
	var summary TransactionsSummary
	rows, err := t.stmts.selectTransactionOnPeriodStmt.Query(from, to)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	summary = readRows(rows)
	return summary, nil
}

func (t *Transactions) Close() error {
	errInsert := t.stmts.insertTransactionStmt.Close()
	errSelectAll := t.stmts.selectAllTransactionsStmt.Close()
	errSelectPeriod := t.stmts.selectTransactionOnPeriodStmt.Close()
	errDB := t.database.Close()
	return errors.Join(errInsert, errSelectAll, errSelectPeriod, errDB)
}
