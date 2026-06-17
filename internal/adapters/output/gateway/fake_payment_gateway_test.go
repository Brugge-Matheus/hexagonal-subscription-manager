package gateway

import (
	"strings"
	"testing"
)

func TestFakePaymentGatewayCharge(t *testing.T) {
	gw := FakePaymentGateway{}

	txn, err := gw.Charge("sub-123", 29.90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if txn == "" {
		t.Fatal("expected non-empty transaction ID")
	}

	if !strings.Contains(txn, "sub-123") {
		t.Fatalf("expected transaction ID to reference subscription, got %s", txn)
	}
}
