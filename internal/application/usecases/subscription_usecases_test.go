package usecases

import (
	"path/filepath"
	"testing"

	"subscription-manager/internal/adapters/output/repositories"
)

func TestSubscriptionUseCasesCRUD(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "subscriptions")
	repository, err := repositories.NewFileSubscription(basePath)
	if err != nil {
		t.Fatalf("unexpected error creating repository: %v", err)
	}

	createUseCase := CreateSubscription{SubscriptionRepository: repository}
	listUseCase := ListSubscriptions{SubscriptionRepository: repository}
	getUseCase := GetSubscription{SubscriptionRepository: repository}
	updateUseCase := UpdateSubscription{SubscriptionRepository: repository}
	deleteUseCase := DeleteSubscription{SubscriptionRepository: repository}

	created, err := createUseCase.Execute("customer-1", "basic-plan")
	if err != nil {
		t.Fatalf("unexpected error creating subscription: %v", err)
	}

	if created.Status != "active" {
		t.Fatalf("expected active status, got %s", created.Status)
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

	if err := deleteUseCase.Execute(created.ID); err != nil {
		t.Fatalf("unexpected error deleting subscription: %v", err)
	}

	if _, err := getUseCase.Execute(created.ID); err == nil {
		t.Fatal("expected error when reading deleted subscription")
	}
}
