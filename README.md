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

