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
	existing, err := u.SubscriptionRepository.FindByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return entities.Subscription{}, ports.ErrSubscriptionNotFound
		}
		return entities.Subscription{}, err
	}

	subscription := entities.Subscription{
		ID:         id,
		CustomerID: customerID,
		PlanID:     planID,
		Status:     status,
		CreatedAt:  existing.CreatedAt,
		CanceledAt: existing.CanceledAt,
	}

	if err := u.SubscriptionRepository.Save(subscription); err != nil {
		return entities.Subscription{}, err
	}

	return subscription, nil
}
