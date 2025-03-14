package statements

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/rickyson96/amartha-reconciliation-service/internal/testutils"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    []string
		want    Statement
		wantErr bool
	}{
		{
			name: "success",
			data: []string{"1", "10", "2025-01-02"},
			want: Statement{
				UniqueIdentifier: "1",
				Amount:           testutils.NewDecimal(t, 10, 0),
				Date:             time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "negative value",
			data: []string{"1", "-10", "2025-01-02"},
			want: Statement{
				UniqueIdentifier: "1",
				Amount:           testutils.NewDecimal(t, -10, 0),
				Date:             time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{"wrong amount", []string{"1", "woo", "2025-01-02"}, Statement{}, true},
		{"wrong date", []string{"1", "10", "woo"}, Statement{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parse(test.data)
			if test.wantErr != (err != nil) {
				t.Errorf("wantErr is %t, but err is %v", test.wantErr, err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("parse(%v) mismatch, (-want,+got):\n%s", test.data, diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type input struct {
		startDate string
		endDate   string
	}
	tests := []struct {
		input input
		date  string
		want  bool
	}{
		{input{"2025-01-02", "2025-01-04"}, "2023-01-02", false},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-01", false},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-02", true},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-03", true},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-04", true},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-05", false},
		{input{"2025-01-02", "2025-01-04"}, "2025-01-10", false},
		{input{"2025-01-02", "2025-01-04"}, "2099-01-02", false},
	}

	for _, test := range tests {
		t.Run(test.date, func(t *testing.T) {
			startDate, _ := time.Parse(time.DateOnly, test.input.startDate)
			endDate, _ := time.Parse(time.DateOnly, test.input.endDate)
			stmtDate, _ := time.Parse(time.DateOnly, test.date)
			stmt := Statement{
				UniqueIdentifier: "1",
				Amount:           testutils.NewDecimal(t, 10, 0),
				Date:             stmtDate,
			}
			got := filter(startDate, endDate)(stmt)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("filter(%v,%v)(%v) mismatch, (-want,+got):\n%s", startDate, endDate, stmtDate, diff)
			}
		})
	}
}
