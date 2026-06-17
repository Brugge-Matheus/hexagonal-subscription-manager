// Package cli
package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/domain/entities"
)

type CLI struct {
	CreateSubscriptionUseCase     ports.SubscribeUseCase
	ListSubscriptionsUseCase      ports.ListSubscriptionsUseCase
	GetSubscriptionUseCase        ports.GetSubscriptionUseCase
	UpdateSubscriptionUseCase     ports.UpdateSubscriptionUseCase
	CancelSubscriptionUseCase     ports.CancelSubscriptionUseCase
	ReactivateSubscriptionUseCase ports.ReactivateSubscriptionUseCase
	DeleteSubscriptionUseCase     ports.DeleteSubscriptionUseCase
	Out                           io.Writer
}

func (c CLI) Call(argv []string) error {
	if len(argv) == 0 {
		c.printUsage()
		return nil
	}

	switch argv[0] {
	case "create-subscription":
		return c.createSubscription(argv[1:])
	case "list-subscriptions":
		return c.listSubscriptions()
	case "show-subscription":
		return c.showSubscription(argv[1:])
	case "update-subscription":
		return c.updateSubscription(argv[1:])
	case "cancel-subscription":
		return c.cancelSubscription(argv[1:])
	case "reactivate-subscription":
		return c.reactivateSubscription(argv[1:])
	case "delete-subscription":
		return c.deleteSubscription(argv[1:])
	default:
		c.printUsage()
		return nil
	}
}

func (c CLI) createSubscription(args []string) error {
	if len(args) < 2 {
		return c.writeLine("missing arguments for create-subscription")
	}

	subscription, err := c.CreateSubscriptionUseCase.Execute(args[0], args[1])
	if err != nil {
		return err
	}

	return c.writeSubscription("Subscription created successfully.", subscription)
}

func (c CLI) listSubscriptions() error {
	subscriptions, err := c.ListSubscriptionsUseCase.Execute()
	if err != nil {
		return err
	}

	if len(subscriptions) == 0 {
		return c.writeLine("No subscriptions found.")
	}

	if err := c.writeLine("Subscriptions:"); err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		line := fmt.Sprintf(
			"ID: %s | Customer: %s | Plan: %s | Status: %s",
			subscription.ID,
			subscription.CustomerID,
			subscription.PlanID,
			subscription.Status,
		)
		if err := c.writeLine(line); err != nil {
			return err
		}
	}

	return nil
}

func (c CLI) showSubscription(args []string) error {
	if len(args) < 1 {
		return c.writeLine("missing subscription id for show-subscription")
	}

	subscription, err := c.GetSubscriptionUseCase.Execute(args[0])
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			return c.writeLine("Subscription not found.")
		}

		return err
	}

	return c.writeSubscription("Subscription found.", subscription)
}

func (c CLI) updateSubscription(args []string) error {
	if len(args) < 4 {
		return c.writeLine("missing arguments for update-subscription")
	}

	subscription, err := c.UpdateSubscriptionUseCase.Execute(args[0], args[1], args[2], args[3])
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			return c.writeLine("Subscription not found.")
		}

		return err
	}

	return c.writeSubscription("Subscription updated successfully.", subscription)
}

func (c CLI) reactivateSubscription(args []string) error {
	if len(args) < 1 {
		return c.writeLine("missing subscription id for reactivate-subscription")
	}

	subscription, err := c.ReactivateSubscriptionUseCase.Execute(args[0])
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrSubscriptionNotFound):
			return c.writeLine("Subscription not found.")
		case errors.Is(err, ports.ErrSubscriptionNotSuspended):
			return c.writeLine("Subscription is not suspended and cannot be reactivated.")
		}
		return err
	}

	return c.writeSubscription("Subscription reactivated successfully.", subscription)
}

func (c CLI) deleteSubscription(args []string) error {
	if len(args) < 1 {
		return c.writeLine("missing subscription id for delete-subscription")
	}

	err := c.DeleteSubscriptionUseCase.Execute(args[0])
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			return c.writeLine("Subscription not found.")
		}

		return err
	}

	return c.writeLine("Subscription deleted successfully.")
}

func (c CLI) cancelSubscription(args []string) error {
	if len(args) < 1 {
		return c.writeLine("missing subscription id for cancel-subscription")
	}

	result, err := c.CancelSubscriptionUseCase.Execute(args[0])
	if err != nil {
		if errors.Is(err, ports.ErrSubscriptionNotFound) {
			return c.writeLine("Subscription not found.")
		}

		return err
	}

	title := "Subscription canceled successfully."
	if result.RefundEligible {
		title += "\nRefund: proportional refund eligible."
	} else {
		title += "\nRefund: no proportional refund."
	}

	return c.writeSubscription(title, result.Subscription)
}

func (c CLI) writeSubscription(title string, subscription entities.Subscription) error {
	lines := []string{
		title,
		"ID: " + subscription.ID,
		"Customer: " + subscription.CustomerID,
		"Plan: " + subscription.PlanID,
		"Status: " + subscription.Status,
	}

	return c.writeLine(strings.Join(lines, "\n"))
}

func (c CLI) printUsage() {
	_ = c.writeLine(`Usage:
  go run ./cmd/app create-subscription CUSTOMER_ID PLAN_ID
  go run ./cmd/app list-subscriptions
  go run ./cmd/app show-subscription SUBSCRIPTION_ID
  go run ./cmd/app update-subscription SUBSCRIPTION_ID CUSTOMER_ID PLAN_ID STATUS
  go run ./cmd/app cancel-subscription SUBSCRIPTION_ID
  go run ./cmd/app reactivate-subscription SUBSCRIPTION_ID
  go run ./cmd/app delete-subscription SUBSCRIPTION_ID`)
}

func (c CLI) writeLine(text string) error {
	_, err := fmt.Fprintln(c.Out, text)
	return err
}
