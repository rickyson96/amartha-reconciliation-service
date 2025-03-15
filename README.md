# Introduction

This is the implementation of Reconciliation Service, with the goal is to identify unmatched and discrepant transactions between transactions and statements.

# Data Model

Transaction:
- `trxID` : Unique identifier for the transaction (string)
- `amount` : Transaction amount (decimal)
- `type` : Transaction type (enum: DEBIT, CREDIT)
- `transactionTime` : Date and time of the transaction (datetime)

Statement:
- `unique_identifier` : Unique identifier for the transaction in the bank statement (string) (varies by bank, not necessarily equivalent to `trxID` )
- `amount` : Transaction amount (decimal) (can be negative for debits)
- `date` : Date of the transaction (date)

# Assumptions

Glossary:
- transactions: System's transactions
- statements: Bank Statements

Premade assumptions on the issues are:
- Both input are from CSV files.
- The discrepancies only occurs in amount
- transactions' trxID is not statements' unique_identifier
- transactions will all be in one CSV
- statements can come in multiple CSVs

My additional assumptions:
- statement doesn't include time (only date)
- timezone doesn't matter (all date data and operation will be on the same timezone)
- time would be formatted in `yyyy-mm-dd hh:MM:ss`, while date will be formatted in `yyyy-mm-dd`
- the csv format follows the data model's ordering
- the csv contains header
- uniqueness can be defined by `date`+`amount`+`type`.
- in case of multiple transaction with the same uniqueness, we will just assume that it's the same transaction, and report the last read transactions as the discrepancies if any was found.
- the app's interface would be on CLI, with the need to provide exactly 4 arguments

# Implementation details

I've put all the code in internal part, and top level files are the glue files and input parsers.
The directory structure looks like this:
```
├── README.md
├── go.mod
├── main.go
├── internal
│   ├── csv_parser...
│   ├── parsers
│   │   ├── statements...
│   │   └── transactions...
│   ├── processes
│   │   └── reconciliation...
│   └── testutils...
├── process_concurrent.go
├── process.go
└── transactions.csv
```

The corresponding parsers are located in subdirectory `parsers`, and processors are located in `processes`. `csv_parser` are the helper struct to parse CSV.

I build it this way so that we can add more `parsers` along the way when we need to parse other data for the processes. It also supports for adding different processors, if we end up needing to put other type of processor than reconciliation.

For the reconciliation processor, I've also build two types of implementation, the synchronous and the concurrent one.

1. The synchronous version
For the synchronous version, the codebase is definitely simpler and easier to work with. I think it's suitable when we don't have a really big files to process. One downside of this approach is that it will read all of the files into the memory, which will have a big memory usage according to the file size.

We try to limit this by using filters, but if we need to reconcile big chunks of data, we still will run into the high memory usage.

2. The concurrent version
For the concurrent version, the codebase is more complex, since it needs to manage multiple goroutines and uses channel to communicate with each other. This approach is useful if we want to process a very big CSV files. It lowers the memory usage by quite a bunch.

This is achieved by running multiple goroutines to read and write. 
The reader can ingest both CSV concurrently.
The writer can parse the result directly and drop the read data, thus lowering the memory usage.

This is proven on the benchmark, though we need to take the benchmark with a grain of salt, since it generates the same data everytime, the writer can throw away the read data, making the allocations very low.
Theoretically, it can still lowers the memory allocations by half on the worst case.

Here's the benchmark result:
```
BenchmarkProcess10
BenchmarkProcess10-12                     100389             11355 ns/op            8556 B/op        213 allocs/op
BenchmarkProcessConcurrent10
BenchmarkProcessConcurrent10-12           645346              1900 ns/op             944 B/op         19 allocs/op
BenchmarkProcess100
BenchmarkProcess100-12                     10693            110647 ns/op           77782 B/op       2019 allocs/op
BenchmarkProcessConcurrent100
BenchmarkProcessConcurrent100-12          607368              1879 ns/op             944 B/op         19 allocs/op
BenchmarkProcess1000
BenchmarkProcess1000-12                     1004           1207086 ns/op          777349 B/op      20032 allocs/op
BenchmarkProcessConcurrent1000
BenchmarkProcessConcurrent1000-12         604436              1983 ns/op             945 B/op         19 allocs/op
BenchmarkProcess10000
BenchmarkProcess10000-12                      88          13561066 ns/op        10189276 B/op     200115 allocs/op
BenchmarkProcessConcurrent10000
BenchmarkProcessConcurrent10000-12        551257              1993 ns/op             954 B/op         19 allocs/op
BenchmarkProcess100000
BenchmarkProcess100000-12                      8         129620848 ns/op        110421067 B/op   2000589 allocs/op
BenchmarkProcessConcurrent100000
BenchmarkProcessConcurrent100000-12       496756              2089 ns/op            1057 B/op         23 allocs/op
```

