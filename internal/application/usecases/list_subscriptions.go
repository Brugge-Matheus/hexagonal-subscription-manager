package usecases

import (
	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type ListSubscriptions struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (l ListSubscriptions) Execute() ([]entities.Subscription, error) {
	return l.SubscriptionRepository.All()
}
