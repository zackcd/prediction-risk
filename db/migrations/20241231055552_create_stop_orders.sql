-- migrate:up
CREATE SCHEMA IF NOT EXISTS event_contract;

CREATE DOMAIN event_contract.contract_price_cents AS INTEGER CHECK (
    value >= 0
    AND value <= 100
);

CREATE TYPE event_contract.order_status AS ENUM ('ACTIVE', 'TRIGGERED', 'CANCELLED', 'EXPIRED');

CREATE TYPE event_contract.order_type AS ENUM ('STOP');

CREATE TYPE event_contract.order_side AS ENUM ('YES', 'NO');

CREATE TABLE event_contract.order (
    order_id UUID PRIMARY KEY,
    order_type event_contract.order_type NOT NULL,
    ticker VARCHAR NOT NULL,
    side event_contract.order_side NOT NULL,
    status event_contract.order_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE event_contract.stop_order (
    order_id UUID PRIMARY KEY REFERENCES event_contract.order (order_id) ON DELETE CASCADE,
    trigger_price event_contract.contract_price_cents,
    limit_price event_contract.contract_price_cents
);

CREATE UNIQUE INDEX idx_unique_active_stop_order ON event_contract.order (ticker, side)
WHERE
    status = 'ACTIVE'
    AND order_type = 'STOP';

-- migrate:down
DROP SCHEMA IF EXISTS event_contract;
