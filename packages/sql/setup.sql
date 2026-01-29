CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE triptype_position AS ENUM ('prepared', 'on-the-way', 'arrived', 'completed');

CREATE TABLE IF NOT EXISTS drivers(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_dr TEXT NOT NULL,
    surname_dr TEXT NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    car_type TEXT NOT NULL,
    count_trips INTEGER DEAFULT 0,
    count_succesful_trips INTEGER DEAFULT 0;
);

CREATE TABLE IF NOT EXISTS trips(
    id TEXT PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES drivers(id),
    position triptype_position NOT NULL,
    destination TEXT NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP
)

CREATE TABLE IF NOT EXISTS outbox(
    trip_id TEXT NOT NULL REFERENCES trips(id),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    trip_position NOT NULL,
    trip_destination NOT NULL
)