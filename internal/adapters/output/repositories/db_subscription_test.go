package repositories

import (
	"os"
	"testing"
	"time"

	"subscription-manager/internal/domain/entities"
)

func TestDBSubscriptionCRUD(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping DB integration test")
	}

	repository, err := NewDBSubscription(dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	const id = "test-db-sub-1"

	t.Cleanup(func() {
		repository.Delete(id)
	})

	sub := entities.Subscription{
		ID:         id,
		CustomerID: "customer-1",
		PlanID:     "basic-plan",
		Status:     entities.StatusActive,
		CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving subscription: %v", err)
	}

	found, err := repository.FindByID(id)
	if err != nil {
		t.Fatalf("unexpected error finding subscription: %v", err)
	}

	if found.CustomerID != sub.CustomerID {
		t.Fatalf("expected customer %s, got %s", sub.CustomerID, found.CustomerID)
	}

	if found.Status != entities.StatusActive {
		t.Fatalf("expected active status, got %s", found.Status)
	}

	if found.CanceledAt != nil {
		t.Fatal("expected nil canceled_at")
	}

	sub.Status = entities.StatusSuspended
	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error updating subscription: %v", err)
	}

	updated, err := repository.FindByID(id)
	if err != nil {
		t.Fatalf("unexpected error finding updated subscription: %v", err)
	}

	if updated.Status != entities.StatusSuspended {
		t.Fatalf("expected suspended after update, got %s", updated.Status)
	}

	all, err := repository.FindAll()
	if err != nil {
		t.Fatalf("unexpected error listing subscriptions: %v", err)
	}

	foundInList := false
	for _, s := range all {
		if s.ID == id {
			foundInList = true
		}
	}

	if !foundInList {
		t.Fatal("subscription not found in FindAll()")
	}

	if err := repository.Delete(id); err != nil {
		t.Fatalf("unexpected error deleting subscription: %v", err)
	}

	if _, err := repository.FindByID(id); err == nil {
		t.Fatal("expected error finding deleted subscription")
	}
}

func TestDBSubscriptionCanceledAt(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping DB integration test")
	}

	repository, err := NewDBSubscription(dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	const id = "test-db-sub-canceled"

	t.Cleanup(func() {
		repository.Delete(id)
	})

	sub := entities.Subscription{
		ID:        id,
		CustomerID: "customer-1",
		PlanID:    "basic-plan",
		Status:    entities.StatusActive,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	sub.Cancel()

	if err := repository.Save(sub); err != nil {
		t.Fatalf("unexpected error saving canceled subscription: %v", err)
	}

	found, err := repository.FindByID(id)
	if err != nil {
		t.Fatalf("unexpected error finding subscription: %v", err)
	}

	if found.CanceledAt == nil {
		t.Fatal("expected non-nil canceled_at")
	}

	if found.Status != entities.StatusCanceled {
		t.Fatalf("expected canceled status, got %s", found.Status)
	}
}
