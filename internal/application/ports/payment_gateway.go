package ports

type PaymentGateway interface {
	Charge(subscriptionID string, amount float64) (transactionID string, err error)
	Refund(transactionID string) error
}
