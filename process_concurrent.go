package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	csvparser "github.com/rickyson96/amartha-reconciliation-service/internal/csv_parser"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
	"github.com/rickyson96/amartha-reconciliation-service/internal/processes/reconciliation"
)

type statementReader struct {
	filesWithReader map[string]*csvparser.CSVParser[statements.Statement]
	files           []string
	osFiles         []*os.File
	readCount       int
}

func newStatementReader(files []string, startDate, endDate time.Time) (*statementReader, error) {
	reader := statementReader{
		filesWithReader: map[string]*csvparser.CSVParser[statements.Statement]{},
		files:           files,
		osFiles:         []*os.File{},
		readCount:       0,
	}

	for _, stmtFile := range files {
		stmtFilePath, err := filepath.Abs(stmtFile)
		if err != nil {
			return nil, err
		}
		file, err := os.Open(stmtFilePath)
		if err != nil {
			return nil, err
		}
		reader.osFiles = append(reader.osFiles, file)

		stmtParser := statements.NewCSVParser(file, startDate, endDate)
		reader.filesWithReader[stmtFile] = stmtParser
	}

	return &reader, nil
}

func (r *statementReader) Read() (reconciliation.StatementFilePair, error) {
	if len(r.files) == r.readCount {
		return reconciliation.StatementFilePair{}, io.EOF
	}

	currentFile := r.files[r.readCount]
	currentReader := r.filesWithReader[currentFile]

	statement, err := currentReader.Read()
	if err == io.EOF {
		r.readCount++
		return r.Read()
	}
	if err != nil {
		return reconciliation.StatementFilePair{}, io.EOF
	}

	return reconciliation.StatementFilePair{
		Name:      currentFile,
		Statement: statement,
	}, nil
}

func (r *statementReader) Close() error {
	var errs []error
	for _, f := range r.osFiles {
		err := f.Close()
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func processConcurrent(transactionFile string, statementFiles []string, startDate, endDate time.Time) (reconciliation.Result, error) {
	trxFilePath, err := filepath.Abs(transactionFile)
	if err != nil {
		return reconciliation.Result{}, err
	}
	trxFile, err := os.Open(trxFilePath)
	if err != nil {
		return reconciliation.Result{}, err
	}
	defer trxFile.Close()

	transactionParser := transactions.NewCSVParser(trxFile, startDate, endDate)

	statementParser, err := newStatementReader(statementFiles, startDate, endDate)
	if err != nil {
		return reconciliation.Result{}, err
	}
	defer statementParser.Close()

	return reconciliation.ProcessConcurrent(transactionParser.Read, statementParser.Read)
}
