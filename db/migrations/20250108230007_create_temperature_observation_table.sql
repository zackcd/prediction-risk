-- migrate:up
CREATE SCHEMA IF NOT EXISTS weather;

CREATE TABLE weather.nws_station (
    station_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO
    weather.nws_station (station_id, name)
VALUES
    ('KNYC', 'New York City, Central Park');

INSERT INTO
    weather.nws_station (station_id, name)
VALUES
    ('047740', 'San Diego Lindbe, CA');

CREATE TYPE weather.temperature_unit AS ENUM ('CELSIUS', 'FAHRENHEIT');

CREATE TABLE weather.temperature_observation (
    observation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    station_id VARCHAR(50) NOT NULL REFERENCES weather.nws_station (station_id),
    temperature DOUBLE PRECISION NOT NULL,
    temperature_unit weather.temperature_unit NOT NULL,
    timestamp TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_station_timestamp UNIQUE (station_id, timestamp)
);

CREATE INDEX idx_temperature_observations_station_timestamp ON weather.temperature_observation (station_id, timestamp DESC);

-- migrate:down
DROP SCHEMA IF EXISTS weather CASCADE;
