package alert

import (
	"fraud-detection/pkg/transaction"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func SendAlert(t transaction.Transaction) {
	log.Info("ALERT: Fraudulent activity detected in transaction ID: %s\n", t.ID)
}
