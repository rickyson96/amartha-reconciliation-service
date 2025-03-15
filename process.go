package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/transactions"
	"github.com/rickyson96/amartha-reconciliation-service/internal/processes/reconciliation"
)

func parseTransactions(fileName string, startDate, endDate time.Time) ([]transactions.Transaction, error) {
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	transactionParser := transactions.NewCSVParser(file, startDate, endDate)

	return transactionParser.Parse()
}

func parseStatements(files []string, startDate, endDate time.Time) (map[string][]statements.Statement, error) {
	statementsMap := make(map[string][]statements.Statement)
	for _, fileName := range files {
		filePath, err := filepath.Abs(fileName)
		if err != nil {
			return nil, err
		}
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		statementParser := statements.NewCSVParser(file, startDate, endDate)

		stmts, err := statementParser.Parse()
		if err != nil {
			return nil, err
		}

		statementsMap[fileName] = stmts
	}

	return statementsMap, nil
}

func process(transactionFile string, statementFiles []string, startDate, endDate time.Time) (reconciliation.Result, error) {
	trxs, err := parseTransactions(transactionFile, startDate, endDate)
	if err != nil {
		return reconciliation.Result{}, err
	}

	stmts, err := parseStatements(statementFiles, startDate, endDate)
	if err != nil {
		return reconciliation.Result{}, err
	}

	return reconciliation.Process(trxs, stmts), nil
}
