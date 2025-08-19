package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/alert"
	"fraud-detection/pkg/db"
	"fraud-detection/pkg/fraud"
	"fraud-detection/pkg/location"
	"fraud-detection/pkg/logger"
	"fraud-detection/pkg/mlservice"
	"fraud-detection/pkg/routing"
	"fraud-detection/pkg/transaction"
	"fraud-detection/repository"
	"log"
	"net/http"
	"time"
)

var conn, err = db.Connect()
var appLogger = logger.New()
var router = routing.NewRouter()
var mlClient = mlservice.NewMLClient()

func main() {
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to PostgreSQL")

	CreateTable()

	// handle the requests
	router.HandleFunc("/transaction", CreateAndHandleTransaction).Methods("POST")
	router.HandleFunc("/transactions", GetTransactions).Methods("GET")
	router.HandleFunc("/location/transactions", GetAccountLocation).Methods("GET")
	router.HandleFunc("/account/transactions", GetTransactionsByAccount).Methods("GET")
	// show info
	appLogger.Info("Fraud detection system started")
	// show status from server
	appLogger.Fatal(http.ListenAndServe(":8080", router))
}

func CreateTable() {
	_, err = conn.Exec(context.Background(), `
		DROP TABLE IF EXISTS transactions;
		CREATE TABLE IF NOT EXISTS transactions (
            transaction_id VARCHAR(255) PRIMARY KEY,
            amount DECIMAL(10,2) NOT NULL,
            account_id VARCHAR(255) NOT NULL,
            location VARCHAR(255) NOT NULL,
            transaction_time TIMESTAMP NOT NULL,
            elapsed_time FLOAT NOT NULL,
            frequency INTEGER NOT NULL,
            fraud_label INTEGER NOT NULL,
            CONSTRAINT chk_fraud_label CHECK (fraud_label IN (0, 1))
        );

        CREATE INDEX IF NOT EXISTS idx_account_id ON transactions(account_id);
        CREATE INDEX IF NOT EXISTS idx_transaction_time ON transactions(transaction_time);
    `)

	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func CreateAndHandleTransaction(writer http.ResponseWriter, request *http.Request) {
	appLogger.Info("Incoming transaction request received")
	writer.Header().Set("Content-Type", "application/json")

	var transactionBody transaction.Transaction
	var transactionModel transaction.Transaction

	if err := json.NewDecoder(request.Body).Decode(&transactionBody); err != nil {
		appLogger.Error("Failed to decode transaction request body", "error", err)
		http.Error(writer, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	transactions, err := repository.GetTransactionsByAccountID(context.Background(), conn, transactionBody.AccountID)
	if err != nil {
		appLogger.Error("Failed to retrieve account transactions",
			"account_id", transactionBody.AccountID,
			"error", err,
		)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	var elapsedTime float64 = 0
	var frequency int = 0
	now := time.Now().UTC()
	dayAgo := now.Add(-24 * time.Hour)

	if len(transactions) > 0 {
		lastTransaction := transactions[len(transactions)-1]
		elapsedTime = time.Since(lastTransaction.TransactionTime).Hours()

		for _, t := range transactions {
			if t.TransactionTime.After(dayAgo) {
				frequency++
			}
		}
	}

	transactionModel = transaction.NewTransaction(
		transactionBody.Amount,
		transactionBody.AccountID,
		transactionBody.Location,
		elapsedTime,
		frequency,
	)

	appLogger.Info("Processing transaction",
		"account_id", transactionModel.AccountID,
		"amount", transactionModel.Amount,
	)

	//lastLocation, err := GetAccountLocationWithId(transactionModel.AccountID)
	if err != nil {
		appLogger.Error("Failed to retrieve account location",
			"account_id", transactionModel.AccountID,
			"error", err,
		)
	}

	current := transactionModel.Location
	//current, err := geoip.GetGeoLocation(transactionModel.Location)
	//if err != nil {
	//	appLogger.Error("Failed to get geolocation",
	//		"location", transactionModel.Location,
	//		"error", err,
	//	)
	//}

	features := []float64{
		transactionModel.Amount,
		elapsedTime,
		float64(time.Now().Year()),
		float64(time.Now().Month()),
		float64(time.Now().Day()),
		float64(time.Now().Hour()),
		float64(time.Now().Minute()),
		float64(time.Now().Second()),
		float64(frequency),
		// Add other features as needed
	}

	mlPrediction, err := mlClient.PredictFraud(features)
	if err != nil {
		appLogger.Error("Failed to get ML prediction", "error", err)
		// Continue with local validations only
	}

	isLocalFraud := fraud.CheckTransactionForFraud(transactionModel, transactions, "", current, "DESC", frauddetector.Checker)
	isFraud := isLocalFraud || (mlPrediction != nil && mlPrediction.Fraud == 1)

	if isFraud {
		transactionModel.FraudLabel = transaction.LabelFraud
		appLogger.Warn("Transaction marked as fraudulent",
			"account_id", transactionModel.AccountID,
			"amount", transactionModel.Amount,
			"ml_prediction", mlPrediction != nil && mlPrediction.Fraud == 1,
		)
		alert.SendAlert(transactionModel)
	} else {
		transactionModel.FraudLabel = transaction.LabelLegit
	}

	_, err = repository.InsertTransaction(context.Background(), conn, transactionModel)
	if err != nil {
		log.Fatal("Insert error:", err)
	}

	appLogger.Info("Transaction saved successfully",
		"account_id", transactionModel.AccountID,
		"amount", transactionModel.Amount,
		"fraud_label", transactionModel.FraudLabel,
	)

	// Return the transaction with fraud status
	json.NewEncoder(writer).Encode(transactionModel)
}

// func InsertTransaction(transactionModel transaction.Transaction, writer http.ResponseWriter) {
// 	if err := repository.InsertTransaction(context.Background(), conn, transactionModel); err != nil {
// 		fmt.Println("❌ Failed to insert transaction:", err)
// 	} else {
// 		location, _ := geoip.GetGeoLocation(transactionModel.IPAddress)
// 		SaveAccountLocation(transactionModel, location)
// 		json.NewEncoder(writer).Encode(transactionModel)
// 		fmt.Println("✅ Transaction saved to database.")
// 	}
// }

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

// func SaveAccountLocation(transactionModel transaction.Transaction, location *geoip.LocationData) error {
// 	query := `
//         INSERT INTO account_location (account_id, country, city, latitude, longitude)
//         VALUES ($1, $2, $3, $4, $5)
//         RETURNING id, created_at, updated_at;
//     `

// 	row := conn.QueryRow(context.Background(), query, transactionModel.AccountID, location.Country, location.City, location.Latitude, location.Longitude)

// 	if err := row.Scan(&transactionModel.ID); err != nil {
// 		return fmt.Errorf("unable to save location data: %v", err)
// 	}

// 	log.Printf("Location saved for account %s, ID: %d", transactionModel.AccountID, transactionModel.ID)
// 	return nil
// }

func GetAccountLocationWithId(accountID string) (*location.AccountLocation, error) {
	query := `
		SELECT id, account_id, city
		FROM account_location
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`

	row := conn.QueryRow(context.Background(), query, accountID)

	var local location.AccountLocation
	err := row.Scan(&local.ID, &local.AccountID, &local.City)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // Return nil without error when no location is found
		}
		return nil, fmt.Errorf("unable to retrieve location data: %v", err)
	}

	appLogger.Info("Location retrieved for account", "account_id", local.AccountID)
	return &local, nil
}

func GetAccountLocation(writer http.ResponseWriter, request *http.Request) {
	//writer.Header().Set("Content-Type", "application/json")
	//
	//query := `
	//    SELECT id, account_id, country, city, latitude, longitude
	//    FROM account_location
	//    WHERE account_id = $1
	//    ORDER BY created_at DESC
	//    LIMIT 1;
	//`
	//
	//accountID := request.URL.Query().Get("account_id")
	//
	//row := conn.QueryRow(context.Background(), query, accountID)
	//
	//var local location.AccountLocation
	//err := row.Scan(&local.ID, &local.AccountID, &local.Country, &local.City, &local.Latitude, &local.Longitude)
	//if err != nil {
	//	if err.Error() == "no rows in result set" {
	//		writer.WriteHeader(http.StatusNotFound)
	//		json.NewEncoder(writer).Encode(map[string]string{"message": "No location found"})
	//		return
	//	}
	//	http.Error(writer, fmt.Sprintf("unable to retrieve location data: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//appLogger.Info("Location retrieved for account", "account_id", local.AccountID)
	//json.NewEncoder(writer).Encode(local)
}

func GetTransactionsByAccount(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	accountID := request.URL.Query().Get("account_id")
	if accountID == "" {
		http.Error(writer, "account_id is required", http.StatusBadRequest)
		return
	}

	appLogger.Info("Retrieving transactions", "account_id", accountID)

	transactions, err := repository.GetTransactionsByAccountID(context.Background(), conn, accountID)
	if err != nil {
		appLogger.Error("Failed to get transactions", "error", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(transactions) == 0 {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "No transactions found for this account",
		})
		return
	}

	json.NewEncoder(writer).Encode(transactions)
}
