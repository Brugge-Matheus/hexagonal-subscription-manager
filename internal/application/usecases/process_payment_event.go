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
		if errors.Is(err, ports.ErrNotFound) {
			return ports.ErrSubscriptionNotFound
		}
		return err
	}

	prevStatus := subscription.Status
	switch status {
	case "confirmed":
		subscription.Reactivate()
	case "refused":
		subscription.Suspend()
	default:
		return ErrInvalidPaymentStatus
	}

	if subscription.Status == prevStatus {
		return nil
	}

	if err := p.SubscriptionRepository.Save(subscription); err != nil {
		return err
	}

	switch status {
	case "confirmed":
		p.notify(subscription.CustomerID, "payment_confirmed", "payment confirmed, subscription reactivated")
	case "refused":
		p.notify(subscription.CustomerID, "payment_refused", "payment refused, subscription suspended")
	}
	return nil
}

func (p ProcessPaymentEvent) notify(customerID, event, message string) {
	if p.NotificationService != nil {
		_ = p.NotificationService.Notify(customerID, event, message)
	}
}
