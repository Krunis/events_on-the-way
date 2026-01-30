CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE triptype_position AS ENUM ('prepared', 'on-the-way', 'arrived', 'completed');

CREATE TABLE IF NOT EXISTS drivers(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_dr TEXT NOT NULL,
    surname_dr TEXT NOT NULL,
    registered_at TIMESTAMPTZ NOT NULL,
    car_type TEXT NOT NULL,
    count_trips INTEGER DEFAULT 0,
    count_succesful_trips INTEGER DEAFULT 0;
);

CREATE TABLE IF NOT EXISTS trips(
    id TEXT PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES drivers(id),
    position triptype_position NOT NULL,
    destination TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
)

CREATE TABLE IF NOT EXISTS outbox(
    event_id UUID PRIMARY KEY
    trip_id TEXT NOT NULL REFERENCES trips(id),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    trip_position triptype_position NOT NULL,
    trip_destination TEXT NOT NULL,
    event_status TEXT NOT NULL DEFAULT "NEW",
    created_at TIMESTAMPTZ NOT NULL
)