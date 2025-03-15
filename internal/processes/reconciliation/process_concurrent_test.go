package reconciliation_test

import (
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
	"github.com/rickyson96/amartha-reconciliation-service/internal/processes/reconciliation"
	"github.com/rickyson96/amartha-reconciliation-service/internal/testutils"
)

type testReader[T any] struct {
	data   []T
	offset int
}

func newTestReader[S ~[]T, T any](data S) *testReader[T] {
	return &testReader[T]{data, 0}
}

func fileStatementPairConverter(from map[string][]statements.Statement) []reconciliation.StatementFilePair {
	statements := []reconciliation.StatementFilePair{}
	for file, stmts := range from {
		for _, s := range stmts {
			statements = append(statements, reconciliation.StatementFilePair{file, s})
		}
	}

	return statements
}

func (r *testReader[T]) Read() (T, error) {
	if r.offset >= len(r.data) {
		return *new(T), io.EOF
	}
	data := r.data[r.offset]
	r.offset++
	return data, nil
}

func TestProcessConcurrent(t *testing.T) {
	tests := []struct {
		name        string
		trancations []transactions.Transaction
		statements  map[string][]statements.Statement
		result      reconciliation.Result
	}{
		{
			name:        "success for empty data being processed",
			trancations: []transactions.Transaction{},
			statements:  map[string][]statements.Statement{},
			result: reconciliation.Result{
				Processed: 0,
				Match:     0,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{},
			},
		},
		{
			name: "successful process",
			trancations: []transactions.Transaction{
				{
					TrxID:           "1",
					Amount:          testutils.NewDecimal(t, 10, 0),
					Type:            transactions.TransactionTypeCredit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
			},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 2,
				Match:     2,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: nil,
					Statements:   nil,
				},
			},
		},
		{
			name: "successful process for different time",
			trancations: []transactions.Transaction{
				{
					TrxID:           "1",
					Amount:          testutils.NewDecimal(t, 10, 0),
					Type:            transactions.TransactionTypeCredit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
			},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 2,
				Match:     2,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: nil,
					Statements:   nil,
				},
			},
		},
		{
			name: "successful with negative value for debit",
			trancations: []transactions.Transaction{
				{
					TrxID:           "1",
					Amount:          testutils.NewDecimal(t, 10, 0),
					Type:            transactions.TransactionTypeDebit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
				{
					TrxID:           "2",
					Amount:          testutils.NewDecimal(t, 100, 0),
					Type:            transactions.TransactionTypeCredit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
			},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, -10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}, {
					UniqueIdentifier: "100",
					Amount:           testutils.NewDecimal(t, 100, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 4,
				Match:     4,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: nil,
					Statements:   nil,
				},
			},
		},
		{
			name: "show unmatched for unfound transactions",
			trancations: []transactions.Transaction{
				{
					TrxID:           "1",
					Amount:          testutils.NewDecimal(t, 10, 0),
					Type:            transactions.TransactionTypeCredit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
				{
					TrxID:           "2",
					Amount:          testutils.NewDecimal(t, 20, 0),
					Type:            transactions.TransactionTypeCredit,
					TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
				},
			},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 3,
				Match:     2,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: []transactions.Transaction{{
						TrxID:           "2",
						Amount:          testutils.NewDecimal(t, 20, 0),
						Type:            transactions.TransactionTypeCredit,
						TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
					}},
					Statements: nil,
				},
			},
		},
		{
			name:        "show unmatched for unfound statements",
			trancations: []transactions.Transaction{},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 1,
				Match:     0,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: nil,
					Statements: map[string][]statements.Statement{
						"bank1.csv": {{
							UniqueIdentifier: "10",
							Amount:           testutils.NewDecimal(t, 10, 0),
							Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
						}},
					},
				},
			},
		},
		{
			name: "show both unmatch for unmatch transaction and statements",
			trancations: []transactions.Transaction{{
				TrxID:           "1",
				Amount:          testutils.NewDecimal(t, 100, 0),
				Type:            transactions.TransactionTypeCredit,
				TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
			}},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 2,
				Match:     0,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: []transactions.Transaction{{
						TrxID:           "1",
						Amount:          testutils.NewDecimal(t, 100, 0),
						Type:            transactions.TransactionTypeCredit,
						TransactionTime: time.Date(2025, 03, 14, 10, 10, 10, 10, time.Local),
					}},
					Statements: map[string][]statements.Statement{
						"bank1.csv": {{
							UniqueIdentifier: "10",
							Amount:           testutils.NewDecimal(t, 10, 0),
							Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
						}},
					},
				},
			},
		},
		{
			name:        "show unmatched for unfound statements on different files",
			trancations: []transactions.Transaction{},
			statements: map[string][]statements.Statement{
				"bank1.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
				"bank2.csv": {{
					UniqueIdentifier: "10",
					Amount:           testutils.NewDecimal(t, 10, 0),
					Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
				}},
			},
			result: reconciliation.Result{
				Processed: 2,
				Match:     0,
				Unmatched: struct {
					Transactions []transactions.Transaction
					Statements   map[string][]statements.Statement
				}{
					Transactions: nil,
					Statements: map[string][]statements.Statement{
						"bank1.csv": {{
							UniqueIdentifier: "10",
							Amount:           testutils.NewDecimal(t, 10, 0),
							Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
						}},
						"bank2.csv": {{
							UniqueIdentifier: "10",
							Amount:           testutils.NewDecimal(t, 10, 0),
							Date:             time.Date(2025, 03, 14, 0, 0, 0, 0, time.Local),
						}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transactionReader := newTestReader(test.trancations)
			statementReader := newTestReader(fileStatementPairConverter(test.statements))

			got, err := reconciliation.ProcessConcurrent(transactionReader.Read, statementReader.Read)
			if err != nil {
				t.Errorf("unwanted error: %v", err)
			}
			if diff := cmp.Diff(test.result, got); diff != "" {
				t.Errorf("Process(%s, %s) mismatch, (-want,+got):\n%s", test.trancations, test.statements, diff)
			}
		})
	}
}
