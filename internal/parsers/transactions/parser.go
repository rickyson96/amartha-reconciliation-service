package transactions

import (
	"io"
	"time"

	"github.com/govalues/decimal"
	csvparser "github.com/rickyson96/amartha-reconciliation-service/internal/csv_parser"
)

//go:generate go-enum --alias=CREDIT:Credit,DEBIT:Debit
type (
	// ENUM(CREDIT, DEBIT)
	TransactionType int

	Transaction struct {
		TrxID           string
		Amount          decimal.Decimal
		Type            TransactionType
		TransactionTime time.Time
	}
)

func parse(data []string) (Transaction, error) {
	amount, err := decimal.Parse(data[1])
	if err != nil {
		return Transaction{}, err
	}

	transactionType, err := ParseTransactionType(data[2])
	if err != nil {
		return Transaction{}, err
	}

	transactionTime, err := time.Parse(time.DateTime, data[3])
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		TrxID:           data[0],
		Amount:          amount,
		Type:            transactionType,
		TransactionTime: transactionTime,
	}, nil
}

func filter(startDate, endDate time.Time) func(data Transaction) bool {
	return func(data Transaction) bool {
		if data.TransactionTime.Before(startDate) || !data.TransactionTime.Before(endDate.AddDate(0, 0, 1)) {
			return false
		}
		return true
	}
}

func NewCSVParser(file io.Reader, startDate, endDate time.Time) *csvparser.CSVParser[Transaction] {
	return csvparser.NewCSVParser(
		file,
		parse,
		filter(startDate, endDate),
		csvparser.CSVParserOptions{
			ContainsHeader: true,
			FieldPerRow:    4,
		})
}
