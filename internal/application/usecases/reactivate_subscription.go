package usecases

import (
	"errors"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type ReactivateSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (r ReactivateSubscription) Execute(id string) (entities.Subscription, error) {
	subscription, err := r.SubscriptionRepository.FindByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return entities.Subscription{}, ports.ErrSubscriptionNotFound
		}
		return entities.Subscription{}, err
	}

	if !subscription.IsSuspended() {
		return entities.Subscription{}, ports.ErrSubscriptionNotSuspended
	}

	subscription.Reactivate()

	if err := r.SubscriptionRepository.Save(subscription); err != nil {
		return entities.Subscription{}, err
	}

	return subscription, nil
}
