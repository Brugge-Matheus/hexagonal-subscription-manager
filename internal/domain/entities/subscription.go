// Package entities
package entities

type Subscription struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	PlanID     string `json:"plan_id"`
	Status     string `json:"status"`
}
