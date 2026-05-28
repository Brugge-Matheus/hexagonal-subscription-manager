// Package ports
package ports

import "subscription-manager/internal/domain/entities"

type SubscriptionRepository interface {
	Save(subscription entities.Subscription) error
	FindByID(id string) (entities.Subscription, error)
	All() ([]entities.Subscription, error)
	Delete(id string) error
}
