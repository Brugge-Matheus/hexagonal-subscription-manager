package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"subscription-manager/internal/adapters/input/cli"
	inputwebhook "subscription-manager/internal/adapters/input/webhook"
	"subscription-manager/internal/adapters/output/gateway"
	"subscription-manager/internal/adapters/output/notification"
	"subscription-manager/internal/adapters/output/repositories"
	"subscription-manager/internal/application/ports"
	"subscription-manager/internal/application/usecases"
)

func main() {
	repository := buildRepository()
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
			PaymentGateway:         gateway.FakePaymentGateway{},
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
		ReactivateSubscriptionUseCase: usecases.ReactivateSubscription{
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

func buildRepository() ports.SubscriptionRepository {
	if os.Getenv("REPOSITORY") == "db" {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "postgres://subscription_manager:subscription_manager@localhost:5432/subscription_manager?sslmode=disable"
		}
		repo, err := repositories.NewDBSubscription(dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		log.Println("repository: postgresql")
		return repo
	}

	repo, err := repositories.NewFileSubscription(filepath.Join("data", "subscriptions"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("repository: file")
	return repo
}
