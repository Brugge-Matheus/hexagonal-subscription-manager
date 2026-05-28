package main

import (
	"log"
	"os"
	"path/filepath"

	"subscription-manager/internal/adapters/input/cli"
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
		DeleteSubscriptionUseCase: usecases.DeleteSubscription{
			SubscriptionRepository: repository,
		},
		Out: os.Stdout,
	}

	if err := appCLI.Call(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
