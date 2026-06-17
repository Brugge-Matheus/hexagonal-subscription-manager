package webhook

import (
	"encoding/json"
	"log"
	"net/http"

	"subscription-manager/internal/application/ports"
)

type PaymentEvent struct {
	SubscriptionID string `json:"subscription_id"`
	CustomerID     string `json:"customer_id"`
	PlanID         string `json:"plan_id"`
	Status         string `json:"status"`
}

type Handler struct {
	processPaymentEvent ports.ProcessPaymentEventUseCase
}

func NewHandler(uc ports.ProcessPaymentEventUseCase) *Handler {
	return &Handler{processPaymentEvent: uc}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event PaymentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go func() {
		err := h.processPaymentEvent.Execute(
			event.SubscriptionID,
			event.CustomerID,
			event.PlanID,
			event.Status,
		)
		if err != nil {
			log.Printf("[webhook] error processing event for subscription %s: %v", event.SubscriptionID, err)
		}
	}()
}
