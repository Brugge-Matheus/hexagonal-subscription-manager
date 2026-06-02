package usecases

import (
	"errors"
	"time"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type CancelSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

type CancelSubscriptionResult struct {
	Subscription   entities.Subscription
	RefundEligible bool
}

func (c CancelSubscription) Execute(id string) (CancelSubscriptionResult, error) {
	subscription, err := c.SubscriptionRepository.FindByID(id)
	if err != nil {
		return CancelSubscriptionResult{}, errors.Join(ErrSubscriptionNotFound, err)
	}

	refundEligible := subscription.CanRefundProportionally(time.Now())
	subscription.Cancel()

	if err := c.SubscriptionRepository.Save(subscription); err != nil {
		return CancelSubscriptionResult{}, err
	}

	return CancelSubscriptionResult{
		Subscription:   subscription,
		RefundEligible: refundEligible,
	}, nil
}
