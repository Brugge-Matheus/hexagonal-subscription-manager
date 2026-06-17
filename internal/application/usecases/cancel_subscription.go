package usecases

import (
	"errors"
	"time"

	"subscription-manager/internal/application/ports"
)

type CancelSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (c CancelSubscription) Execute(id string) (ports.CancellationResult, error) {
	subscription, err := c.SubscriptionRepository.FindByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ports.CancellationResult{}, ports.ErrSubscriptionNotFound
		}
		return ports.CancellationResult{}, err
	}

	refundEligible := subscription.CanRefundProportionally(time.Now())
	subscription.Cancel()

	if err := c.SubscriptionRepository.Save(subscription); err != nil {
		return ports.CancellationResult{}, err
	}

	return ports.CancellationResult{
		Subscription:   subscription,
		RefundEligible: refundEligible,
	}, nil
}
