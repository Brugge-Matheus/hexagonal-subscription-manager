package repositories

import (
	"os"
	"path/filepath"
	"testing"

	"subscription-manager/internal/domain/entities"
)

func TestFileSubscriptionCRUD(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "subscriptions")
	repository, err := NewFileSubscription(basePath)
	if err != nil {
		t.Fatalf("unexpected error creating repository: %v", err)
	}

	subscription := entities.Subscription{
		ID:         "sub-1",
		CustomerID: "customer-1",
		PlanID:     "basic-plan",
		Status:     "active",
	}

	if err := repository.Save(subscription); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	if _, err := os.Stat(filepath.Join(basePath, "sub-1.json")); err != nil {
		t.Fatalf("expected subscription file to exist: %v", err)
	}

	found, err := repository.FindByID("sub-1")
	if err != nil {
		t.Fatalf("unexpected error finding subscription: %v", err)
	}

	if found.CustomerID != "customer-1" {
		t.Fatalf("expected customer-1, got %s", found.CustomerID)
	}

	all, err := repository.All()
	if err != nil {
		t.Fatalf("unexpected error listing subscriptions: %v", err)
	}

	if len(all) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(all))
	}

	if err := repository.Delete("sub-1"); err != nil {
		t.Fatalf("unexpected error deleting subscription: %v", err)
	}

	if _, err := repository.FindByID("sub-1"); err == nil {
		t.Fatal("expected error when finding deleted subscription")
	}
}
