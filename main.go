package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rickyson96/amartha-reconciliation-service/internal/parsers/statements"
	"github.com/rickyson96/amartha-reconciliation-service/internal/processes/reconciliation"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage of %s [options] {transaction file} {statement files} {start date} {end date}\n", os.Args[0])
	fmt.Fprintf(w, "Args:\n")
	fmt.Fprintf(w, "  transaction file: The system's transaction csv file to be reconciled\n")
	fmt.Fprintf(w, "  statement files: The bank's statement csv files to be reconciled, accept comma separated value. e.g.: bank1.csv,bank2.csv\n")
	fmt.Fprintf(w, "  start date: The reconciliate start date. e.g.: 2025-01-02\n")
	fmt.Fprintf(w, "  end date: The reconciliate end date. e.g.: 2025-12-31\n")
	fmt.Fprintf(w, "Options:\n")
	flag.PrintDefaults()
}

func fatalWithUsage(format string, msg ...any) {
	fmt.Printf(format+"\n\n", msg...)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	help := flag.Bool("h", false, "show help")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if len(os.Args) != 5 {
		fatalWithUsage("ERROR: Need exactly 4 arguments!")
	}

	transactionFile := os.Args[1]
	statementFileArg := os.Args[2]
	statementFiles := strings.Split(statementFileArg, ",")
	startDateArg := os.Args[3]
	endDateArg := os.Args[4]

	startDate, err := time.Parse(time.DateOnly, startDateArg)
	if err != nil {
		fatalWithUsage("ERROR: start date wrong format: %v", err)
	}

	endDate, err := time.Parse(time.DateOnly, endDateArg)
	if err != nil {
		fatalWithUsage("ERROR: end date wrong format: %v", err)
	}

	trxs, err := parseTransactions(transactionFile, startDate, endDate)
	if err != nil {
		log.Fatalf("parseTransactions error: %v", err)
	}

	stmts, err := parseStatements(statementFiles, startDate, endDate)
	if err != nil {
		log.Fatalf("parseTransactions error: %v", err)
	}

	result := reconciliation.Process(trxs, stmts)

	printReconciliation(result)
}

func printReconciliation(result reconciliation.Result) {
	unmatchedCount := result.Processed - result.Match

	fmt.Printf("Processed Transactions: %d\n", result.Processed)
	fmt.Printf("Matched Transactions: %d\n", result.Match)
	fmt.Printf("Unmatched Transactions: %d\n", result.Processed-result.Match)

	if unmatchedCount == 0 {
		return
	}

	fmt.Println("------------------")
	fmt.Println("Unmatched Details:")

	trxCount := len(result.Unmatched.Transactions)

	stmtCount := 0
	stmtIterator := maps.Values(result.Unmatched.Statements)
	stmtIterator(func(stmts []statements.Statement) bool {
		stmtCount += len(stmts)
		return true
	})

	if trxCount > 0 {
		w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
		fmt.Fprintf(w, "\nUnmatched Transactions: %d\n\n", trxCount)
		fmt.Fprintln(w, "\tTrxID\tType\tAmount\tTransactionTime")
		for _, t := range result.Unmatched.Transactions {
			fmt.Fprintf(w, "\t%v\t%v\t%v\t%v\n", t.TrxID, t.Type, t.Amount, t.TransactionTime)
		}
		w.Flush()
	}

	if stmtCount > 0 {
		w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
		fmt.Fprintf(w, "\nUnmatched Statements: %d\n\n", stmtCount)
		fmt.Fprintln(w, "\tUniqueIdentifier\tAmount\tDate")
		stmtIterator(func(stmts []statements.Statement) bool {
			for _, s := range stmts {
				fmt.Fprintf(w, "\t%v\t%v\t%v\n", s.UniqueIdentifier, s.Amount, s.Date)
			}
			return true
		})
		w.Flush()
	}
}
