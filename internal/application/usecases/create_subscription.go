package usecases

import (
	"fmt"
	"strconv"
	"time"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type CreateSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
	PaymentGateway         ports.PaymentGateway
}

func (c CreateSubscription) Execute(customerID, planID string) (entities.Subscription, error) {
	subscription := entities.Subscription{
		ID:         strconv.FormatInt(time.Now().UnixNano(), 10),
		CustomerID: customerID,
		PlanID:     planID,
		Status:     entities.StatusActive,
		CreatedAt:  time.Now(),
	}

	var txID string
	if c.PaymentGateway != nil {
		var err error
		txID, err = c.PaymentGateway.Charge(subscription.ID, 0)
		if err != nil {
			return entities.Subscription{}, err
		}
	}

	if err := c.SubscriptionRepository.Save(subscription); err != nil {
		if txID != "" {
			if refundErr := c.PaymentGateway.Refund(txID); refundErr != nil {
				return entities.Subscription{}, fmt.Errorf("save: %w; refund also failed: %v", err, refundErr)
			}
		}
		return entities.Subscription{}, err
	}

	return subscription, nil
}
