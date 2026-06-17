package ports

import "errors"

var (
	ErrNotFound                 = errors.New("not found")
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrSubscriptionNotSuspended = errors.New("subscription is not suspended")
)
