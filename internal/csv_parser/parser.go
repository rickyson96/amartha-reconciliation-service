package csvparser

import (
	"encoding/csv"
	"io"
)

type CSVParser[T any] struct {
	csvReader     *csv.Reader
	parser        func(data []string) (T, error)
	filter        func(data T) bool
	hasHeader     bool
	hasReadHeader bool
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
		p.hasReadHeader = true
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

// Read is used to read through the csv file one line at a time.
//
// This mainly utilizes underlying [csv.Reader.Read] method, while
// helping to parse data using the parser function, and filter the
// data usig the filter function. This is useful for streaming csv
// data.
//
// It will try to skip ahead when the data is being filtered
// It'll also skip the header according to [CSVParser.hasHeader].
// It'll return [io.EOF] error when it reaches the last input
func (p *CSVParser[T]) Read() (T, error) {
	if p.hasHeader && !p.hasReadHeader {
		p.csvReader.Read()
		p.hasReadHeader = true
	}

	data, err := p.csvReader.Read()
	if err != nil {
		return *new(T), err
	}

	parsed, err := p.parser(data)
	if err != nil {
		return *new(T), err
	}

	if !p.filter(parsed) {
		return p.Read()
	}
	return parsed, nil
}
