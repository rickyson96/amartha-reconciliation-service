package reconciliation

import (
	"fmt"
	"time"

	"github.com/govalues/decimal"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
)

type Result struct {
	Processed int
	Match     int
	Unmatched struct {
		Transactions []transactions.Transaction
		Statements   map[string][]statements.Statement
	}
}

func uniqueID(trxType transactions.TransactionType, amount decimal.Decimal, date time.Time) string {
	return fmt.Sprintf("%s:%s:%s", trxType, amount.String(), date.Format(time.DateOnly))
}

func Process(trxs []transactions.Transaction, stmtFiles map[string][]statements.Statement) Result {
	uniqueTransactions := make(map[string][]transactions.Transaction, len(trxs))
	var result Result

	for _, t := range trxs {
		result.Processed++

		id := uniqueID(t.Type, t.Amount, t.TransactionTime)
		uniqueTransactions = appendMapOfSlices(uniqueTransactions, id, t)
	}

	for fileName, stmts := range stmtFiles {
		for _, s := range stmts {
			result.Processed++

			stmtType := transactions.TransactionTypeCredit
			if s.Amount.IsNeg() {
				stmtType = transactions.TransactionTypeDebit
			}
			id := uniqueID(stmtType, s.Amount.Abs(), s.Date)
			_, ok := uniqueTransactions[id]
			if !ok {
				result.Unmatched.Statements = appendMapOfSlices(result.Unmatched.Statements, fileName, s)
				continue
			}

			result.Match += 2 // 1 for transaction, 1 for statement
			popMapOfSlices(uniqueTransactions, id)
		}
	}

	for _, trxs := range uniqueTransactions {
		result.Unmatched.Transactions = append(result.Unmatched.Transactions, trxs...)
	}
	return result
}
