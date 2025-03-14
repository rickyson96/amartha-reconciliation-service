package csvparser

import (
	"encoding/csv"
	"io"
)

type CSVParser[T any] struct {
	csvReader *csv.Reader
	parser    func(data []string) (T, error)
	filter    func(data T) bool
	hasHeader bool
}

type CSVParserOptions struct {
	ContainsHeader bool
	FieldPerRow    int
}

// NewCSVParser creates new csv_parser that helps parses into struct
// and filters unwanted data
func NewCSVParser[T any](csvFile io.Reader,
	parser func(data []string) (T, error),
	filter func(data T) bool,
	options CSVParserOptions,
) *CSVParser[T] {
	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = options.FieldPerRow

	csvParser := &CSVParser[T]{
		csvReader: reader,
		parser:    parser,
		filter:    filter,
		hasHeader: options.ContainsHeader,
	}

	return csvParser
}

// Parse generates output based on the CSV being read.
func (p *CSVParser[T]) Parse() ([]T, error) {
	var result []T

	if p.hasHeader {
		// Read off the header
		p.csvReader.Read()
	}

	for {
		data, err := p.csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		parsedData, err := p.parser(data)
		if err != nil {
			return nil, err
		}

		if p.filter(parsedData) {
			result = append(result, parsedData)
		}
	}

	return result, nil
}
