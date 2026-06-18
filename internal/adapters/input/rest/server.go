package rest

import (
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"subscription-manager/internal/application/ports"
)

//go:embed ui.html
var uiHTML []byte

type Server struct {
	CreateSubscriptionUseCase     ports.SubscribeUseCase
	ListSubscriptionsUseCase      ports.ListSubscriptionsUseCase
	GetSubscriptionUseCase        ports.GetSubscriptionUseCase
	UpdateSubscriptionUseCase     ports.UpdateSubscriptionUseCase
	CancelSubscriptionUseCase     ports.CancelSubscriptionUseCase
	ReactivateSubscriptionUseCase ports.ReactivateSubscriptionUseCase
	DeleteSubscriptionUseCase     ports.DeleteSubscriptionUseCase
	ProcessPaymentEventUseCase    ports.ProcessPaymentEventUseCase
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.serveUI)
	mux.HandleFunc("GET /subscriptions", s.list)
	mux.HandleFunc("POST /subscriptions", s.create)
	mux.HandleFunc("GET /subscriptions/{id}", s.get)
	mux.HandleFunc("PUT /subscriptions/{id}", s.update)
	mux.HandleFunc("POST /subscriptions/{id}/cancel", s.cancel)
	mux.HandleFunc("POST /subscriptions/{id}/reactivate", s.reactivate)
	mux.HandleFunc("DELETE /subscriptions/{id}", s.delete)
	mux.HandleFunc("POST /webhook/payment", s.webhookPayment)
	return mux
}

func (s *Server) serveUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(uiHTML)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	subs, err := s.ListSubscriptionsUseCase.Execute()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID string `json:"customer_id"`
		PlanID     string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := s.CreateSubscriptionUseCase.Execute(body.CustomerID, body.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	sub, err := s.GetSubscriptionUseCase.Execute(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID string `json:"customer_id"`
		PlanID     string `json:"plan_id"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := s.UpdateSubscriptionUseCase.Execute(r.PathValue("id"), body.CustomerID, body.PlanID, body.Status)
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (s *Server) cancel(w http.ResponseWriter, r *http.Request) {
	result, err := s.CancelSubscriptionUseCase.Execute(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"subscription":    result.Subscription,
		"refund_eligible": result.RefundEligible,
	})
}

func (s *Server) reactivate(w http.ResponseWriter, r *http.Request) {
	sub, err := s.ReactivateSubscriptionUseCase.Execute(r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrSubscriptionNotFound):
			writeError(w, http.StatusNotFound, "subscription not found")
		case errors.Is(err, ports.ErrSubscriptionNotSuspended):
			writeError(w, http.StatusUnprocessableEntity, "subscription is not suspended")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	if err := s.DeleteSubscriptionUseCase.Execute(r.PathValue("id")); err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) webhookPayment(w http.ResponseWriter, r *http.Request) {
	var event struct {
		SubscriptionID string `json:"subscription_id"`
		CustomerID     string `json:"customer_id"`
		PlanID         string `json:"plan_id"`
		Status         string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	w.WriteHeader(http.StatusAccepted)
	go func() {
		if err := s.ProcessPaymentEventUseCase.Execute(
			event.SubscriptionID, event.CustomerID, event.PlanID, event.Status,
		); err != nil {
			log.Printf("[webhook] error processing payment event: %v", err)
		}
	}()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
