// Package repositories
package repositories

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"subscription-manager/internal/domain/entities"
)

type FileSubscription struct {
	basePath string
}

func NewFileSubscription(basePath string) (FileSubscription, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return FileSubscription{}, err
	}

	return FileSubscription{basePath: basePath}, nil
}

func (f FileSubscription) Save(subscription entities.Subscription) error {
	data, err := json.MarshalIndent(subscription, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.filePath(subscription.ID), data, 0o644)
}

func (f FileSubscription) FindByID(id string) (entities.Subscription, error) {
	data, err := os.ReadFile(f.filePath(id))
	if err != nil {
		return entities.Subscription{}, err
	}

	var subscription entities.Subscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return entities.Subscription{}, err
	}

	return subscription, nil
}

func (f FileSubscription) All() ([]entities.Subscription, error) {
	entries, err := filepath.Glob(filepath.Join(f.basePath, "*.json"))
	if err != nil {
		return nil, err
	}

	sort.Strings(entries)

	subscriptions := make([]entities.Subscription, 0, len(entries))
	for _, entry := range entries {
		data, err := os.ReadFile(entry)
		if err != nil {
			return nil, err
		}

		var subscription entities.Subscription
		if err := json.Unmarshal(data, &subscription); err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (f FileSubscription) Delete(id string) error {
	err := os.Remove(f.filePath(id))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return err
}

func (f FileSubscription) filePath(id string) string {
	return filepath.Join(f.basePath, id+".json")
}
