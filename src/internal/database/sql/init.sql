-- Sample data for the mountainbike tutorial (SQLite)
-- Run after ddl.sql

PRAGMA foreign_keys = ON;

BEGIN;

INSERT INTO teams (id, name, country, founded_year, website) VALUES
    (1, 'Specialized Gravity',        'US', 2005, 'https://www.specialized.com'),
    (2, 'Santa Cruz Syndicate',       'US', 2005, 'https://www.santacruzbicycles.com'),
    (3, 'YT Mob',                     'DE', 2016, 'https://www.yt-industries.com'),
    (4, 'Canyon CLLCTV',              'DE', 2021, 'https://www.canyon.com');

INSERT INTO riders (id, team_id, first_name, last_name, nickname, nationality, birth_date, discipline) VALUES
    (1, 2,    'Jackson',  'Goldstone', NULL,       'CA', '2004-01-06', 'downhill'),
    (2, 1,    'Loïc',     'Bruni',     'Bruni',    'FR', '1994-05-15', 'downhill'),
    (3, 3,    'Vali',     'Höll',      NULL,       'AT', '2001-11-27', 'downhill'),
    (4, 4,    'Jack',     'Moir',      NULL,       'AU', '1995-08-30', 'enduro'),
    (5, 1,    'Christopher', 'Blevins', 'Blevs',   'US', '1998-04-11', 'xc'),
    (6, NULL, 'Max',      'Muster',    'Maxi',     'AT', '1990-07-01', 'trail');

INSERT INTO bikes (id, brand, model, model_year, category, wheel_size, travel_front_mm, travel_rear_mm, weight_kg) VALUES
    (1, 'Santa Cruz', 'V10',          2025, 'downhill', 'mullet', 203, 215, 16.50),
    (2, 'Specialized', 'Demo',        2025, 'downhill', 'mullet', 200, 200, 16.80),
    (3, 'YT',          'Tues',        2024, 'downhill', '29',     200, 200, 17.10),
    (4, 'Canyon',      'Strive',      2024, 'enduro',   '29',     170, 160, 15.20),
    (5, 'Specialized', 'Epic World Cup', 2025, 'xc',    '29',     110,  75, 10.10),
    (6, 'Cube',        'Reaction',    2023, 'hardtail', '29',     100,   0, 12.90),
    (7, 'Specialized', 'Stumpjumper', 2024, 'trail',    '29',     150, 145, 14.30);

INSERT INTO rider_bikes (rider_id, bike_id, active) VALUES
    (1, 1, 1),
    (2, 2, 1),
    (2, 7, 1),
    (3, 3, 1),
    (4, 4, 1),
    (5, 5, 1),
    (5, 7, 0),
    (6, 6, 0),
    (6, 7, 1);

INSERT INTO competitions (id, name, discipline, description) VALUES
    (1, 'UCI Downhill World Cup Leogang', 'downhill', 'Classic DH World Cup round in Saalfelden Leogang'),
    (2, 'UCI Downhill World Cup Mont-Sainte-Anne', 'downhill', 'Rough and fast World Cup track in Quebec'),
    (3, 'Enduro World Cup Finale Ligure', 'enduro', 'Multi-stage enduro race at the Italian riviera'),
    (4, 'UCI XCO World Cup Nove Mesto',   'xc',       'Cross-country World Cup in front of huge crowds');

INSERT INTO competition_instances (id, competition_id, location, country, start_date, end_date, status) VALUES
    (1, 1, 'Leogang',           'AT', '2025-06-06', '2025-06-08', 'finished'),
    (2, 1, 'Leogang',           'AT', '2026-06-05', '2026-06-07', 'finished'),
    (3, 2, 'Mont-Sainte-Anne',  'CA', '2025-09-19', '2025-09-21', 'finished'),
    (4, 3, 'Finale Ligure',     'IT', '2026-10-03', '2026-10-04', 'finished'),
    (5, 4, 'Nove Mesto na Morave', 'CZ', '2026-05-15', '2026-05-17', 'finished'),
    (6, 1, 'Leogang',           'AT', '2027-06-04', '2027-06-06', 'planned');

INSERT INTO rider_competition_instances (rider_id, competition_instance_id, position) VALUES
    (1, 1, 2),
    (2, 1, 1),
    (3, 1, 3),
    (1, 2, 1),
    (2, 2, NULL),  -- DNF
    (3, 2, 2),
    (1, 3, 3),
    (2, 3, 1),
    (3, 3, 2),
    (4, 4, 1),
    (6, 4, 57),
    (5, 5, 1),
    (1, 6, NULL),  -- registered, not yet raced
    (3, 6, NULL);

INSERT INTO follows (follower_id, followed_id) VALUES
    (1, 2),
    (2, 1),
    (3, 2),
    (4, 1),
    (6, 1),
    (6, 2),
    (6, 4),
    (6, 5);

COMMIT;
