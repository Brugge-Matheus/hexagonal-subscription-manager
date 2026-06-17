package repositories

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"subscription-manager/internal/domain/entities"
)

type FileSubscription struct {
	mu       sync.RWMutex
	basePath string
}

func NewFileSubscription(basePath string) (*FileSubscription, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, err
	}
	return &FileSubscription{basePath: basePath}, nil
}

func (f *FileSubscription) Save(subscription entities.Subscription) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, err := json.MarshalIndent(subscription, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.filePath(subscription.ID), data, 0o644)
}

func (f *FileSubscription) FindByID(id string) (entities.Subscription, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	data, err := os.ReadFile(f.filePath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return entities.Subscription{}, ErrSubscriptionNotFound
		}
		return entities.Subscription{}, err
	}
	var subscription entities.Subscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return entities.Subscription{}, err
	}
	return subscription, nil
}

func (f *FileSubscription) FindAll() ([]entities.Subscription, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
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

func (f *FileSubscription) Delete(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	err := os.Remove(f.filePath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrSubscriptionNotFound
		}
		return err
	}
	return nil
}

func (f *FileSubscription) filePath(id string) string {
	return filepath.Join(f.basePath, id+".json")
}
