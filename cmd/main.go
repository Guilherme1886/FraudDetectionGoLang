package main

import (
	"encoding/json"
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/alert"
	"fraud-detection/pkg/fraud"
	"fraud-detection/pkg/transaction"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Create new log
var log = logrus.New()

// MyTransactions keep on memory at least
var MyTransactions []transaction.Transaction

func main() {
	// create new router
	r := mux.NewRouter()
	// setup log level
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	// handle the requests
	r.HandleFunc("/transaction", CreateAndHandleTransaction).Methods("POST")
	r.HandleFunc("/transactions", GetTransactions).Methods("GET")
	// show info
	log.Info("Fraud detection system started")
	// show status from server
	log.Fatal(http.ListenAndServe(":8080", r))
}

func CreateAndHandleTransaction(writer http.ResponseWriter, request *http.Request) {
	// set log
	log.Info("Request received")
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
		transactionBody.ID,
		transactionBody.Amount,
		transactionBody.AccountID,
		transactionBody.Location,
		transactionBody.IPAddress,
	)

	// check if exists fraud on transaction
	if fraud.CheckTransactionForFraud(transactionModel, MyTransactions, frauddetector.Checker) {
		// is suspicious
		alert.SendAlert(transactionModel)
		log.Info("Transaction suspicious")
	} else {
		// no suspicious
		log.Info("Transaction is valid!")
		MyTransactions = append(MyTransactions, transactionModel)
		json.NewEncoder(writer).Encode(transactionModel)
	}
}

func GetTransactions(writer http.ResponseWriter, request *http.Request) {
	// set log
	log.Info("Request received")
	// set header
	writer.Header().Set("Content-Type", "application/json")

	if len(MyTransactions) > 0 {
		json.NewEncoder(writer).Encode(MyTransactions)
	} else {
		json.NewEncoder(writer).Encode("No transactions found")
	}
}
