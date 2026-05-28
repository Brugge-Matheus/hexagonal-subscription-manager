package usecases

import (
	"errors"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type GetSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (g GetSubscription) Execute(id string) (entities.Subscription, error) {
	subscription, err := g.SubscriptionRepository.FindByID(id)
	if err != nil {
		return entities.Subscription{}, errors.Join(ErrSubscriptionNotFound, err)
	}

	return subscription, nil
}
