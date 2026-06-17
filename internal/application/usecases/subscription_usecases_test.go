package usecases

import (
	"errors"
	"testing"
	"time"

	"subscription-manager/internal/adapters/output/repositories"
	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type spyGateway struct {
	chargedFor []string
}

func (s *spyGateway) Charge(subscriptionID string, amount float64) (string, error) {
	s.chargedFor = append(s.chargedFor, subscriptionID)
	return "txn_" + subscriptionID, nil
}

func (s *spyGateway) Refund(_ string) error { return nil }

func TestCreateSubscriptionCallsGateway(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	spy := &spyGateway{}
	uc := CreateSubscription{
		SubscriptionRepository: repository,
		PaymentGateway:         spy,
	}

	created, err := uc.Execute("customer-1", "basic-plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.chargedFor) != 1 {
		t.Fatalf("expected gateway to be called once, called %d times", len(spy.chargedFor))
	}

	if spy.chargedFor[0] != created.ID {
		t.Fatalf("expected gateway called with %s, got %s", created.ID, spy.chargedFor[0])
	}
}

func TestSubscriptionUseCasesCRUD(t *testing.T) {
	repository := repositories.NewInMemorySubscription()

	createUseCase := CreateSubscription{SubscriptionRepository: repository}
	listUseCase := ListSubscriptions{SubscriptionRepository: repository}
	getUseCase := GetSubscription{SubscriptionRepository: repository}
	updateUseCase := UpdateSubscription{SubscriptionRepository: repository}
	cancelUseCase := CancelSubscription{SubscriptionRepository: repository}
	deleteUseCase := DeleteSubscription{SubscriptionRepository: repository}

	created, err := createUseCase.Execute("customer-1", "basic-plan")
	if err != nil {
		t.Fatalf("unexpected error creating subscription: %v", err)
	}

	if created.Status != "active" {
		t.Fatalf("expected active status, got %s", created.Status)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be filled")
	}

	listed, err := listUseCase.Execute()
	if err != nil {
		t.Fatalf("unexpected error listing subscriptions: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(listed))
	}

	found, err := getUseCase.Execute(created.ID)
	if err != nil {
		t.Fatalf("unexpected error getting subscription: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, found.ID)
	}

	updated, err := updateUseCase.Execute(created.ID, "customer-2", "premium-plan", "suspended")
	if err != nil {
		t.Fatalf("unexpected error updating subscription: %v", err)
	}

	if updated.Status != "suspended" {
		t.Fatalf("expected suspended status, got %s", updated.Status)
	}

	canceled, err := cancelUseCase.Execute(created.ID)
	if err != nil {
		t.Fatalf("unexpected error canceling subscription: %v", err)
	}

	if !canceled.Subscription.IsCanceled() {
		t.Fatal("expected subscription to be canceled")
	}

	if canceled.Subscription.Status != entities.StatusCanceled {
		t.Fatalf("expected canceled status, got %s", canceled.Subscription.Status)
	}

	if !canceled.RefundEligible {
		t.Fatal("expected refund eligibility for recent subscription")
	}

	if err := deleteUseCase.Execute(created.ID); err != nil {
		t.Fatalf("unexpected error deleting subscription: %v", err)
	}

	if _, err := getUseCase.Execute(created.ID); err == nil {
		t.Fatal("expected error when reading deleted subscription")
	}
}

func TestCancelSubscriptionWithoutRefundAfterSevenDays(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	cancelUseCase := CancelSubscription{SubscriptionRepository: repository}

	createdAt := time.Now().Add(-8 * 24 * time.Hour)
	subscription := entities.Subscription{
		ID:         "sub-old",
		CustomerID: "customer-1",
		PlanID:     "basic-plan",
		Status:     entities.StatusActive,
		CreatedAt:  createdAt,
	}

	if err := repository.Save(subscription); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	result, err := cancelUseCase.Execute(subscription.ID)
	if err != nil {
		t.Fatalf("unexpected error canceling old subscription: %v", err)
	}

	if result.RefundEligible {
		t.Fatal("expected no refund for subscription older than seven days")
	}
}

func TestReactivateSubscription(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ReactivateSubscription{SubscriptionRepository: repository}

	sub := entities.Subscription{
		ID:        "sub-reactivate-1",
		Status:    entities.StatusSuspended,
		CreatedAt: time.Now(),
	}
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	result, err := uc.Execute(sub.ID)
	if err != nil {
		t.Fatalf("unexpected error reactivating subscription: %v", err)
	}

	if result.Status != entities.StatusActive {
		t.Fatalf("expected active status, got %s", result.Status)
	}
}

func TestReactivateSubscriptionNotSuspended(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ReactivateSubscription{SubscriptionRepository: repository}

	sub := entities.Subscription{
		ID:        "sub-reactivate-2",
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
	}
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	_, err := uc.Execute(sub.ID)
	if !errors.Is(err, ports.ErrSubscriptionNotSuspended) {
		t.Fatalf("expected ports.ErrSubscriptionNotSuspended, got %v", err)
	}
}

func TestReactivateSubscriptionNotFound(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ReactivateSubscription{SubscriptionRepository: repository}

	_, err := uc.Execute("nonexistent")
	if !errors.Is(err, ports.ErrSubscriptionNotFound) {
		t.Fatalf("expected ports.ErrSubscriptionNotFound, got %v", err)
	}
}

func TestProcessPaymentEventConfirmed(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ProcessPaymentEvent{SubscriptionRepository: repository}

	sub := entities.Subscription{
		ID:         "sub-pay-1",
		CustomerID: "customer-1",
		PlanID:     "basic",
		Status:     entities.StatusSuspended,
		CreatedAt:  time.Now(),
	}
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	if err := uc.Execute("sub-pay-1", "customer-1", "basic", "confirmed"); err != nil {
		t.Fatalf("unexpected error processing confirmed payment: %v", err)
	}

	found, err := repository.FindByID("sub-pay-1")
	if err != nil {
		t.Fatalf("unexpected error finding subscription: %v", err)
	}

	if found.Status != entities.StatusActive {
		t.Fatalf("expected active status after confirmed payment, got %s", found.Status)
	}
}

func TestProcessPaymentEventRefused(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ProcessPaymentEvent{SubscriptionRepository: repository}

	sub := entities.Subscription{
		ID:         "sub-pay-2",
		CustomerID: "customer-1",
		PlanID:     "basic",
		Status:     entities.StatusActive,
		CreatedAt:  time.Now(),
	}
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	if err := uc.Execute("sub-pay-2", "customer-1", "basic", "refused"); err != nil {
		t.Fatalf("unexpected error processing refused payment: %v", err)
	}

	found, err := repository.FindByID("sub-pay-2")
	if err != nil {
		t.Fatalf("unexpected error finding subscription: %v", err)
	}

	if found.Status != entities.StatusSuspended {
		t.Fatalf("expected suspended status after refused payment, got %s", found.Status)
	}
}

func TestProcessPaymentEventInvalidStatus(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ProcessPaymentEvent{SubscriptionRepository: repository}

	sub := entities.Subscription{
		ID:     "sub-pay-3",
		Status: entities.StatusActive,
	}
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	err := uc.Execute("sub-pay-3", "", "", "unknown")
	if !errors.Is(err, ErrInvalidPaymentStatus) {
		t.Fatalf("expected ErrInvalidPaymentStatus, got %v", err)
	}
}

func TestProcessPaymentEventNotFound(t *testing.T) {
	repository := repositories.NewInMemorySubscription()
	uc := ProcessPaymentEvent{SubscriptionRepository: repository}

	err := uc.Execute("nonexistent", "", "", "confirmed")
	if !errors.Is(err, ports.ErrSubscriptionNotFound) {
		t.Fatalf("expected ports.ErrSubscriptionNotFound, got %v", err)
	}
}
