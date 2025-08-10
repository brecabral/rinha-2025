package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/brecabral/rinha-2025/internal/dto"
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
	DB    *sql.DB
	stmts *stmtsTransactions
}

func NewTransactionsDB(db *sql.DB) (*Transactions, error) {
	preparedStmts, err := prepareStmts(db)
	if err != nil {
		return nil, err
	}

	return &Transactions{
		DB:    db,
		stmts: preparedStmts,
	}, nil
}

func (t *Transactions) SaveTransaction(data dto.DatabaseSaveTransaction) error {
	processorType := "default"
	if !data.ProcessorDefault {
		processorType = "fallback"
	}
	_, err := t.stmts.insertTransactionStmt.Exec(data.ID, data.Amount, data.RequestedAt, processorType)
	return err
}

func readTransactions(rows *sql.Rows) dto.DatabaseReadTransactions {
	var data dto.DatabaseReadTransactions
	for rows.Next() {
		var (
			processorType string
			totalRequests int
			totalAmount   float64
		)
		rows.Scan(&processorType, &totalRequests, &totalAmount)
		switch processorType {
		case "default":
			data.DefaultProcessor.TotalRequests = totalRequests
			data.DefaultProcessor.TotalAmount = totalAmount
		case "fallback":
			data.FallbackProcessor.TotalRequests = totalRequests
			data.FallbackProcessor.TotalAmount = totalAmount
		}
	}
	return data
}

func (t *Transactions) ReadAllTransactions() (dto.DatabaseReadTransactions, error) {
	var data dto.DatabaseReadTransactions
	rows, err := t.stmts.selectAllTransactionsStmt.Query()
	if err != nil {
		return data, err
	}
	defer rows.Close()
	data = readTransactions(rows)
	return data, nil
}

func (t *Transactions) ReadTransactionsOnPeriod(from time.Time, to time.Time) (dto.DatabaseReadTransactions, error) {
	var data dto.DatabaseReadTransactions
	rows, err := t.stmts.selectTransactionOnPeriodStmt.Query(from, to)
	if err != nil {
		return data, err
	}
	defer rows.Close()
	data = readTransactions(rows)
	return data, nil
}

func (t *Transactions) Close() error {
	errInsert := t.stmts.insertTransactionStmt.Close()
	errSelectAll := t.stmts.selectAllTransactionsStmt.Close()
	errSelectPeriod := t.stmts.selectTransactionOnPeriodStmt.Close()
	errDB := t.DB.Close()
	return errors.Join(errInsert, errSelectAll, errSelectPeriod, errDB)
}
