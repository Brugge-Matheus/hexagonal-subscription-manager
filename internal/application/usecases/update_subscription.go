package usecases

import (
	"errors"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type UpdateSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (u UpdateSubscription) Execute(id, customerID, planID, status string) (entities.Subscription, error) {
	_, err := u.SubscriptionRepository.FindByID(id)
	if err != nil {
		return entities.Subscription{}, errors.Join(ErrSubscriptionNotFound, err)
	}

	subscription := entities.Subscription{
		ID:         id,
		CustomerID: customerID,
		PlanID:     planID,
		Status:     status,
	}

	if err := u.SubscriptionRepository.Save(subscription); err != nil {
		return entities.Subscription{}, err
	}

	return subscription, nil
}
