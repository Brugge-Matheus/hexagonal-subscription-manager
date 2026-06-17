package usecases

import (
	"errors"

	"subscription-manager/internal/application/ports"
)

type ProcessPaymentEvent struct {
	SubscriptionRepository ports.SubscriptionRepository
	NotificationService    ports.NotificationService
}

func (p ProcessPaymentEvent) Execute(subscriptionID, customerID, planID, status string) error {
	subscription, err := p.SubscriptionRepository.FindByID(subscriptionID)
	if err != nil {
		return errors.Join(ErrSubscriptionNotFound, err)
	}

	switch status {
	case "confirmed":
		subscription.Reactivate()
		p.notify(subscription.CustomerID, "payment_confirmed", "payment confirmed, subscription reactivated")
	case "refused":
		subscription.Suspend()
		p.notify(subscription.CustomerID, "payment_refused", "payment refused, subscription suspended")
	default:
		return ErrInvalidPaymentStatus
	}

	return p.SubscriptionRepository.Save(subscription)
}

func (p ProcessPaymentEvent) notify(customerID, event, message string) {
	if p.NotificationService != nil {
		_ = p.NotificationService.Notify(customerID, event, message)
	}
}
