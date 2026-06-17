package repositories

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"subscription-manager/internal/domain/entities"
)

type DBSubscription struct {
	db *sql.DB
}

func NewDBSubscription(dsn string) (*DBSubscription, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DBSubscription{db: db}, nil
}

func (r *DBSubscription) Save(subscription entities.Subscription) error {
	var canceledAt sql.NullTime
	if subscription.CanceledAt != nil {
		canceledAt = sql.NullTime{Time: *subscription.CanceledAt, Valid: true}
	}

	_, err := r.db.Exec(`
		INSERT INTO subscriptions (id, customer_id, plan_id, status, created_at, canceled_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			plan_id     = EXCLUDED.plan_id,
			status      = EXCLUDED.status,
			canceled_at = EXCLUDED.canceled_at
	`,
		subscription.ID,
		subscription.CustomerID,
		subscription.PlanID,
		subscription.Status,
		subscription.CreatedAt,
		canceledAt,
	)
	return err
}

func (r *DBSubscription) FindByID(id string) (entities.Subscription, error) {
	row := r.db.QueryRow(`
		SELECT id, customer_id, plan_id, status, created_at, canceled_at
		FROM subscriptions WHERE id = $1
	`, id)

	var sub entities.Subscription
	var canceledAt sql.NullTime
	err := row.Scan(&sub.ID, &sub.CustomerID, &sub.PlanID, &sub.Status, &sub.CreatedAt, &canceledAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Subscription{}, ErrSubscriptionNotFound
		}
		return entities.Subscription{}, err
	}
	if canceledAt.Valid {
		sub.CanceledAt = &canceledAt.Time
	}
	return sub, nil
}

func (r *DBSubscription) All() ([]entities.Subscription, error) {
	rows, err := r.db.Query(`
		SELECT id, customer_id, plan_id, status, created_at, canceled_at
		FROM subscriptions ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entities.Subscription
	for rows.Next() {
		var sub entities.Subscription
		var canceledAt sql.NullTime
		if err := rows.Scan(&sub.ID, &sub.CustomerID, &sub.PlanID, &sub.Status, &sub.CreatedAt, &canceledAt); err != nil {
			return nil, err
		}
		if canceledAt.Valid {
			sub.CanceledAt = &canceledAt.Time
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

func (r *DBSubscription) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}
