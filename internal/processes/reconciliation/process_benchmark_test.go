package reconciliation_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
	"github.com/rickyson96/amartha-reconciliation-service/internal/processes/reconciliation"
	"github.com/rickyson96/amartha-reconciliation-service/internal/testutils"
)

type testData struct {
	transactions []transactions.Transaction
	statements   map[string][]statements.Statement
}

func generateTestData(b *testing.B) testData {
	td := testData{
		transactions: []transactions.Transaction{},
		statements:   map[string][]statements.Statement{},
	}
	for i := range 100 {
		td.transactions = append(td.transactions,
			transactions.Transaction{
				TrxID:           strconv.Itoa(i),
				Amount:          testutils.NewDecimal(b, 10, 0),
				Type:            transactions.TransactionTypeCredit,
				TransactionTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			transactions.Transaction{
				TrxID:           strconv.Itoa(i),
				Amount:          testutils.NewDecimal(b, 100, 0),
				Type:            transactions.TransactionTypeDebit,
				TransactionTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		)

		td.statements[strconv.Itoa(i)] = []statements.Statement{{
			UniqueIdentifier: strconv.Itoa(i),
			Amount:           testutils.NewDecimal(b, 10, 0),
			Date:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}, {
			UniqueIdentifier: strconv.Itoa(i),
			Amount:           testutils.NewDecimal(b, -100, 0),
			Date:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}}
	}

	return td
}

func BenchmarkProcess(b *testing.B) {
	testData := generateTestData(b)

	// Reset timer to ignore setup time
	b.ResetTimer()

	for b.Loop() {
		reconciliation.Process(testData.transactions, testData.statements)
	}
}

func BenchmarkProcessConcurrent(b *testing.B) {
	testData := generateTestData(b)
	trxReader := newTestReader(testData.transactions)
	stmtReader := newTestReader(fileStatementPairConverter(testData.statements))

	// Reset timer to ignore setup time
	b.ResetTimer()

	for b.Loop() {
		reconciliation.ProcessConcurrent(trxReader.Read, stmtReader.Read)
	}
}
