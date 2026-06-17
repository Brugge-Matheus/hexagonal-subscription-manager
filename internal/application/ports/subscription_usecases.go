package ports

import "subscription-manager/internal/domain/entities"

// CancellationResult is the output of a subscription cancellation.
type CancellationResult struct {
	Subscription   entities.Subscription
	RefundEligible bool
}

type SubscribeUseCase interface {
	Execute(customerID, planID string) (entities.Subscription, error)
}

type ListSubscriptionsUseCase interface {
	Execute() ([]entities.Subscription, error)
}

type GetSubscriptionUseCase interface {
	Execute(id string) (entities.Subscription, error)
}

type UpdateSubscriptionUseCase interface {
	Execute(id, customerID, planID, status string) (entities.Subscription, error)
}

type CancelSubscriptionUseCase interface {
	Execute(subscriptionID string) (CancellationResult, error)
}

type ReactivateSubscriptionUseCase interface {
	Execute(id string) (entities.Subscription, error)
}

type DeleteSubscriptionUseCase interface {
	Execute(id string) error
}

type ProcessPaymentEventUseCase interface {
	Execute(subscriptionID, customerID, planID, status string) error
}
