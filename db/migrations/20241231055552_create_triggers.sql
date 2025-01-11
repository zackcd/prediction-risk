-- migrate:up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS event_contract;

-- Enumerated types
CREATE TYPE event_contract.trigger_type AS ENUM ('STOP');

CREATE TYPE event_contract.order_side AS ENUM ('BUY', 'SELL');

CREATE TYPE event_contract.trigger_status AS ENUM ('ACTIVE', 'TRIGGERED', 'CANCELLED', 'EXPIRED');

CREATE TYPE event_contract.price_direction AS ENUM ('ABOVE', 'BELOW');

CREATE TYPE event_contract.contract_side AS ENUM ('YES', 'NO');

CREATE DOMAIN event_contract.contract_price_cents AS INTEGER CHECK (
    value >= 0
    AND value <= 100
);

-- Triggers table to store all trigger types including stop triggers
CREATE TABLE event_contract.trigger (
    trigger_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    trigger_type event_contract.trigger_type NOT NULL,
    contract_ticker VARCHAR(255) NOT NULL,
    contract_side event_contract.contract_side NOT NULL,
    status event_contract.trigger_status NOT NULL DEFAULT 'ACTIVE',
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW ()
);

CREATE TABLE event_contract.price_trigger_condition (
    trigger_id UUID PRIMARY KEY REFERENCES event_contract.trigger (trigger_id) ON DELETE CASCADE,
    contract_ticker VARCHAR(255) NOT NULL,
    contract_side event_contract.contract_side NOT NULL,
    threshold_price NOT NULL event_contract.contract_price_cents,
    direction NOT NULL price_direction,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW ()
);

CREATE TABLE event_contract.trigger_action (
    action_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    trigger_id UUID NOT NULL REFERENCES event_contract.trigger (trigger_id) ON DELETE CASCADE,
    contract_ticker VARCHAR(255) NOT NULL,
    contract_side event_contract.contract_side NOT NULL,
    order_side event_contract.order_side NOT NULL,
    order_size NUMERIC(20, 0), -- Nullable for "full position" in case of sells
    limit_price event_contract.contract_price_cents, -- Nullable for market orders
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    -- Ensure order_size is not null for buy orders
    CONSTRAINT valid_buy_order_size CHECK (
        (order_side = 'SELL')
        OR (
            order_side = 'BUY'
            AND order_size IS NOT NULL
        )
    )
);

-- Indexes
CREATE INDEX idx_triggers_contract_status ON triggers (contract_ticker, contract_side, status);

CREATE INDEX idx_actions_trigger ON trigger_actions (trigger_id);

-- Ensure stop triggers have exactly one action
CREATE UNIQUE INDEX idx_stop_trigger_single_action ON trigger_actions (trigger_id)
WHERE
    EXISTS (
        SELECT
            1
        FROM
            triggers t
        WHERE
            t.trigger_id = trigger_actions.trigger_id
            AND t.trigger_type = 'STOP'
    );

-- migrate:down
DROP SCHEMA IF EXISTS event_contract;

DROP EXTENSION IF EXISTS "uuid-ossp";
