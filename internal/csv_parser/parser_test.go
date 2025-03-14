package csvparser_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	csvparser "github.com/rickyson96/amartha-reconciliation-service/internal/csv_parser"
)

type TestModel struct {
	A string
	B string
}

func TestCSVParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		csvInput    string
		withHeader  bool
		fieldPerRow int
		parser      func(data []string) (TestModel, error)
		filter      func(data TestModel) bool
		want        []TestModel
		wantErr     bool
	}{
		{
			name:        "success",
			csvInput:    "a,b\nc,d",
			withHeader:  false,
			fieldPerRow: 2,
			parser: func(data []string) (TestModel, error) {
				return TestModel{data[0], data[1]}, nil
			},
			filter: func(data TestModel) bool {
				return true
			},
			want:    []TestModel{{"a", "b"}, {"c", "d"}},
			wantErr: false,
		},
		{
			name:        "contains header",
			csvInput:    "a,b\nc,d",
			withHeader:  true,
			fieldPerRow: 2,
			parser: func(data []string) (TestModel, error) {
				return TestModel{data[0], data[1]}, nil
			},
			filter: func(data TestModel) bool {
				return true
			},
			want:    []TestModel{{"c", "d"}},
			wantErr: false,
		},
		{
			name:        "field less than expected",
			csvInput:    "a,b\nc",
			withHeader:  true,
			fieldPerRow: 2,
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "field more than expected",
			csvInput:    "a,b,c",
			withHeader:  false,
			fieldPerRow: 2,
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "fail parsing",
			csvInput:    "a,b\nc,d",
			withHeader:  false,
			fieldPerRow: 2,
			parser: func(data []string) (TestModel, error) {
				return TestModel{}, errors.New("error")
			},
			wantErr: true,
		},
		{
			name:        "filter out all data",
			csvInput:    "a,b\nc,d",
			withHeader:  false,
			fieldPerRow: 2,
			parser: func(data []string) (TestModel, error) {
				return TestModel{data[0], data[1]}, nil
			},
			filter: func(data TestModel) bool {
				return false
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:        "filter out specific data",
			csvInput:    "a,b\nc,d",
			withHeader:  false,
			fieldPerRow: 2,
			parser: func(data []string) (TestModel, error) {
				return TestModel{data[0], data[1]}, nil
			},
			filter: func(data TestModel) bool {
				return data.A == "a"
			},
			want:    []TestModel{{"a", "b"}},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer := bytes.NewBufferString(test.csvInput)
			p := csvparser.NewCSVParser(buffer, test.parser, test.filter, csvparser.CSVParserOptions{
				ContainsHeader: test.withHeader,
				FieldPerRow:    test.fieldPerRow,
			})
			got, err := p.Parse()
			if test.wantErr != (err != nil) {
				t.Errorf("wantErr is %t, but err is %v", test.wantErr, err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Parse() mismatch, (-want,+got):\n%s", diff)
			}
		})
	}
}
