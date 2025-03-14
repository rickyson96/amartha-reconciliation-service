package statements

import (
	"io"
	"time"

	"github.com/govalues/decimal"
	csvparser "github.com/rickyson96/amartha-reconciliation-service/internal/csv_parser"
)

type (
	Statement struct {
		UniqueIdentifier string
		Amount           decimal.Decimal
		Date             time.Time
	}
)

func parse(data []string) (Statement, error) {
	amount, err := decimal.Parse(data[1])
	if err != nil {
		return Statement{}, err
	}

	statementTime, err := time.Parse(time.DateOnly, data[2])
	if err != nil {
		return Statement{}, err
	}

	return Statement{
		UniqueIdentifier: data[0],
		Amount:           amount,
		Date:             statementTime,
	}, nil
}

func filter(startDate, endDate time.Time) func(data Statement) bool {
	return func(data Statement) bool {
		if data.Date.Before(startDate) || data.Date.After(endDate) {
			return false
		}
		return true
	}
}

func NewCSVParser(filepath io.Reader, startDate, endDate time.Time) *csvparser.CSVParser[Statement] {
	return csvparser.NewCSVParser(
		filepath,
		parse,
		filter(startDate, endDate),
		csvparser.CSVParserOptions{
			ContainsHeader: true,
			FieldPerRow:    3,
		})
}
