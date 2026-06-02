// Package entities
package entities

import "time"

type Subscription struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	PlanID     string     `json:"plan_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	CanceledAt *time.Time `json:"canceled_at"`
}

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusCanceled  = "canceled"
)

func (s Subscription) IsCanceled() bool {
	return s.CanceledAt != nil
}

func (s Subscription) IsActive() bool {
	return s.Status == StatusActive && s.CanceledAt == nil
}

func (s Subscription) IsSuspended() bool {
	return s.Status == StatusSuspended && s.CanceledAt == nil
}

func (s Subscription) CanRefundProportionally(now time.Time) bool {
	if s.CreatedAt.IsZero() {
		return false
	}

	return now.Sub(s.CreatedAt) < 7*24*time.Hour
}

func (s *Subscription) Cancel() {
	if s.CanceledAt == nil {
		now := time.Now()
		s.CanceledAt = &now
		s.Status = StatusCanceled
	}
}

func (s *Subscription) Suspend() {
	if s.IsActive() {
		s.Status = StatusSuspended
	}
}

func (s *Subscription) Reactivate() {
	if s.IsSuspended() {
		s.Status = StatusActive
	}
}
