BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id            TEXT         PRIMARY KEY,
    name          TEXT         NOT NULL,
    email         TEXT         NOT NULL,
    password_hash TEXT         NOT NULL,
    role          TEXT         NOT NULL DEFAULT 'user'
                               CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_active_idx
    ON users (email)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS cities (
    id          TEXT         PRIMARY KEY,
    user_id     TEXT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT cities_user_city_unique UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS cities_user_id_idx ON cities (user_id);

CREATE TABLE IF NOT EXISTS weather_history (
    id           TEXT          PRIMARY KEY,
    user_id      TEXT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    city         TEXT          NOT NULL,
    temp_c       NUMERIC(5,2)  NOT NULL,
    feels_like   NUMERIC(5,2)  NOT NULL,
    humidity     INT           NOT NULL,
    wind_kph     NUMERIC(6,2)  NOT NULL,
    condition    TEXT          NOT NULL,
    requested_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS weather_history_user_city_idx
    ON weather_history (user_id, LOWER(city), requested_at DESC);

COMMIT;
