package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/alert"
	"fraud-detection/pkg/db"
	"fraud-detection/pkg/fraud"
	"fraud-detection/pkg/logger"
	"fraud-detection/pkg/routing"
	"fraud-detection/pkg/transaction"
	"fraud-detection/repository"
	"log"
	"net/http"
	"os"
)

var conn, err = db.Connect()
var appLogger = logger.New()
var router = routing.NewRouter()

func main() {
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	fmt.Println("✅ Connected to PostgreSQL")

	CreateTable()

	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	// handle the requests
	router.HandleFunc("/transaction", CreateAndHandleTransaction).Methods("POST")
	router.HandleFunc("/transactions", GetTransactions).Methods("GET")
	// show info
	appLogger.Info("Fraud detection system started")
	// show status from server
	appLogger.Fatal(http.ListenAndServe(":8080", router))
}

func CreateTable() {
	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			account_id TEXT NOT NULL,
			amount NUMERIC NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			location TEXT,
			ip_address TEXT
		);
	`)
}

func CreateAndHandleTransaction(writer http.ResponseWriter, request *http.Request) {
	// set logger
	appLogger.Info("Request received")
	// set header
	writer.Header().Set("Content-Type", "application/json")
	// model for map body parameter
	var transactionBody transaction.Transaction
	var transactionModel transaction.Transaction
	// handle error
	if err := json.NewDecoder(request.Body).Decode(&transactionBody); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		json.NewEncoder(writer).Encode("Error decoding JSON")
		return
	}

	transactionModel = transaction.NewTransaction(
		transactionBody.Amount,
		transactionBody.AccountID,
		transactionBody.Location,
		transactionBody.IPAddress,
	)

	transactions, err := repository.GetTransactionsByAccountID(context.Background(), conn, transactionBody.AccountID)

	if err != nil {
		fmt.Println("❌ Failed to get transaction:", err)
	}

	// check if exists fraud on transaction
	if fraud.CheckTransactionForFraud(transactionModel, transactions, frauddetector.Checker) {
		// is suspicious
		appLogger.Info("Transaction suspicious")
		alert.SendAlert(transactionModel)
	} else {
		// no suspicious
		appLogger.Info("Transaction is valid!")
		InsertTransaction(transactionModel, writer)
	}
}

func InsertTransaction(transactionModel transaction.Transaction, writer http.ResponseWriter) {
	if err := repository.InsertTransaction(context.Background(), conn, transactionModel); err != nil {
		fmt.Println("❌ Failed to insert transaction:", err)
	} else {
		fmt.Println("✅ Transaction saved to database.")
		json.NewEncoder(writer).Encode(transactionModel)
	}
}

func GetTransactions(writer http.ResponseWriter, request *http.Request) {
	// set logger
	appLogger.Info("Request received")
	// set header
	writer.Header().Set("Content-Type", "application/json")

	transactions, err := repository.GetTransactions(context.Background(), conn)

	if err != nil {
		fmt.Println("❌ Failed to get transaction:", err)
	}

	if len(transactions) > 0 {
		json.NewEncoder(writer).Encode(transactions)
	} else {
		json.NewEncoder(writer).Encode("No transactions found")
	}
}
