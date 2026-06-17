package usecases

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
)
