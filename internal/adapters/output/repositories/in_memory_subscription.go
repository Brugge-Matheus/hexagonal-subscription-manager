package repositories

import (
	"sync"

	"subscription-manager/internal/domain/entities"
)

type InMemorySubscription struct {
	mu            sync.RWMutex
	subscriptions map[string]entities.Subscription
}

func NewInMemorySubscription() *InMemorySubscription {
	return &InMemorySubscription{
		subscriptions: make(map[string]entities.Subscription),
	}
}

func (r *InMemorySubscription) Save(subscription entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscriptions[subscription.ID] = subscription
	return nil
}

func (r *InMemorySubscription) FindByID(id string) (entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	subscription, ok := r.subscriptions[id]
	if !ok {
		return entities.Subscription{}, ErrSubscriptionNotFound
	}
	return subscription, nil
}

func (r *InMemorySubscription) All() ([]entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	subscriptions := make([]entities.Subscription, 0, len(r.subscriptions))
	for _, subscription := range r.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (r *InMemorySubscription) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.subscriptions[id]; !ok {
		return ErrSubscriptionNotFound
	}
	delete(r.subscriptions, id)
	return nil
}
