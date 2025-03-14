package transactions

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/govalues/decimal"
)

func newDecimal(t *testing.T, value int64, scale int) decimal.Decimal {
	t.Helper()
	d, err := decimal.New(value, scale)
	if err != nil {
		t.Errorf("newDecimal(%d, %d) failed: %v", value, scale, err)
	}
	return d
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    []string
		want    Transaction
		wantErr bool
	}{
		{
			name: "success parse",
			data: []string{"1", "10", "DEBIT", "2025-10-01 11:12:13"},
			want: Transaction{
				TrxID:           "1",
				Amount:          newDecimal(t, 10, 0),
				Type:            TransactionTypeDebit,
				TransactionTime: time.Date(2025, 10, 01, 11, 12, 13, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "success parse decimals",
			data: []string{"1", "10.01", "CREDIT", "2025-10-01 11:12:13"},
			want: Transaction{
				TrxID:           "1",
				Amount:          newDecimal(t, 1001, 2),
				Type:            TransactionTypeCredit,
				TransactionTime: time.Date(2025, 10, 01, 11, 12, 13, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:    "fail on wrong amount",
			data:    []string{"1", "a", "CREDIT", "2025-10-01 11:12:13"},
			want:    Transaction{},
			wantErr: true,
		},
		{
			name:    "fail on wrong enum",
			data:    []string{"1", "1", "credit", "2025-10-01 11:12:13"},
			want:    Transaction{},
			wantErr: true,
		},
		{
			name:    "fail on wrong date",
			data:    []string{"1", "1", "credit", "haha"},
			want:    Transaction{},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parse(test.data)
			if test.wantErr != (err != nil) {
				t.Errorf("wantErr is %t, but err is %v", test.wantErr, err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("parse(%s) mismatch, (-want,+got):\n%s", test.data, diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type testData struct {
		startDate string
		endDate   string
	}
	tests := []struct {
		input   testData
		trxDate string
		want    bool
	}{
		{testData{"2025-01-01", "2025-01-03"}, "2024-12-31 11:12:13", false},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-01 00:00:00", true},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-01 11:12:13", true},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-02 11:12:13", true},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-03 11:12:13", true},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-03 23:59:59", true},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-03 24:00:00", false},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-04 24:00:00", false},
		{testData{"2025-01-01", "2025-01-03"}, "2025-01-04 11:12:13", false},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			startDate, _ := time.Parse(time.DateOnly, test.input.startDate)
			endDate, _ := time.Parse(time.DateOnly, test.input.endDate)
			trxDate, _ := time.Parse(time.DateTime, test.trxDate)
			trx := Transaction{
				TrxID:           "1",
				Amount:          newDecimal(t, 10, 0),
				Type:            TransactionTypeDebit,
				TransactionTime: trxDate,
			}

			f := filter(startDate, endDate)
			got := f(trx)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("filter mismatch, (-want,+got):\n%s", diff)
			}
		})
	}
}
