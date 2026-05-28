// Package usecases
package usecases

import (
	"strconv"
	"time"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type CreateSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (c CreateSubscription) Execute(customerID, planID string) (entities.Subscription, error) {
	subscription := entities.Subscription{
		ID: strconv.FormatInt(time.Now().UnixNano(), 10),
		CustomerID: customerID,
		PlanID: planID,
		Status: "active",
	}

	if err := c.SubscriptionRepository.Save(subscription); err != nil {
		return entities.Subscription{}, err
	}

	return subscription, nil
}
