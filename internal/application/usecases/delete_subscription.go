package usecases

import (
	"errors"

	"subscription-manager/internal/application/ports"
)

type DeleteSubscription struct {
	SubscriptionRepository ports.SubscriptionRepository
}

func (d DeleteSubscription) Execute(id string) error {
	_, err := d.SubscriptionRepository.FindByID(id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ports.ErrSubscriptionNotFound
		}
		return err
	}

	return d.SubscriptionRepository.Delete(id)
}
