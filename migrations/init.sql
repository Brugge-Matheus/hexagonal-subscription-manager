CREATE TABLE IF NOT EXISTS subscriptions (
    id          VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    plan_id     VARCHAR(255) NOT NULL,
    status      VARCHAR(50)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL,
    canceled_at TIMESTAMPTZ
);
