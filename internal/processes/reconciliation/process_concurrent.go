package reconciliation

import (
	"context"
	"io"
	"sync"

	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
	"golang.org/x/sync/errgroup"
)

// Reader is the function to stream the data
// The reader provider must return io.EOF as error when the data
// reading is done.
type Reader[T any] func() (T, error)

// StatementFilePair is the pair of filename and statement data
type StatementFilePair struct {
	Name      string
	Statement statements.Statement
}

type workingMap struct {
	m           sync.Mutex
	result      Result
	transaction map[string][]transactions.Transaction
	statements  map[string][]StatementFilePair
}

func transactionReader(wm *workingMap, trxCh <-chan transactions.Transaction) {
	for {
		trx, more := <-trxCh
		if !more {
			return
		}

		transactionRead(wm, trx)
	}
}

func transactionRead(wm *workingMap, trx transactions.Transaction) {
	id := uniqueID(trx.Type, trx.Amount, trx.TransactionTime)

	wm.m.Lock()
	defer wm.m.Unlock()

	wm.result.Processed++
	if _, ok := wm.statements[id]; ok {
		wm.result.Match += 2
		popMapOfSlices(wm.statements, id)
	} else {
		wm.transaction = appendMapOfSlices(wm.transaction, id, trx)
	}
}

func statementReader(wm *workingMap, stmtCh <-chan StatementFilePair) {
	for {
		stmt, more := <-stmtCh
		if !more {
			return
		}

		statementRead(wm, stmt)
	}
}

func statementRead(wm *workingMap, stmt StatementFilePair) {
	stmtType := transactions.TransactionTypeCredit
	if stmt.Statement.Amount.IsNeg() {
		stmtType = transactions.TransactionTypeDebit
	}

	id := uniqueID(stmtType, stmt.Statement.Amount.Abs(), stmt.Statement.Date)

	wm.m.Lock()
	defer wm.m.Unlock()

	wm.result.Processed++
	if _, ok := wm.transaction[id]; ok {
		wm.result.Match += 2
		popMapOfSlices(wm.transaction, id)
	} else {
		wm.statements = appendMapOfSlices(wm.statements, id, stmt)
	}
}

func transactionWriter(ctx context.Context, trxReader Reader[transactions.Transaction], trxCh chan<- transactions.Transaction) error {
	defer close(trxCh)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			trx, err := trxReader()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return err
			}

			trxCh <- trx
		}
	}
}

func statementWriter(ctx context.Context, stmtReader Reader[StatementFilePair], stmtCh chan<- StatementFilePair) error {
	defer close(stmtCh)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			stmt, err := stmtReader()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return err
			}

			stmtCh <- stmt
		}
	}
}

func ProcessConcurrent(trx Reader[transactions.Transaction], stmt Reader[StatementFilePair]) (Result, error) {
	wm := workingMap{
		m:           sync.Mutex{},
		result:      Result{},
		transaction: make(map[string][]transactions.Transaction),
		statements:  make(map[string][]StatementFilePair),
	}

	trxCh := make(chan transactions.Transaction)
	stmtCh := make(chan StatementFilePair)
	errg, ctx := errgroup.WithContext(context.Background())
	errg.Go(func() error {
		transactionReader(&wm, trxCh)
		return nil
	})
	errg.Go(func() error {
		statementReader(&wm, stmtCh)
		return nil
	})

	errg.Go(func() error { return transactionWriter(ctx, trx, trxCh) })
	errg.Go(func() error { return statementWriter(ctx, stmt, stmtCh) })

	if err := errg.Wait(); err != nil {
		return Result{}, err
	}

	for _, stmts := range wm.statements {
		for _, stmtPair := range stmts {
			wm.result.Unmatched.Statements = appendMapOfSlices(wm.result.Unmatched.Statements, stmtPair.Name, stmtPair.Statement)
		}
	}

	for _, trxs := range wm.transaction {
		wm.result.Unmatched.Transactions = append(wm.result.Unmatched.Transactions, trxs...)
	}

	return wm.result, nil
}

// TODO testing and implement reader on CSVParser
