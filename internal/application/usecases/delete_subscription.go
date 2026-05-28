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
		return errors.Join(ErrSubscriptionNotFound, err)
	}

	return d.SubscriptionRepository.Delete(id)
}
