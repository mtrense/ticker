
BEGIN;

CREATE TABLE IF NOT EXISTS event_streams (
    global_sequence SERIAL PRIMARY KEY,
    aggregate VARCHAR(50) ARRAY NOT NULL,
--     aggregate_sequence INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    revision SMALLINT NOT NULL,
    payload JSONB
--     UNIQUE (aggregate, aggregate_sequence)
);

CREATE INDEX IF NOT EXISTS event_streams_aggregate_idx ON event_streams USING GIN(aggregate);

COMMIT;
