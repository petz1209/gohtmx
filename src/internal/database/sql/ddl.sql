-- Mountainbike tutorial schema (SQLite)
-- Tables: teams, riders, bikes, rider_bikes, competitions,
--         competition_instances, rider_competition_instances, follows
--
-- Note: SQLite only enforces foreign keys when enabled per connection:
--   PRAGMA foreign_keys = ON;
-- (e.g. in Go via DSN "file:mtb.db?_pragma=foreign_keys(1)" for modernc.org/sqlite
--  or "file:mtb.db?_foreign_keys=on" for mattn/go-sqlite3)

PRAGMA foreign_keys = ON;

BEGIN;

DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS rider_competition_instances;
DROP TABLE IF EXISTS competition_instances;
DROP TABLE IF EXISTS competitions;
DROP TABLE IF EXISTS rider_bikes;
DROP TABLE IF EXISTS bikes;
DROP TABLE IF EXISTS riders;
DROP TABLE IF EXISTS teams;

-- Teams (e.g. "Specialized Factory Racing")
CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL UNIQUE,
    country       TEXT    CHECK (length(country) = 2),     -- ISO 3166-1 alpha-2
    founded_year  INTEGER CHECK (founded_year BETWEEN 1970 AND 2100),
    website       TEXT,
    created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Riders, optionally belonging to a team
CREATE TABLE riders (
    id            INTEGER PRIMARY KEY,
    team_id       INTEGER REFERENCES teams (id) ON DELETE SET NULL,
    first_name    TEXT    NOT NULL,
    last_name     TEXT    NOT NULL,
    nickname      TEXT,
    nationality   TEXT    CHECK (length(nationality) = 2), -- ISO 3166-1 alpha-2
    birth_date    TEXT,                                    -- ISO 8601 'YYYY-MM-DD'
    discipline    TEXT    CHECK (discipline IN ('xc', 'trail', 'enduro', 'downhill', 'freeride', 'dirt')),
    created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_riders_team_id ON riders (team_id);

-- Bikes (catalog of bike models)
CREATE TABLE bikes (
    id               INTEGER PRIMARY KEY,
    brand            TEXT    NOT NULL,
    model            TEXT    NOT NULL,
    model_year       INTEGER CHECK (model_year BETWEEN 1980 AND 2100),
    category         TEXT    NOT NULL CHECK (category IN ('hardtail', 'xc', 'trail', 'enduro', 'downhill', 'dirt')),
    wheel_size       TEXT    CHECK (wheel_size IN ('26', '27.5', '29', 'mullet')),
    travel_front_mm  INTEGER CHECK (travel_front_mm >= 0),
    travel_rear_mm   INTEGER CHECK (travel_rear_mm >= 0),  -- 0 = hardtail
    weight_kg        REAL    CHECK (weight_kg > 0),
    created_at       TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Which rider rides which bike (n:m)
CREATE TABLE rider_bikes (
    id        INTEGER PRIMARY KEY,
    rider_id  INTEGER NOT NULL REFERENCES riders (id) ON DELETE CASCADE,
    bike_id   INTEGER NOT NULL REFERENCES bikes (id) ON DELETE CASCADE,
    active    INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    UNIQUE (rider_id, bike_id)
);

CREATE INDEX idx_rider_bikes_bike_id ON rider_bikes (bike_id);

-- A competition / race series (e.g. "UCI Downhill World Cup Leogang")
CREATE TABLE competitions (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL UNIQUE,
    discipline    TEXT    NOT NULL CHECK (discipline IN ('xc', 'enduro', 'downhill', 'freeride', 'dirt')),
    description   TEXT,
    created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- A concrete edition of a competition (e.g. Leogang 2025)
CREATE TABLE competition_instances (
    id              INTEGER PRIMARY KEY,
    competition_id  INTEGER NOT NULL REFERENCES competitions (id) ON DELETE CASCADE,
    location        TEXT    NOT NULL,
    country         TEXT    CHECK (length(country) = 2),
    start_date      TEXT    NOT NULL,                      -- ISO 8601 'YYYY-MM-DD'
    end_date        TEXT    NOT NULL,
    status          TEXT    NOT NULL DEFAULT 'planned'
                            CHECK (status IN ('planned', 'running', 'finished', 'cancelled')),
    created_at      TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_date >= start_date),
    UNIQUE (competition_id, start_date)
);

CREATE INDEX idx_competition_instances_competition_id ON competition_instances (competition_id);

-- Rider participation / result in a competition instance
CREATE TABLE rider_competition_instances (
    id                       INTEGER PRIMARY KEY,
    rider_id                 INTEGER NOT NULL REFERENCES riders (id) ON DELETE CASCADE,
    competition_instance_id  INTEGER NOT NULL REFERENCES competition_instances (id) ON DELETE CASCADE,
    position                 INTEGER CHECK (position > 0),  -- NULL = DNF / not yet raced
    UNIQUE (rider_id, competition_instance_id)
);

CREATE INDEX idx_rider_competition_instances_instance_id ON rider_competition_instances (competition_instance_id);

-- Rider follows rider (social graph)
CREATE TABLE follows (
    follower_id   INTEGER NOT NULL REFERENCES riders (id) ON DELETE CASCADE,
    followed_id   INTEGER NOT NULL REFERENCES riders (id) ON DELETE CASCADE,
    created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followed_id),
    CHECK (follower_id <> followed_id)
) WITHOUT ROWID;

CREATE INDEX idx_follows_followed_id ON follows (followed_id);

COMMIT;
