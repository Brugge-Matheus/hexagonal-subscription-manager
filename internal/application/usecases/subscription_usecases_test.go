package usecases

import (
	"testing"
	"time"

	"subscription-manager/internal/adapters/output/repositories"
	"subscription-manager/internal/domain/entities"
)

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
