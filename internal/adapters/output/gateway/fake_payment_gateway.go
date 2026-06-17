package gateway

import (
	"fmt"
	"log"
)

type FakePaymentGateway struct{}

func (f FakePaymentGateway) Charge(subscriptionID string, amount float64) (string, error) {
	transactionID := fmt.Sprintf("fake_txn_%s", subscriptionID)
	log.Printf("[gateway] charge initiated: subscription=%s amount=%.2f transaction=%s", subscriptionID, amount, transactionID)
	return transactionID, nil
}

func (f FakePaymentGateway) Refund(transactionID string) error {
	log.Printf("[gateway] refund initiated: transaction=%s", transactionID)
	return nil
}
