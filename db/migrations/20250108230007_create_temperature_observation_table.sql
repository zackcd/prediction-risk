-- migrate:up
CREATE SCHEMA IF NOT EXISTS weather;

CREATE TABLE weather.station (
    station_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    nws_station_id VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO
    weather.station (nws_station_id, name)
VALUES
    ('KNYC', 'New York City, Central Park');

CREATE TYPE weather.temperature_unit AS ENUM ('CELSIUS', 'FAHRENHEIT');

CREATE TABLE weather.temperature_observations (
    observation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    station_id UUID NOT NULL,
    temperature DOUBLE PRECISION NOT NULL,
    unit weather.temperature_unit NOT NULL,
    timestamp TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_station_timestamp UNIQUE (station_id, timestamp),
    FOREIGN KEY (station_id) REFERENCES weather.station (station_id)
);

CREATE INDEX idx_temperature_observations_station_timestamp ON weather.temperature_observations (station_id, timestamp DESC);

-- migrate:down
DROP SCHEMA IF EXISTS weather CASCADE;
