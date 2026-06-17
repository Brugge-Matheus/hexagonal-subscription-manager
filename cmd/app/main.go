package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"subscription-manager/internal/adapters/input/cli"
	inputwebhook "subscription-manager/internal/adapters/input/webhook"
	"subscription-manager/internal/adapters/output/notification"
	"subscription-manager/internal/adapters/output/repositories"
	"subscription-manager/internal/application/usecases"
)

func main() {
	repository, err := repositories.NewFileSubscription(
		filepath.Join("data", "subscriptions"),
	)
	if err != nil {
		log.Fatal(err)
	}

	notifier := notification.LogNotification{Out: os.Stdout}

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		addr := ":8080"
		if len(os.Args) > 2 {
			addr = os.Args[2]
		}

		handler := inputwebhook.NewHandler(usecases.ProcessPaymentEvent{
			SubscriptionRepository: repository,
			NotificationService:    notifier,
		})

		mux := http.NewServeMux()
		mux.Handle("/webhook/payment", handler)

		log.Printf("webhook server listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}

		return
	}

	appCLI := cli.CLI{
		CreateSubscriptionUseCase: usecases.CreateSubscription{
			SubscriptionRepository: repository,
		},
		ListSubscriptionsUseCase: usecases.ListSubscriptions{
			SubscriptionRepository: repository,
		},
		GetSubscriptionUseCase: usecases.GetSubscription{
			SubscriptionRepository: repository,
		},
		UpdateSubscriptionUseCase: usecases.UpdateSubscription{
			SubscriptionRepository: repository,
		},
		CancelSubscriptionUseCase: usecases.CancelSubscription{
			SubscriptionRepository: repository,
		},
		DeleteSubscriptionUseCase: usecases.DeleteSubscription{
			SubscriptionRepository: repository,
		},
		Out: os.Stdout,
	}

	if err := appCLI.Call(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
