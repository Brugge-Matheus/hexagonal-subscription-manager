// Package repositories
package repositories

import "subscription-manager/internal/domain/entities"

type InMemorySubscription struct {
	subscriptions map[string]entities.Subscription
}

func NewInMemorySubscription() InMemorySubscription {
	return InMemorySubscription{
		subscriptions: make(map[string]entities.Subscription),
	}
}

func (r InMemorySubscription) Save(subscription entities.Subscription) error {
	r.subscriptions[subscription.ID] = subscription
	return nil
}

func (r InMemorySubscription) FindByID(id string) (entities.Subscription, error) {
	subscription, ok := r.subscriptions[id]
	if !ok {
		return entities.Subscription{}, ErrSubscriptionNotFound
	}

	return subscription, nil
}

func (r InMemorySubscription) All() ([]entities.Subscription, error) {
	subscriptions := make([]entities.Subscription, 0, len(r.subscriptions))
	for _, subscription := range r.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (r InMemorySubscription) Delete(id string) error {
	if _, ok := r.subscriptions[id]; !ok {
		return ErrSubscriptionNotFound
	}

	delete(r.subscriptions, id)
	return nil
}
